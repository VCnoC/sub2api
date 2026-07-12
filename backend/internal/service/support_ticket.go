// Package service 提供站内工单业务契约与领域错误。
package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/domain"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

const (
	TicketStatusPendingAdmin = domain.TicketStatusPendingAdmin
	TicketStatusPendingUser  = domain.TicketStatusPendingUser
	TicketStatusClosed       = domain.TicketStatusClosed

	TicketCategoryAccount = domain.TicketCategoryAccount
	TicketCategoryBilling = domain.TicketCategoryBilling
	TicketCategoryAPI     = domain.TicketCategoryAPI
	TicketCategoryModel   = domain.TicketCategoryModel
	TicketCategoryOther   = domain.TicketCategoryOther

	TicketPriorityNormal = domain.TicketPriorityNormal
	TicketPriorityHigh   = domain.TicketPriorityHigh
	TicketPriorityUrgent = domain.TicketPriorityUrgent

	TicketMessageKindPublic   = domain.TicketMessageKindPublic
	TicketMessageKindInternal = domain.TicketMessageKindInternal
	TicketMessageKindSystem   = domain.TicketMessageKindSystem

	TicketVisibilityUser  = domain.TicketVisibilityUser
	TicketVisibilityAdmin = domain.TicketVisibilityAdmin
)

var (
	ErrTicketNotFound       = infraerrors.NotFound("TICKET_NOT_FOUND", "ticket not found")
	ErrTicketInvalidInput   = infraerrors.BadRequest("TICKET_INVALID_INPUT", "ticket input is invalid")
	ErrTicketInvalidStatus  = infraerrors.BadRequest("TICKET_INVALID_STATUS", "ticket status is invalid")
	ErrTicketInvalidFile    = infraerrors.BadRequest("TICKET_ATTACHMENT_INVALID", "ticket attachment is invalid")
	ErrTicketOpenLimit      = infraerrors.TooManyRequests("TICKET_OPEN_LIMIT_REACHED", "too many open tickets")
	ErrTicketDailyLimit     = infraerrors.TooManyRequests("TICKET_DAILY_LIMIT_REACHED", "daily ticket limit reached")
	ErrTicketAssignee       = infraerrors.BadRequest("TICKET_ASSIGNEE_INVALID", "ticket assignee must be an active administrator")
	ErrTicketAttachmentGone = infraerrors.NotFound("TICKET_ATTACHMENT_NOT_FOUND", "ticket attachment not found")
)

type TicketUserSummary struct {
	ID       int64  `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

type TicketAttachment struct {
	ID           int64      `json:"id"`
	MessageID    int64      `json:"message_id"`
	UploaderID   int64      `json:"uploader_id"`
	OriginalName string     `json:"original_name"`
	StorageKey   string     `json:"-"`
	MediaType    string     `json:"media_type"`
	SizeBytes    int64      `json:"size_bytes"`
	DeleteAfter  *time.Time `json:"delete_after,omitempty"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
	DeletedBy    *int64     `json:"deleted_by,omitempty"`
	DeleteReason *string    `json:"delete_reason,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

type TicketMessage struct {
	ID          int64              `json:"id"`
	TicketID    int64              `json:"ticket_id"`
	AuthorID    *int64             `json:"author_id,omitempty"`
	Author      *TicketUserSummary `json:"author,omitempty"`
	AuthorRole  string             `json:"author_role,omitempty"`
	Kind        string             `json:"kind"`
	Visibility  string             `json:"visibility"`
	Body        string             `json:"body"`
	Metadata    json.RawMessage    `json:"metadata,omitempty"`
	Attachments []TicketAttachment `json:"attachments"`
	CreatedAt   time.Time          `json:"created_at"`
}

type Ticket struct {
	ID            int64              `json:"id"`
	UserID        int64              `json:"user_id"`
	User          TicketUserSummary  `json:"user"`
	Subject       string             `json:"subject"`
	Category      string             `json:"category"`
	Status        string             `json:"status"`
	Priority      string             `json:"priority"`
	AssigneeID    *int64             `json:"assignee_id,omitempty"`
	Assignee      *TicketUserSummary `json:"assignee,omitempty"`
	ClosedBy      *int64             `json:"closed_by,omitempty"`
	ClosedAt      *time.Time         `json:"closed_at,omitempty"`
	LastMessageAt time.Time          `json:"last_message_at"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`
	Unread        bool               `json:"unread"`
	Messages      []TicketMessage    `json:"messages,omitempty"`
}

type TicketListFilters struct {
	Status   string
	Category string
	Priority string
	Assignee string
	Search   string
}

type CreateTicketRecordInput struct {
	UserID      int64
	Subject     string
	Category    string
	Body        string
	Attachments []TicketAttachment
	Now         time.Time
	DayStart    time.Time
}

type AddTicketReplyRecordInput struct {
	TicketID    int64
	ActorID     int64
	IsAdmin     bool
	Kind        string
	Body        string
	Attachments []TicketAttachment
	Now         time.Time
}

type UpdateTicketRecordInput struct {
	TicketID    int64
	ActorID     int64
	Priority    *string
	SetAssignee bool
	AssigneeID  *int64
	SetClosed   bool
	Closed      bool
	Now         time.Time
	DeleteAfter time.Time
}

type TicketRepository interface {
	Create(ctx context.Context, input CreateTicketRecordInput) (*Ticket, error)
	List(ctx context.Context, viewerID int64, isAdmin bool, params pagination.PaginationParams, filters TicketListFilters) ([]Ticket, *pagination.PaginationResult, error)
	Get(ctx context.Context, ticketID, viewerID int64, isAdmin bool) (*Ticket, error)
	AddReply(ctx context.Context, input AddTicketReplyRecordInput) (*Ticket, error)
	Update(ctx context.Context, input UpdateTicketRecordInput) (*Ticket, error)
	MarkRead(ctx context.Context, ticketID, viewerID int64, isAdmin bool, at time.Time) error
	UnreadCount(ctx context.Context, viewerID int64, isAdmin bool) (int64, error)
	GetAttachment(ctx context.Context, attachmentID, viewerID int64, isAdmin bool) (*TicketAttachment, error)
	DeleteAttachment(ctx context.Context, attachmentID, actorID int64, reason string, at time.Time) (*TicketAttachment, error)
	ListAttachmentsDue(ctx context.Context, now time.Time, limit int) ([]TicketAttachment, error)
	MarkAttachmentDeleted(ctx context.Context, attachmentID int64, reason string, at time.Time) error
	ListActiveAdmins(ctx context.Context) ([]TicketUserSummary, error)
	GetUserSummary(ctx context.Context, userID int64) (*TicketUserSummary, error)
}
