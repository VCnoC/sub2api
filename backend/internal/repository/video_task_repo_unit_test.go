//go:build unit

// Package repository 测试视频任务与余额扣费、失败退款的事务边界。
package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestUsageBillingApplyRollsBackBalanceWhenVideoTaskInsertFails(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	mock.ExpectQuery(`(?s)INSERT INTO usage_billing_dedup.*RETURNING id`).
		WithArgs("request-123", int64(22), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(`(?s)SELECT request_fingerprint.*FROM usage_billing_dedup_archive`).
		WithArgs("request-123", int64(22)).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(conditionalBalanceDeductSQL).
		WithArgs(2.5, int64(11)).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(97.5))
	mock.ExpectExec(`(?s)INSERT INTO video_tasks`).
		WithArgs("video-123", "request-123", int64(11), int64(22), int64(33), int64(44), 2.5).
		WillReturnError(errors.New("video task insert failed"))
	mock.ExpectRollback()

	repo := &usageBillingRepository{db: db}
	_, err = repo.Apply(context.Background(), &service.UsageBillingCommand{
		RequestID:   "request-123",
		UserID:      11,
		APIKeyID:    22,
		AccountID:   33,
		BalanceCost: 2.5,
		VideoTask: &service.VideoTaskBillingSnapshot{
			UpstreamTaskID: "video-123",
			GroupID:        44,
		},
	})

	require.EqualError(t, err, "video task insert failed")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestVideoTaskFailAndRefundIsIdempotent(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	repo := &videoTaskRepository{db: db}

	mock.ExpectBegin()
	mock.ExpectQuery(`(?s)SELECT status, refunded_at, refund_amount, user_id, api_key_id, billing_request_id.*FOR UPDATE`).
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"status", "refunded_at", "refund_amount", "user_id", "api_key_id", "billing_request_id"}).
			AddRow(service.VideoTaskStatusPending, nil, 2.5, int64(11), int64(22), "request-123"))
	mock.ExpectQuery(`(?s)UPDATE users.*RETURNING balance`).
		WithArgs(2.5, int64(11)).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(100.0))
	mock.ExpectExec(`(?s)UPDATE video_tasks.*SET status = 'failed'`).
		WithArgs(int64(7), "generation failed").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`(?s)UPDATE usage_logs.*WHERE request_id = \$1 AND api_key_id = \$2`).
		WithArgs("request-123", int64(22)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	first, err := repo.FailAndRefund(context.Background(), 7, "generation failed")
	require.NoError(t, err)
	require.True(t, first.Applied)
	require.NotNil(t, first.NewBalance)
	require.InDelta(t, 100, *first.NewBalance, 1e-12)

	mock.ExpectBegin()
	mock.ExpectQuery(`(?s)SELECT status, refunded_at, refund_amount, user_id, api_key_id, billing_request_id.*FOR UPDATE`).
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"status", "refunded_at", "refund_amount", "user_id", "api_key_id", "billing_request_id"}).
			AddRow(service.VideoTaskStatusFailed, time.Now(), 2.5, int64(11), int64(22), "request-123"))
	mock.ExpectRollback()

	second, err := repo.FailAndRefund(context.Background(), 7, "generation failed")
	require.NoError(t, err)
	require.False(t, second.Applied)
	require.NoError(t, mock.ExpectationsWereMet())
}
