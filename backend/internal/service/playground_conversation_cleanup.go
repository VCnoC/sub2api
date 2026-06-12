// Package service 提供对话广场会话过期清理定时任务。
//
// PlaygroundConversationCleanupService 每隔 cleanup_interval_minutes 分钟执行一次批量物理删除，
// 将 last_activity_at 超过 conversation_retention_days 天的会话清除。
// retention_days ≤ 0 时禁用清理；通过 startOnce 防止重复启动。
package service

import (
	"context"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
)

const (
	// playgroundCleanupWorkerName TimingWheel 任务名称（全局唯一）
	playgroundCleanupWorkerName = "playground_conversation_cleanup"

	// playgroundCleanupMaxBatches 单次 runOnce 允许的最大批次数，防止死循环
	playgroundCleanupMaxBatches = 100

	// playgroundCleanupRunTimeout 单次清理任务的最大执行时长
	playgroundCleanupRunTimeout = 5 * time.Minute
)

// PlaygroundConversationCleanupService 负责定期清理过期的对话广场会话。
type PlaygroundConversationCleanupService struct {
	repo        PlaygroundConversationRepository
	timingWheel *TimingWheelService
	cfg         *config.Config

	startOnce sync.Once
	stopOnce  sync.Once

	workerCtx    context.Context
	workerCancel context.CancelFunc
}

// NewPlaygroundConversationCleanupService 创建 PlaygroundConversationCleanupService 实例。
func NewPlaygroundConversationCleanupService(
	repo PlaygroundConversationRepository,
	timingWheel *TimingWheelService,
	cfg *config.Config,
) *PlaygroundConversationCleanupService {
	ctx, cancel := context.WithCancel(context.Background())
	return &PlaygroundConversationCleanupService{
		repo:         repo,
		timingWheel:  timingWheel,
		cfg:          cfg,
		workerCtx:    ctx,
		workerCancel: cancel,
	}
}

// retentionDays 返回会话保留天数配置值（0 表示禁用清理）。
func (s *PlaygroundConversationCleanupService) retentionDays() int {
	if s == nil || s.cfg == nil {
		return 3
	}
	return s.cfg.Playground.Cleanup.ConversationRetentionDays
}

// cleanupInterval 返回清理任务调度间隔。
func (s *PlaygroundConversationCleanupService) cleanupInterval() time.Duration {
	if s == nil || s.cfg == nil {
		return 60 * time.Minute
	}
	if s.cfg.Playground.Cleanup.CleanupIntervalMinutes > 0 {
		return time.Duration(s.cfg.Playground.Cleanup.CleanupIntervalMinutes) * time.Minute
	}
	return 60 * time.Minute
}

// batchSize 返回每批删除的条数上限。
func (s *PlaygroundConversationCleanupService) batchSize() int {
	if s == nil || s.cfg == nil {
		return 500
	}
	if s.cfg.Playground.Cleanup.CleanupBatchSize > 0 {
		return s.cfg.Playground.Cleanup.CleanupBatchSize
	}
	return 500
}

// Start 注册到 TimingWheel 并启动定时清理任务。
// 若 retention_days ≤ 0，跳过启动并记录禁用日志。
// 多次调用仅第一次有效（startOnce 保护）。
func (s *PlaygroundConversationCleanupService) Start() {
	if s == nil {
		return
	}
	if s.retentionDays() <= 0 {
		logger.LegacyPrintf("service.playground_cleanup",
			"[PlaygroundCleanup] not started (disabled: conversation_retention_days=%d)",
			s.retentionDays())
		return
	}
	if s.repo == nil || s.timingWheel == nil {
		logger.LegacyPrintf("service.playground_cleanup",
			"[PlaygroundCleanup] not started (missing deps)")
		return
	}

	interval := s.cleanupInterval()
	s.startOnce.Do(func() {
		s.timingWheel.ScheduleRecurring(playgroundCleanupWorkerName, interval, s.runOnce)
		logger.LegacyPrintf("service.playground_cleanup",
			"[PlaygroundCleanup] started (interval=%s retention_days=%d batch_size=%d)",
			interval, s.retentionDays(), s.batchSize())
	})
}

// Stop 停止定时清理任务。
func (s *PlaygroundConversationCleanupService) Stop() {
	if s == nil {
		return
	}
	s.stopOnce.Do(func() {
		if s.workerCancel != nil {
			s.workerCancel()
		}
		if s.timingWheel != nil {
			s.timingWheel.Cancel(playgroundCleanupWorkerName)
		}
		logger.LegacyPrintf("service.playground_cleanup", "[PlaygroundCleanup] stopped")
	})
}

// runOnce 执行一次过期会话清理，循环批量删除直到无更多数据或达到批次上限。
func (s *PlaygroundConversationCleanupService) runOnce() {
	if s == nil || s.repo == nil {
		return
	}

	// 构建带超时的上下文；若 workerCtx 已取消则跳过
	parent := context.Background()
	if s.workerCtx != nil {
		parent = s.workerCtx
	}
	ctx, cancel := context.WithTimeout(parent, playgroundCleanupRunTimeout)
	defer cancel()

	// 过期时间基准：当前时刻减去保留天数
	before := time.Now().Add(-time.Duration(s.retentionDays()) * 24 * time.Hour)
	batchSize := s.batchSize()
	deletedTotal := 0

	for batchNum := 1; batchNum <= playgroundCleanupMaxBatches; batchNum++ {
		// 检查外部取消信号
		if ctx.Err() != nil {
			logger.LegacyPrintf("service.playground_cleanup",
				"[PlaygroundCleanup] run interrupted: batch=%d deleted_total=%d err=%v",
				batchNum, deletedTotal, ctx.Err())
			return
		}

		deleted, err := s.repo.DeleteExpired(ctx, before, batchSize)
		if err != nil {
			logger.LegacyPrintf("service.playground_cleanup",
				"[PlaygroundCleanup] delete_expired failed: batch=%d deleted_total=%d err=%v",
				batchNum, deletedTotal, err)
			return
		}

		deletedTotal += deleted

		// 本批次未删满，说明已清理完毕
		if deleted < batchSize {
			break
		}

		// 达到最大批次上限，退出并记录警告
		if batchNum == playgroundCleanupMaxBatches {
			logger.LegacyPrintf("service.playground_cleanup",
				"[PlaygroundCleanup] max_batches reached: batch=%d deleted_total=%d (will retry next cycle)",
				batchNum, deletedTotal)
			break
		}
	}

	if deletedTotal > 0 {
		logger.LegacyPrintf("service.playground_cleanup",
			"[PlaygroundCleanup] run done: deleted_total=%d before=%s",
			deletedTotal, before.UTC().Format("2006-01-02T15:04:05Z"))
	}
}
