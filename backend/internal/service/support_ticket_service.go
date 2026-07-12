// Package service 编排站内工单生命周期、附件和通知。
package service

import (
	"context"
	"fmt"
	"log/slog"
	"mime/multipart"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

const (
	ticketMaxBodyLength       = 50_000
	ticketAttachmentRetention = 30 * 24 * time.Hour
	defaultTicketCleanupBatch = 100
)

type TicketService struct {
	repo          TicketRepository
	files         *TicketFileStore
	notifications *NotificationEmailService
}

func NewTicketService(repo TicketRepository, files *TicketFileStore, notifications *NotificationEmailService) *TicketService {
	return &TicketService{repo: repo, files: files, notifications: notifications}
}

type CreateTicketInput struct {
	UserID   int64
	Subject  string
	Category string
	Body     string
	Files    []*multipart.FileHeader
}

type ReplyTicketInput struct {
	TicketID int64
	ActorID  int64
	IsAdmin  bool
	Internal bool
	Body     string
	Files    []*multipart.FileHeader
}

type UpdateTicketInput struct {
	TicketID    int64
	ActorID     int64
	Priority    *string
	SetAssignee bool
	AssigneeID  *int64
	Closed      *bool
}

func (s *TicketService) Create(ctx context.Context, input CreateTicketInput) (*Ticket, error) {
	input.Subject = strings.TrimSpace(input.Subject)
	input.Body = strings.TrimSpace(input.Body)
	input.Category = strings.TrimSpace(input.Category)
	if input.UserID <= 0 || input.Subject == "" || len(input.Subject) > 200 || input.Body == "" || len(input.Body) > ticketMaxBodyLength || !validTicketCategory(input.Category) {
		return nil, ErrTicketInvalidInput
	}
	attachments, cleanup, err := s.files.SaveUploads(input.Files)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	localNow := now.In(time.Local)
	dayStart := time.Date(localNow.Year(), localNow.Month(), localNow.Day(), 0, 0, 0, 0, localNow.Location())
	ticket, err := s.repo.Create(ctx, CreateTicketRecordInput{
		UserID: input.UserID, Subject: input.Subject, Category: input.Category, Body: input.Body,
		Attachments: attachments, Now: now, DayStart: dayStart,
	})
	if err != nil {
		cleanup()
		return nil, err
	}
	s.notifyAdmins(ctx, ticket, "ticket.created")
	return ticket, nil
}

func (s *TicketService) List(ctx context.Context, viewerID int64, isAdmin bool, params pagination.PaginationParams, filters TicketListFilters) ([]Ticket, *pagination.PaginationResult, error) {
	if filters.Status != "" && !validTicketStatus(filters.Status) {
		return nil, nil, ErrTicketInvalidStatus
	}
	if filters.Category != "" && !validTicketCategory(filters.Category) {
		return nil, nil, ErrTicketInvalidInput
	}
	if filters.Priority != "" && !validTicketPriority(filters.Priority) {
		return nil, nil, ErrTicketInvalidInput
	}
	filters.Search = strings.TrimSpace(filters.Search)
	if len(filters.Search) > 200 {
		filters.Search = filters.Search[:200]
	}
	return s.repo.List(ctx, viewerID, isAdmin, params, filters)
}

func (s *TicketService) Get(ctx context.Context, ticketID, viewerID int64, isAdmin bool) (*Ticket, error) {
	ticket, err := s.repo.Get(ctx, ticketID, viewerID, isAdmin)
	if err != nil {
		return nil, err
	}
	if err := s.repo.MarkRead(ctx, ticketID, viewerID, isAdmin, time.Now()); err != nil {
		return nil, err
	}
	ticket.Unread = false
	return ticket, nil
}

func (s *TicketService) Reply(ctx context.Context, input ReplyTicketInput) (*Ticket, error) {
	input.Body = strings.TrimSpace(input.Body)
	if input.TicketID <= 0 || input.ActorID <= 0 || len(input.Body) > ticketMaxBodyLength || (input.Body == "" && len(input.Files) == 0) {
		return nil, ErrTicketInvalidInput
	}
	if input.Internal && (!input.IsAdmin || len(input.Files) > 0) {
		return nil, ErrTicketInvalidInput
	}
	attachments, cleanup, err := s.files.SaveUploads(input.Files)
	if err != nil {
		return nil, err
	}
	kind := TicketMessageKindPublic
	if input.Internal {
		kind = TicketMessageKindInternal
	}
	ticket, err := s.repo.AddReply(ctx, AddTicketReplyRecordInput{
		TicketID: input.TicketID, ActorID: input.ActorID, IsAdmin: input.IsAdmin,
		Kind: kind, Body: input.Body, Attachments: attachments, Now: time.Now(),
	})
	if err != nil {
		cleanup()
		return nil, err
	}
	if !input.Internal {
		if input.IsAdmin {
			s.notifyUser(ctx, ticket, "ticket.admin_reply")
		} else {
			s.notifyAdmins(ctx, ticket, "ticket.user_reply")
		}
	}
	return ticket, nil
}

func (s *TicketService) Update(ctx context.Context, input UpdateTicketInput) (*Ticket, error) {
	if input.TicketID <= 0 || input.ActorID <= 0 || (input.Priority == nil && !input.SetAssignee && input.Closed == nil) {
		return nil, ErrTicketInvalidInput
	}
	if input.Priority != nil {
		value := strings.TrimSpace(*input.Priority)
		if !validTicketPriority(value) {
			return nil, ErrTicketInvalidInput
		}
		input.Priority = &value
	}
	now := time.Now()
	ticket, err := s.repo.Update(ctx, UpdateTicketRecordInput{
		TicketID: input.TicketID, ActorID: input.ActorID, Priority: input.Priority,
		SetAssignee: input.SetAssignee, AssigneeID: input.AssigneeID,
		SetClosed: input.Closed != nil, Closed: input.Closed != nil && *input.Closed,
		Now: now, DeleteAfter: now.Add(ticketAttachmentRetention),
	})
	if err != nil {
		return nil, err
	}
	if input.Closed != nil && *input.Closed {
		s.notifyUser(ctx, ticket, "ticket.closed")
	}
	if input.SetAssignee && input.AssigneeID != nil {
		s.notifyAssignee(ctx, ticket, *input.AssigneeID)
	}
	return ticket, nil
}

func (s *TicketService) UnreadCount(ctx context.Context, viewerID int64, isAdmin bool) (int64, error) {
	return s.repo.UnreadCount(ctx, viewerID, isAdmin)
}

func (s *TicketService) OpenAttachment(ctx context.Context, attachmentID, viewerID int64, isAdmin bool) (*TicketAttachment, *os.File, error) {
	item, err := s.repo.GetAttachment(ctx, attachmentID, viewerID, isAdmin)
	if err != nil {
		return nil, nil, err
	}
	file, err := s.files.Open(item.StorageKey)
	if err != nil {
		return nil, nil, ErrTicketAttachmentGone
	}
	return item, file, nil
}

func (s *TicketService) DeleteAttachment(ctx context.Context, attachmentID, actorID int64, reason string) error {
	reason = strings.TrimSpace(reason)
	if reason == "" || len(reason) > 255 {
		return ErrTicketInvalidInput
	}
	item, err := s.repo.GetAttachment(ctx, attachmentID, actorID, true)
	if err != nil {
		return err
	}
	if err := s.files.Delete(item.StorageKey); err != nil {
		return err
	}
	_, err = s.repo.DeleteAttachment(ctx, attachmentID, actorID, reason, time.Now())
	return err
}

func (s *TicketService) CleanupAttachments(ctx context.Context, now time.Time) (int, error) {
	items, err := s.repo.ListAttachmentsDue(ctx, now, defaultTicketCleanupBatch)
	if err != nil {
		return 0, err
	}
	cleaned := 0
	for i := range items {
		item := items[i]
		if err := s.files.Delete(item.StorageKey); err != nil {
			slog.Error("ticket attachment cleanup failed", "attachment_id", item.ID, "error", err)
			continue
		}
		if err := s.repo.MarkAttachmentDeleted(ctx, item.ID, "retention_expired", now); err != nil {
			slog.Error("ticket attachment cleanup mark failed", "attachment_id", item.ID, "error", err)
			continue
		}
		cleaned++
	}
	return cleaned, nil
}

func (s *TicketService) notifyAdmins(ctx context.Context, ticket *Ticket, event string) {
	if s.notifications == nil || ticket == nil {
		return
	}
	admins, err := s.repo.ListActiveAdmins(ctx)
	if err != nil {
		slog.Error("list ticket notification admins failed", "ticket_id", ticket.ID, "error", err)
		return
	}
	if ticket.AssigneeID != nil {
		filtered := admins[:0]
		for _, admin := range admins {
			if admin.ID == *ticket.AssigneeID {
				filtered = append(filtered, admin)
			}
		}
		if len(filtered) > 0 {
			admins = filtered
		}
	}
	for _, admin := range admins {
		s.sendNotification(ctx, event, ticket, admin)
	}
}

func (s *TicketService) notifyUser(ctx context.Context, ticket *Ticket, event string) {
	if ticket == nil {
		return
	}
	s.sendNotification(ctx, event, ticket, ticket.User)
}

func (s *TicketService) notifyAssignee(ctx context.Context, ticket *Ticket, assigneeID int64) {
	admins, err := s.repo.ListActiveAdmins(ctx)
	if err != nil {
		return
	}
	for _, admin := range admins {
		if admin.ID == assigneeID {
			s.sendNotification(ctx, "ticket.assigned", ticket, admin)
			return
		}
	}
}

func (s *TicketService) sendNotification(ctx context.Context, event string, ticket *Ticket, recipient TicketUserSummary) {
	if s.notifications == nil || strings.TrimSpace(recipient.Email) == "" {
		return
	}
	name := recipient.Username
	if name == "" {
		name = recipient.Email
	}
	messageID := int64(0)
	if count := len(ticket.Messages); count > 0 {
		messageID = ticket.Messages[count-1].ID
	}
	path := fmt.Sprintf("/tickets/%d", ticket.ID)
	if recipient.ID != ticket.UserID {
		path = fmt.Sprintf("/admin/tickets/%d", ticket.ID)
	}
	ticketURL := strings.TrimRight(s.notifications.baseURL(ctx), "/") + path
	err := s.notifications.Send(ctx, NotificationEmailSendInput{
		Event: event, RecipientEmail: recipient.Email, RecipientName: name, UserID: recipient.ID,
		SourceType: "support_ticket", SourceID: strconv.FormatInt(ticket.ID, 10), ReminderKey: strconv.FormatInt(messageID, 10),
		Variables: map[string]string{
			"ticket_id": strconv.FormatInt(ticket.ID, 10), "ticket_subject": ticket.Subject,
			"ticket_status": ticket.Status, "ticket_url": ticketURL,
		},
	})
	if err != nil {
		slog.Error("ticket notification email failed", "event", event, "ticket_id", ticket.ID, "recipient_id", recipient.ID, "error", err)
	}
}

func validTicketCategory(value string) bool {
	switch value {
	case domain.TicketCategoryAccount, domain.TicketCategoryBilling, domain.TicketCategoryAPI, domain.TicketCategoryModel, domain.TicketCategoryOther:
		return true
	default:
		return false
	}
}

func validTicketStatus(value string) bool {
	switch value {
	case domain.TicketStatusPendingAdmin, domain.TicketStatusPendingUser, domain.TicketStatusClosed:
		return true
	default:
		return false
	}
}

func validTicketPriority(value string) bool {
	switch value {
	case domain.TicketPriorityNormal, domain.TicketPriorityHigh, domain.TicketPriorityUrgent:
		return true
	default:
		return false
	}
}

// TicketAttachmentCleanupService 定时清理关闭超过 30 天的附件。
type TicketAttachmentCleanupService struct {
	service *TicketService
	stop    chan struct{}
}

func NewTicketAttachmentCleanupService(service *TicketService) *TicketAttachmentCleanupService {
	return &TicketAttachmentCleanupService{service: service, stop: make(chan struct{})}
}

func (s *TicketAttachmentCleanupService) Start() {
	go func() {
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
				_, _ = s.service.CleanupAttachments(ctx, time.Now())
				cancel()
			case <-s.stop:
				return
			}
		}
	}()
}

func (s *TicketAttachmentCleanupService) Stop() { close(s.stop) }
