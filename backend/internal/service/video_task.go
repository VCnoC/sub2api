// Package service 定义视频任务轮询与失败退款的领域契约。
package service

import (
	"context"
	"time"
)

const (
	VideoTaskStatusPending   = "pending"
	VideoTaskStatusCompleted = "completed"
	VideoTaskStatusFailed    = "failed"
)

type VideoTaskBillingSnapshot struct {
	UpstreamTaskID string
	GroupID        int64
}

type VideoTask struct {
	ID               int64
	UpstreamTaskID   string
	BillingRequestID string
	UserID           int64
	APIKeyID         int64
	AccountID        int64
	GroupID          int64
	RefundAmount     float64
	Status           string
	PollAttempts     int
}

type VideoTaskRefundResult struct {
	Applied    bool
	UserID     int64
	NewBalance *float64
}

type VideoTaskRepository interface {
	ClaimDue(ctx context.Context, limit int, lease time.Duration) ([]VideoTask, error)
	MarkCompleted(ctx context.Context, taskID int64) error
	ScheduleRetry(ctx context.Context, taskID int64, delay time.Duration, lastError string) error
	FailAndRefund(ctx context.Context, taskID int64, lastError string) (*VideoTaskRefundResult, error)
}
