// Package service 运行视频任务终态轮询，并在明确失败时幂等退回余额。
package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	"github.com/tidwall/gjson"
	"go.uber.org/zap"
)

const (
	videoTaskPollBatchSize = 20
	videoTaskLease         = 2 * time.Minute
	videoTaskIdleDelay     = 2 * time.Second
	videoTaskQueryTimeout  = 30 * time.Second
	videoTaskBodyLimit     = int64(2 << 20)
)

type VideoTaskWorkerRuntime struct {
	repo         VideoTaskRepository
	accountRepo  AccountRepository
	httpUpstream HTTPUpstream
	billingCache *BillingCacheService
	cfg          *config.Config

	mu     sync.Mutex
	cancel context.CancelFunc
	done   chan struct{}
}

func ProvideVideoTaskWorkerRuntime(
	repo VideoTaskRepository,
	accountRepo AccountRepository,
	httpUpstream HTTPUpstream,
	billingCache *BillingCacheService,
	cfg *config.Config,
) *VideoTaskWorkerRuntime {
	runtime := &VideoTaskWorkerRuntime{
		repo: repo, accountRepo: accountRepo, httpUpstream: httpUpstream,
		billingCache: billingCache, cfg: cfg,
	}
	runtime.Start()
	return runtime
}

func (r *VideoTaskWorkerRuntime) Start() {
	if r == nil || r.repo == nil || r.accountRepo == nil || r.httpUpstream == nil || r.cfg == nil || r.cfg.RunMode == config.RunModeSimple {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.cancel != nil {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	r.cancel = cancel
	r.done = done
	go func() {
		defer close(done)
		r.run(ctx)
	}()
}

func (r *VideoTaskWorkerRuntime) Stop() {
	if r == nil {
		return
	}
	r.mu.Lock()
	cancel, done := r.cancel, r.done
	r.cancel, r.done = nil, nil
	r.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	if done != nil {
		<-done
	}
}

func (r *VideoTaskWorkerRuntime) run(ctx context.Context) {
	for ctx.Err() == nil {
		count, err := r.RunOnce(ctx)
		if err != nil {
			logger.L().Warn("video_task.poll_batch_failed", zap.Error(err))
			sleepOrDone(ctx, videoTaskIdleDelay)
			continue
		}
		if count == 0 {
			sleepOrDone(ctx, videoTaskIdleDelay)
		}
	}
}

func (r *VideoTaskWorkerRuntime) RunOnce(ctx context.Context) (int, error) {
	if r == nil || r.repo == nil {
		return 0, nil
	}
	tasks, err := r.repo.ClaimDue(ctx, videoTaskPollBatchSize, videoTaskLease)
	if err != nil {
		return 0, err
	}
	for i := range tasks {
		if ctx.Err() != nil {
			return i, ctx.Err()
		}
		r.processTask(ctx, &tasks[i])
	}
	return len(tasks), nil
}

func (r *VideoTaskWorkerRuntime) processTask(ctx context.Context, task *VideoTask) {
	if task == nil {
		return
	}
	retry := func(message string) {
		if err := r.repo.ScheduleRetry(ctx, task.ID, videoTaskRetryDelay(task.PollAttempts), truncateVideoTaskError(message)); err != nil && ctx.Err() == nil {
			logger.L().Warn("video_task.schedule_retry_failed", zap.Int64("task_id", task.ID), zap.Error(err))
		}
	}

	account, err := r.accountRepo.GetByID(ctx, task.AccountID)
	if err != nil || account == nil {
		retry("load account: " + errorText(err))
		return
	}
	if !account.IsVideo() || account.GetVideoAPIKey() == "" || account.GetVideoBaseURL() == "" {
		retry("video account credentials are unavailable")
		return
	}
	baseURL, err := (&AccountTestService{cfg: r.cfg}).validateUpstreamBaseURL(account.GetVideoBaseURL())
	if err != nil {
		retry("invalid video base URL: " + err.Error())
		return
	}

	queryURL := buildOpenAIEndpointURL(baseURL, "/v1/videos/"+url.PathEscape(task.UpstreamTaskID))
	queryCtx, cancel := context.WithTimeout(ctx, videoTaskQueryTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(queryCtx, http.MethodGet, queryURL, nil)
	if err != nil {
		retry("build video query: " + err.Error())
		return
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+account.GetVideoAPIKey())

	resp, err := r.httpUpstream.Do(req, upstreamModelsProxyURL(account), account.ID, account.Concurrency)
	if err != nil {
		retry("query video upstream: " + err.Error())
		return
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(io.LimitReader(resp.Body, videoTaskBodyLimit+1))
	if err != nil {
		retry("read video upstream response: " + err.Error())
		return
	}
	if int64(len(body)) > videoTaskBodyLimit {
		retry("video upstream response is too large")
		return
	}

	status := normalizeVideoTaskUpstreamStatus(firstVideoTaskStatus(body))
	switch status {
	case VideoTaskStatusCompleted:
		if err := r.repo.MarkCompleted(ctx, task.ID); err != nil && ctx.Err() == nil {
			logger.L().Warn("video_task.mark_completed_failed", zap.Int64("task_id", task.ID), zap.Error(err))
		}
	case VideoTaskStatusFailed:
		message := firstVideoTaskError(body)
		result, err := r.repo.FailAndRefund(ctx, task.ID, truncateVideoTaskError(message))
		if err != nil {
			logger.L().Warn("video_task.refund_failed", zap.Int64("task_id", task.ID), zap.Error(err))
			return
		}
		if result != nil && result.Applied && r.billingCache != nil {
			if err := r.billingCache.InvalidateUserBalance(ctx, result.UserID); err != nil {
				logger.L().Warn("video_task.refund_cache_invalidate_failed", zap.Int64("user_id", result.UserID), zap.Error(err))
			}
		}
	default:
		retry(fmt.Sprintf("video status not terminal: http=%d status=%s", resp.StatusCode, status))
	}
}

func firstVideoTaskStatus(body []byte) string {
	for _, path := range []string{"status", "data.status", "video.status", "result.status"} {
		if value := strings.TrimSpace(gjson.GetBytes(body, path).String()); value != "" {
			return value
		}
	}
	return ""
}

func normalizeVideoTaskUpstreamStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "completed", "done", "succeeded", "success":
		return VideoTaskStatusCompleted
	case "failed", "error", "expired", "cancelled", "canceled":
		return VideoTaskStatusFailed
	default:
		return VideoTaskStatusPending
	}
}

func firstVideoTaskError(body []byte) string {
	for _, path := range []string{"error.message", "error.error", "error", "message"} {
		if value := strings.TrimSpace(gjson.GetBytes(body, path).String()); value != "" {
			return value
		}
	}
	return "upstream video generation failed"
}

func videoTaskRetryDelay(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	delay := 10 * time.Second
	for i := 1; i < attempt && delay < 2*time.Minute; i++ {
		delay *= 2
	}
	if delay > 2*time.Minute {
		return 2 * time.Minute
	}
	return delay
}

func truncateVideoTaskError(message string) string {
	message = strings.TrimSpace(message)
	if len(message) > 1000 {
		return message[:1000]
	}
	return message
}

func errorText(err error) string {
	if err == nil {
		return "not found"
	}
	return err.Error()
}
