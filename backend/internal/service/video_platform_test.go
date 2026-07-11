// Package service 测试独立视频平台的透传、计费与异步终态处理。
package service

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestParseVideoPlatformRequestSupportsSecondsAndSize(t *testing.T) {
	info := ParseGrokMediaRequest("application/json", []byte(`{"seconds":"6","size":"480p"}`))

	require.True(t, info.DurationProvided)
	require.Equal(t, 6, info.DurationSeconds)
	require.Equal(t, VideoBillingResolution480P, info.Resolution)
}

func TestForwardVideoPlatformPreservesCPARequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	body := []byte(`{"prompt":"waves","size":"1280x720"}`)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/videos", strings.NewReader(string(body)))
	c.Request.Header.Set("Content-Type", "application/json")

	upstream := &httpUpstreamRecorder{resp: &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(`{"id":"video-123","status":"queued"}`)),
	}}
	svc := &OpenAIGatewayService{httpUpstream: upstream, cfg: &config.Config{}}
	account := &Account{
		ID:          7,
		Platform:    PlatformVideo,
		Type:        AccountTypeAPIKey,
		Concurrency: 3,
		Credentials: map[string]any{
			"base_url": "https://video.example.com",
			"api_key":  "video-key",
		},
	}

	result, err := svc.ForwardGrokMedia(context.Background(), c, account, GrokMediaEndpointVideosCreate, "", body, "application/json")
	require.NoError(t, err)
	require.Equal(t, "https://video.example.com/v1/videos", upstream.lastReq.URL.String())
	require.Equal(t, "Bearer video-key", upstream.lastReq.Header.Get("Authorization"))
	require.JSONEq(t, string(body), string(upstream.lastBody))
	require.Equal(t, "video-123", result.ResponseID)
	require.Equal(t, DefaultVideoPlatformModel, result.Model)
	require.Equal(t, VideoPlatformDefaultDurationSeconds, result.VideoDurationSeconds)
	require.Equal(t, VideoBillingResolution720P, result.VideoResolution)
}

func TestCalculateOpenAIVideoCostHonorsBillingMode(t *testing.T) {
	price := 0.25
	svc := &OpenAIGatewayService{billingService: NewBillingService(&config.Config{}, nil)}
	result := &OpenAIForwardResult{
		VideoCount:           1,
		VideoResolution:      VideoBillingResolution720P,
		VideoDurationSeconds: 8,
	}

	for _, tc := range []struct {
		mode string
		want float64
	}{
		{VideoBillingModePerSecond, 2},
		{VideoBillingModePerRequest, 0.25},
	} {
		t.Run(tc.mode, func(t *testing.T) {
			apiKey := &APIKey{Group: &Group{VideoBillingMode: tc.mode, VideoPrice720P: &price}}
			cost := svc.calculateOpenAIVideoCost(context.Background(), DefaultVideoPlatformModel, apiKey, result, 1)
			require.InDelta(t, tc.want, cost.ActualCost, 1e-12)
		})
	}
}

type videoTaskRepositoryStub struct {
	VideoTaskRepository
	tasks          []VideoTask
	completedCalls int
	retryCalls     int
	refundCalls    int
}

func (s *videoTaskRepositoryStub) ClaimDue(context.Context, int, time.Duration) ([]VideoTask, error) {
	return s.tasks, nil
}

func (s *videoTaskRepositoryStub) MarkCompleted(context.Context, int64) error {
	s.completedCalls++
	return nil
}

func (s *videoTaskRepositoryStub) ScheduleRetry(context.Context, int64, time.Duration, string) error {
	s.retryCalls++
	return nil
}

func (s *videoTaskRepositoryStub) FailAndRefund(context.Context, int64, string) (*VideoTaskRefundResult, error) {
	s.refundCalls++
	return &VideoTaskRefundResult{Applied: true, UserID: 11}, nil
}

type videoTaskAccountRepositoryStub struct {
	AccountRepository
	account *Account
}

func (s *videoTaskAccountRepositoryStub) GetByID(context.Context, int64) (*Account, error) {
	return s.account, nil
}

type videoTaskHTTPUpstreamStub struct {
	HTTPUpstream
	status int
	body   string
	err    error
	req    *http.Request
}

func (s *videoTaskHTTPUpstreamStub) Do(req *http.Request, _ string, _ int64, _ int) (*http.Response, error) {
	s.req = req
	if s.err != nil {
		return nil, s.err
	}
	return &http.Response{
		StatusCode: s.status,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(s.body)),
	}, nil
}

func TestVideoTaskWorkerClassifiesUpstreamResults(t *testing.T) {
	account := &Account{
		ID:          9,
		Platform:    PlatformVideo,
		Type:        AccountTypeAPIKey,
		Concurrency: 2,
		Credentials: map[string]any{"base_url": "https://video.example.com", "api_key": "video-key"},
	}

	for _, tc := range []struct {
		name          string
		status        int
		body          string
		err           error
		wantCompleted int
		wantRetry     int
		wantRefund    int
	}{
		{name: "completed", status: http.StatusOK, body: `{"status":"done"}`, wantCompleted: 1},
		{name: "explicit failure on gateway error", status: http.StatusBadGateway, body: `{"status":"failed","error":{"message":"generation failed"}}`, wantRefund: 1},
		{name: "unknown status", status: http.StatusOK, body: `{"status":"processing"}`, wantRetry: 1},
		{name: "network error", err: errors.New("network unavailable"), wantRetry: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &videoTaskRepositoryStub{tasks: []VideoTask{{ID: 1, UpstreamTaskID: "video-123", AccountID: account.ID, PollAttempts: 1}}}
			upstream := &videoTaskHTTPUpstreamStub{status: tc.status, body: tc.body, err: tc.err}
			worker := &VideoTaskWorkerRuntime{
				repo:         repo,
				accountRepo:  &videoTaskAccountRepositoryStub{account: account},
				httpUpstream: upstream,
				cfg:          &config.Config{},
			}

			count, err := worker.RunOnce(context.Background())
			require.NoError(t, err)
			require.Equal(t, 1, count)
			require.Equal(t, tc.wantCompleted, repo.completedCalls)
			require.Equal(t, tc.wantRetry, repo.retryCalls)
			require.Equal(t, tc.wantRefund, repo.refundCalls)
			if upstream.req != nil {
				require.Equal(t, "https://video.example.com/v1/videos/video-123", upstream.req.URL.String())
				require.Equal(t, "Bearer video-key", upstream.req.Header.Get("Authorization"))
			}
		})
	}
}
