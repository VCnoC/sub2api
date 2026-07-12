package service

import (
	"context"
	"errors"
	"mime/multipart"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

type ticketRepositoryStub struct {
	addReplyInput *AddTicketReplyRecordInput
	updateInput   *UpdateTicketRecordInput
	ticket        *Ticket
	admins        []TicketUserSummary
	adminsErr     error
}

func (*ticketRepositoryStub) Create(context.Context, CreateTicketRecordInput) (*Ticket, error) {
	return nil, nil
}
func (*ticketRepositoryStub) List(context.Context, int64, bool, pagination.PaginationParams, TicketListFilters) ([]Ticket, *pagination.PaginationResult, error) {
	return nil, nil, nil
}
func (r *ticketRepositoryStub) Get(context.Context, int64, int64, bool) (*Ticket, error) {
	return r.ticket, nil
}
func (r *ticketRepositoryStub) AddReply(_ context.Context, input AddTicketReplyRecordInput) (*Ticket, error) {
	r.addReplyInput = &input
	return r.ticket, nil
}
func (r *ticketRepositoryStub) Update(_ context.Context, input UpdateTicketRecordInput) (*Ticket, error) {
	r.updateInput = &input
	return r.ticket, nil
}
func (*ticketRepositoryStub) MarkRead(context.Context, int64, int64, bool, time.Time) error {
	return nil
}
func (*ticketRepositoryStub) UnreadCount(context.Context, int64, bool) (int64, error) { return 0, nil }
func (*ticketRepositoryStub) GetAttachment(context.Context, int64, int64, bool) (*TicketAttachment, error) {
	return nil, ErrTicketAttachmentGone
}
func (*ticketRepositoryStub) DeleteAttachment(context.Context, int64, int64, string, time.Time) (*TicketAttachment, error) {
	return nil, nil
}
func (*ticketRepositoryStub) ListAttachmentsDue(context.Context, time.Time, int) ([]TicketAttachment, error) {
	return nil, nil
}
func (*ticketRepositoryStub) MarkAttachmentDeleted(context.Context, int64, string, time.Time) error {
	return nil
}
func (r *ticketRepositoryStub) ListActiveAdmins(context.Context) ([]TicketUserSummary, error) {
	return r.admins, r.adminsErr
}
func (*ticketRepositoryStub) GetUserSummary(context.Context, int64) (*TicketUserSummary, error) {
	return nil, nil
}

func TestTicketServiceReplyBuildsPublicAndInternalTransitions(t *testing.T) {
	tests := []struct {
		name        string
		input       ReplyTicketInput
		wantKind    string
		wantIsAdmin bool
		wantBody    string
	}{
		{
			name:     "user public reply",
			input:    ReplyTicketInput{TicketID: 10, ActorID: 20, Body: "  user reply  "},
			wantKind: TicketMessageKindPublic, wantBody: "user reply",
		},
		{
			name:     "admin internal note",
			input:    ReplyTicketInput{TicketID: 10, ActorID: 30, IsAdmin: true, Internal: true, Body: "  private note  "},
			wantKind: TicketMessageKindInternal, wantIsAdmin: true, wantBody: "private note",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := &ticketRepositoryStub{ticket: &Ticket{ID: 10}}
			svc := NewTicketService(repo, &TicketFileStore{root: t.TempDir()}, nil)

			_, err := svc.Reply(context.Background(), test.input)
			require.NoError(t, err)
			require.NotNil(t, repo.addReplyInput)
			require.Equal(t, test.wantKind, repo.addReplyInput.Kind)
			require.Equal(t, test.wantIsAdmin, repo.addReplyInput.IsAdmin)
			require.Equal(t, test.wantBody, repo.addReplyInput.Body)
		})
	}
}

func TestTicketServiceRejectsInvalidInternalAttachmentAndEmptyReply(t *testing.T) {
	repo := &ticketRepositoryStub{ticket: &Ticket{ID: 10}}
	svc := NewTicketService(repo, &TicketFileStore{root: t.TempDir()}, nil)

	_, err := svc.Reply(context.Background(), ReplyTicketInput{TicketID: 10, ActorID: 30, IsAdmin: true, Internal: true, Body: "note", Files: ticketOversizeHeaders(1)})
	require.ErrorIs(t, err, ErrTicketInvalidInput)

	_, err = svc.Reply(context.Background(), ReplyTicketInput{TicketID: 10, ActorID: 20})
	require.ErrorIs(t, err, ErrTicketInvalidInput)
	require.Nil(t, repo.addReplyInput)
}

func TestTicketServiceUpdatePassesCloseRetentionAndAssignee(t *testing.T) {
	repo := &ticketRepositoryStub{ticket: &Ticket{ID: 10}}
	svc := NewTicketService(repo, &TicketFileStore{root: t.TempDir()}, nil)
	priority := TicketPriorityUrgent
	assigneeID := int64(30)
	closed := true

	before := time.Now().Add(ticketAttachmentRetention)
	_, err := svc.Update(context.Background(), UpdateTicketInput{
		TicketID: 10, ActorID: 30, Priority: &priority,
		SetAssignee: true, AssigneeID: &assigneeID, Closed: &closed,
	})
	after := time.Now().Add(ticketAttachmentRetention)

	require.NoError(t, err)
	require.NotNil(t, repo.updateInput)
	require.True(t, repo.updateInput.SetClosed)
	require.True(t, repo.updateInput.Closed)
	require.Equal(t, &assigneeID, repo.updateInput.AssigneeID)
	require.Equal(t, &priority, repo.updateInput.Priority)
	require.False(t, repo.updateInput.DeleteAfter.Before(before))
	require.False(t, repo.updateInput.DeleteAfter.After(after))
}

func TestTicketServiceAdminNotificationRecipients(t *testing.T) {
	adminA := TicketUserSummary{ID: 1, Email: "admin-a@example.com", Username: "Admin A"}
	adminB := TicketUserSummary{ID: 2, Email: "admin-b@example.com", Username: "Admin B"}

	tests := []struct {
		name       string
		assigneeID *int64
		wantSent   int64
	}{
		{name: "assigned active admin only", assigneeID: ticketInt64(2), wantSent: 1},
		{name: "missing assignee falls back to all active admins", assigneeID: ticketInt64(99), wantSent: 2},
		{name: "unassigned notifies all active admins", wantSent: 2},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			settings := newNotificationEmailMemorySettingRepo()
			smtpServer := startNotificationEmailTestSMTPServer(t)
			require.NoError(t, settings.SetMultiple(ctx, smtpServer.settings()))
			notifications := NewNotificationEmailService(settings, NewEmailService(settings, nil))
			ticket := &Ticket{
				ID: 10, UserID: 20, Subject: "API request failed", Status: TicketStatusPendingAdmin,
				AssigneeID: test.assigneeID, Messages: []TicketMessage{{ID: 100}},
			}
			repo := &ticketRepositoryStub{ticket: ticket, admins: []TicketUserSummary{adminA, adminB}}
			svc := NewTicketService(repo, &TicketFileStore{root: t.TempDir()}, notifications)

			_, err := svc.Reply(ctx, ReplyTicketInput{TicketID: 10, ActorID: 20, Body: "still failing"})

			require.NoError(t, err)
			require.Equal(t, test.wantSent, smtpServer.messageCount())
		})
	}
}

func TestTicketServiceNotificationFailureDoesNotRollbackReply(t *testing.T) {
	repo := &ticketRepositoryStub{
		ticket: &Ticket{ID: 10, UserID: 20, Subject: "Billing question", Status: TicketStatusPendingAdmin, Messages: []TicketMessage{{ID: 101}}},
		admins: []TicketUserSummary{{ID: 1, Email: "admin@example.com"}},
	}
	settings := newNotificationEmailMemorySettingRepo()
	notifications := NewNotificationEmailService(settings, nil)
	svc := NewTicketService(repo, &TicketFileStore{root: t.TempDir()}, notifications)

	item, err := svc.Reply(context.Background(), ReplyTicketInput{TicketID: 10, ActorID: 20, Body: "please help"})

	require.NoError(t, err)
	require.Same(t, repo.ticket, item)
	require.NotNil(t, repo.addReplyInput)
}

func TestTicketServiceNotificationDeduplicatesSameMessage(t *testing.T) {
	ctx := context.Background()
	settings := newNotificationEmailMemorySettingRepo()
	smtpServer := startNotificationEmailTestSMTPServer(t)
	require.NoError(t, settings.SetMultiple(ctx, smtpServer.settings()))
	notifications := NewNotificationEmailService(settings, NewEmailService(settings, nil))
	repo := &ticketRepositoryStub{
		ticket: &Ticket{ID: 10, UserID: 20, Subject: "Model error", Status: TicketStatusPendingAdmin, Messages: []TicketMessage{{ID: 102}}},
		admins: []TicketUserSummary{{ID: 1, Email: "admin@example.com"}},
	}
	svc := NewTicketService(repo, &TicketFileStore{root: t.TempDir()}, notifications)

	for range 2 {
		_, err := svc.Reply(ctx, ReplyTicketInput{TicketID: 10, ActorID: 20, Body: "same committed message"})
		require.NoError(t, err)
	}

	require.Equal(t, int64(1), smtpServer.messageCount())
}

func TestTicketServiceAdminLookupFailureDoesNotRollbackReply(t *testing.T) {
	repo := &ticketRepositoryStub{
		ticket:    &Ticket{ID: 10, UserID: 20, Subject: "API error", Status: TicketStatusPendingAdmin},
		adminsErr: errors.New("database unavailable"),
	}
	notifications := NewNotificationEmailService(newNotificationEmailMemorySettingRepo(), nil)
	svc := NewTicketService(repo, &TicketFileStore{root: t.TempDir()}, notifications)

	item, err := svc.Reply(context.Background(), ReplyTicketInput{TicketID: 10, ActorID: 20, Body: "retry details"})

	require.NoError(t, err)
	require.Same(t, repo.ticket, item)
}

func ticketInt64(value int64) *int64 { return &value }

func ticketOversizeHeaders(count int) []*multipart.FileHeader {
	return make([]*multipart.FileHeader, count)
}
