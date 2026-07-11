// Package repository 实现视频任务的数据库租约和幂等退款事务。
package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type videoTaskRepository struct {
	db *sql.DB
}

func NewVideoTaskRepository(_ *dbent.Client, db *sql.DB) service.VideoTaskRepository {
	return &videoTaskRepository{db: db}
}

func (r *videoTaskRepository) ClaimDue(ctx context.Context, limit int, lease time.Duration) ([]service.VideoTask, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("video task repository db is nil")
	}
	if limit <= 0 {
		limit = 20
	}
	if lease <= 0 {
		lease = 2 * time.Minute
	}
	rows, err := r.db.QueryContext(ctx, `
		WITH due AS (
			SELECT id
			FROM video_tasks
			WHERE status = 'pending'
			  AND next_poll_at <= NOW()
			  AND (locked_until IS NULL OR locked_until < NOW())
			ORDER BY next_poll_at, id
			LIMIT $1
			FOR UPDATE SKIP LOCKED
		)
		UPDATE video_tasks AS task
		SET locked_until = NOW() + ($2 * INTERVAL '1 second'),
			poll_attempts = task.poll_attempts + 1,
			updated_at = NOW()
		FROM due
		WHERE task.id = due.id
		RETURNING task.id, task.upstream_task_id, task.billing_request_id,
			task.user_id, task.api_key_id, task.account_id, task.group_id,
			task.refund_amount, task.status, task.poll_attempts
	`, limit, int64(lease/time.Second))
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	tasks := make([]service.VideoTask, 0, limit)
	for rows.Next() {
		var task service.VideoTask
		if err := rows.Scan(
			&task.ID, &task.UpstreamTaskID, &task.BillingRequestID,
			&task.UserID, &task.APIKeyID, &task.AccountID, &task.GroupID,
			&task.RefundAmount, &task.Status, &task.PollAttempts,
		); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, rows.Err()
}

func (r *videoTaskRepository) MarkCompleted(ctx context.Context, taskID int64) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE video_tasks
		SET status = 'completed', terminal_at = NOW(), locked_until = NULL,
			last_error = NULL, updated_at = NOW()
		WHERE id = $1 AND status = 'pending'
	`, taskID)
	return err
}

func (r *videoTaskRepository) ScheduleRetry(ctx context.Context, taskID int64, delay time.Duration, lastError string) error {
	if delay <= 0 {
		delay = 15 * time.Second
	}
	_, err := r.db.ExecContext(ctx, `
		UPDATE video_tasks
		SET next_poll_at = NOW() + ($2 * INTERVAL '1 second'),
			locked_until = NULL, last_error = NULLIF($3, ''), updated_at = NOW()
		WHERE id = $1 AND status = 'pending'
	`, taskID, int64(delay/time.Second), lastError)
	return err
}

func (r *videoTaskRepository) FailAndRefund(ctx context.Context, taskID int64, lastError string) (_ *service.VideoTaskRefundResult, err error) {
	if r == nil || r.db == nil {
		return nil, errors.New("video task repository db is nil")
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()

	var status string
	var refundedAt sql.NullTime
	var refundAmount float64
	var userID, apiKeyID int64
	var billingRequestID string
	err = tx.QueryRowContext(ctx, `
		SELECT status, refunded_at, refund_amount, user_id, api_key_id, billing_request_id
		FROM video_tasks
		WHERE id = $1
		FOR UPDATE
	`, taskID).Scan(&status, &refundedAt, &refundAmount, &userID, &apiKeyID, &billingRequestID)
	if err != nil {
		return nil, err
	}
	if status == service.VideoTaskStatusCompleted || refundedAt.Valid {
		return &service.VideoTaskRefundResult{Applied: false, UserID: userID}, nil
	}

	result := &service.VideoTaskRefundResult{Applied: true, UserID: userID}
	if refundAmount > 0 {
		var newBalance float64
		err = tx.QueryRowContext(ctx, `
			UPDATE users
			SET balance = balance + $1, updated_at = NOW()
			WHERE id = $2 AND deleted_at IS NULL
			RETURNING balance
		`, refundAmount, userID).Scan(&newBalance)
		if err != nil {
			return nil, err
		}
		result.NewBalance = &newBalance
	}

	if _, err = tx.ExecContext(ctx, `
		UPDATE video_tasks
		SET status = 'failed', terminal_at = NOW(), refunded_at = NOW(),
			locked_until = NULL, last_error = NULLIF($2, ''), updated_at = NOW()
		WHERE id = $1
	`, taskID, lastError); err != nil {
		return nil, err
	}
	if _, err = tx.ExecContext(ctx, `
		UPDATE usage_logs
		SET image_output_cost = 0, input_cost = 0, output_cost = 0,
			cache_creation_cost = 0, cache_read_cost = 0,
			total_cost = 0, actual_cost = 0
		WHERE request_id = $1 AND api_key_id = $2
	`, billingRequestID, apiKeyID); err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}
	tx = nil
	return result, nil
}
