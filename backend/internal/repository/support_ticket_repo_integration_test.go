//go:build integration

package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestSupportTicketRepositoryLifecycleIsolationAndUnread(t *testing.T) {
	ctx := context.Background()
	userID := createTicketTestUser(t, "user")
	otherUserID := createTicketTestUser(t, "user")
	adminID := createTicketTestUser(t, "admin")
	repo := NewSupportTicketRepository(integrationDB)
	now := time.Now().UTC()

	ticket, err := repo.Create(ctx, service.CreateTicketRecordInput{
		UserID: userID, Subject: "model failed", Category: service.TicketCategoryModel,
		Body: "initial report", Now: now, DayStart: now.Add(-time.Hour),
		Attachments: []service.TicketAttachment{{OriginalName: "report.txt", StorageKey: fmt.Sprintf("ticket-%d.txt", now.UnixNano()), MediaType: "text/plain", SizeBytes: 12}},
	})
	require.NoError(t, err)
	require.Equal(t, service.TicketStatusPendingAdmin, ticket.Status)
	require.Equal(t, service.TicketPriorityNormal, ticket.Priority)
	require.Len(t, ticket.Messages, 1)
	require.Len(t, ticket.Messages[0].Attachments, 1)
	attachmentID := ticket.Messages[0].Attachments[0].ID

	_, err = repo.Get(ctx, ticket.ID, otherUserID, false)
	require.ErrorIs(t, err, service.ErrTicketNotFound)
	_, err = repo.GetAttachment(ctx, attachmentID, otherUserID, false)
	require.ErrorIs(t, err, service.ErrTicketAttachmentGone)
	_, err = repo.GetAttachment(ctx, attachmentID, adminID, true)
	require.NoError(t, err)

	count, err := repo.UnreadCount(ctx, adminID, true)
	require.NoError(t, err)
	require.EqualValues(t, 1, count)
	require.NoError(t, repo.MarkRead(ctx, ticket.ID, adminID, true, now.Add(time.Second)))

	ticket, err = repo.AddReply(ctx, service.AddTicketReplyRecordInput{TicketID: ticket.ID, ActorID: adminID, IsAdmin: true, Kind: service.TicketMessageKindPublic, Body: "public answer", Now: now.Add(2 * time.Second)})
	require.NoError(t, err)
	require.Equal(t, service.TicketStatusPendingUser, ticket.Status)

	ticket, err = repo.AddReply(ctx, service.AddTicketReplyRecordInput{TicketID: ticket.ID, ActorID: adminID, IsAdmin: true, Kind: service.TicketMessageKindInternal, Body: "private note", Now: now.Add(3 * time.Second)})
	require.NoError(t, err)
	require.Equal(t, service.TicketStatusPendingUser, ticket.Status)

	deleteAfter := now.Add(30 * 24 * time.Hour)
	priority := service.TicketPriorityUrgent
	ticket, err = repo.Update(ctx, service.UpdateTicketRecordInput{
		TicketID: ticket.ID, ActorID: adminID, Priority: &priority, SetAssignee: true, AssigneeID: &adminID,
		SetClosed: true, Closed: true, Now: now.Add(4 * time.Second), DeleteAfter: deleteAfter,
	})
	require.NoError(t, err)
	require.Equal(t, service.TicketStatusClosed, ticket.Status)
	require.Equal(t, service.TicketPriorityUrgent, ticket.Priority)
	require.Equal(t, &adminID, ticket.AssigneeID)

	var storedDeleteAfter sql.NullTime
	require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT delete_after FROM support_ticket_attachments WHERE id=$1`, attachmentID).Scan(&storedDeleteAfter))
	require.True(t, storedDeleteAfter.Valid)
	require.WithinDuration(t, deleteAfter, storedDeleteAfter.Time, time.Second)

	ticket, err = repo.AddReply(ctx, service.AddTicketReplyRecordInput{TicketID: ticket.ID, ActorID: userID, Kind: service.TicketMessageKindPublic, Body: "still broken", Now: now.Add(5 * time.Second)})
	require.NoError(t, err)
	require.Equal(t, service.TicketStatusPendingAdmin, ticket.Status)
	require.NoError(t, integrationDB.QueryRowContext(ctx, `SELECT delete_after FROM support_ticket_attachments WHERE id=$1`, attachmentID).Scan(&storedDeleteAfter))
	require.False(t, storedDeleteAfter.Valid)

	count, err = repo.UnreadCount(ctx, adminID, true)
	require.NoError(t, err)
	require.EqualValues(t, 1, count)

	userView, err := repo.Get(ctx, ticket.ID, userID, false)
	require.NoError(t, err)
	for _, message := range userView.Messages {
		require.NotEqual(t, service.TicketMessageKindInternal, message.Kind)
	}
	require.True(t, containsTicketEvent(userView.Messages, "reopened"))

	adminView, err := repo.Get(ctx, ticket.ID, adminID, true)
	require.NoError(t, err)
	require.True(t, containsTicketKind(adminView.Messages, service.TicketMessageKindInternal))
}

func TestSupportTicketRepositoryConcurrentOpenLimit(t *testing.T) {
	ctx := context.Background()
	userID := createTicketTestUser(t, "user")
	repo := NewSupportTicketRepository(integrationDB)
	now := time.Now().UTC()
	for index := 0; index < 4; index++ {
		_, err := repo.Create(ctx, service.CreateTicketRecordInput{UserID: userID, Subject: fmt.Sprintf("seed-%d", index), Category: service.TicketCategoryOther, Body: "body", Now: now.Add(time.Duration(index) * time.Millisecond), DayStart: now.Add(-time.Hour)})
		require.NoError(t, err)
	}

	start := make(chan struct{})
	errorsCh := make(chan error, 2)
	var group sync.WaitGroup
	for index := 0; index < 2; index++ {
		group.Add(1)
		go func(index int) {
			defer group.Done()
			<-start
			_, err := repo.Create(ctx, service.CreateTicketRecordInput{UserID: userID, Subject: fmt.Sprintf("race-%d", index), Category: service.TicketCategoryOther, Body: "body", Now: now.Add(time.Second), DayStart: now.Add(-time.Hour)})
			errorsCh <- err
		}(index)
	}
	close(start)
	group.Wait()
	close(errorsCh)

	var successes, limited int
	for err := range errorsCh {
		if err == nil {
			successes++
		} else if errors.Is(err, service.ErrTicketOpenLimit) {
			limited++
		}
	}
	require.Equal(t, 1, successes)
	require.Equal(t, 1, limited)
}

func TestSupportTicketRepositoryDailyLimit(t *testing.T) {
	ctx := context.Background()
	userID := createTicketTestUser(t, "user")
	repo := NewSupportTicketRepository(integrationDB)
	now := time.Now().UTC()
	for index := 0; index < 10; index++ {
		_, err := integrationDB.ExecContext(ctx, `INSERT INTO support_tickets (user_id,subject,status,closed_at,created_at,updated_at,last_message_at) VALUES ($1,$2,'closed',$3,$3,$3,$3)`, userID, fmt.Sprintf("closed-%d", index), now.Add(time.Duration(index)*time.Millisecond))
		require.NoError(t, err)
	}

	_, err := repo.Create(ctx, service.CreateTicketRecordInput{UserID: userID, Subject: "eleventh", Category: service.TicketCategoryOther, Body: "body", Now: now.Add(time.Second), DayStart: now.Add(-time.Hour)})
	require.ErrorIs(t, err, service.ErrTicketDailyLimit)
}

func createTicketTestUser(t *testing.T, role string) int64 {
	t.Helper()
	ctx := context.Background()
	email := fmt.Sprintf("support-ticket-%s-%d@example.com", role, time.Now().UnixNano())
	var id int64
	require.NoError(t, integrationDB.QueryRowContext(ctx, `INSERT INTO users (email,password_hash,role,status,balance,concurrency) VALUES ($1,'test-hash',$2,'active',0,5) RETURNING id`, email, role).Scan(&id))
	t.Cleanup(func() {
		_, _ = integrationDB.ExecContext(ctx, `DELETE FROM support_tickets WHERE user_id=$1`, id)
		_, _ = integrationDB.ExecContext(ctx, `DELETE FROM users WHERE id=$1`, id)
	})
	return id
}

func containsTicketKind(messages []service.TicketMessage, kind string) bool {
	for _, message := range messages {
		if message.Kind == kind {
			return true
		}
	}
	return false
}

func containsTicketEvent(messages []service.TicketMessage, event string) bool {
	for _, message := range messages {
		var metadata struct {
			Event string `json:"event"`
		}
		if json.Unmarshal(message.Metadata, &metadata) == nil && metadata.Event == event {
			return true
		}
	}
	return false
}
