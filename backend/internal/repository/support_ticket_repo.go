// Package repository 持久化站内工单及其不可变消息。
package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type supportTicketRepository struct{ db *sql.DB }

func NewSupportTicketRepository(db *sql.DB) service.TicketRepository {
	return &supportTicketRepository{db: db}
}

func (r *supportTicketRepository) Create(ctx context.Context, input service.CreateTicketRecordInput) (*service.Ticket, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var lockedID int64
	if err := tx.QueryRowContext(ctx, `SELECT id FROM users WHERE id = $1 AND deleted_at IS NULL FOR UPDATE`, input.UserID).Scan(&lockedID); err != nil {
		return nil, service.ErrTicketNotFound
	}
	var openCount, dailyCount int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM support_tickets WHERE user_id = $1 AND status <> 'closed'`, input.UserID).Scan(&openCount); err != nil {
		return nil, err
	}
	if openCount >= 5 {
		return nil, service.ErrTicketOpenLimit
	}
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM support_tickets WHERE user_id = $1 AND created_at >= $2`, input.UserID, input.DayStart).Scan(&dailyCount); err != nil {
		return nil, err
	}
	if dailyCount >= 10 {
		return nil, service.ErrTicketDailyLimit
	}

	var ticket service.Ticket
	err = tx.QueryRowContext(ctx, `
		INSERT INTO support_tickets (user_id, subject, category, status, priority, last_message_at, created_at, updated_at)
		VALUES ($1, $2, $3, 'pending_admin', 'normal', $4, $4, $4)
		RETURNING id, user_id, subject, category, status, priority, assignee_id, closed_by, closed_at, last_message_at, created_at, updated_at`,
		input.UserID, input.Subject, input.Category, input.Now,
	).Scan(&ticket.ID, &ticket.UserID, &ticket.Subject, &ticket.Category, &ticket.Status, &ticket.Priority, &ticket.AssigneeID, &ticket.ClosedBy, &ticket.ClosedAt, &ticket.LastMessageAt, &ticket.CreatedAt, &ticket.UpdatedAt)
	if err != nil {
		return nil, err
	}

	messageID, err := insertTicketMessage(ctx, tx, ticket.ID, &input.UserID, service.TicketMessageKindPublic, service.TicketVisibilityUser, input.Body, nil, input.Now)
	if err != nil {
		return nil, err
	}
	if err := insertTicketAttachments(ctx, tx, messageID, input.UserID, input.Attachments, input.Now); err != nil {
		return nil, err
	}
	if err := upsertTicketRead(ctx, tx, ticket.ID, input.UserID, messageID, input.Now); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return r.Get(ctx, ticket.ID, input.UserID, false)
}

func (r *supportTicketRepository) List(ctx context.Context, viewerID int64, isAdmin bool, params pagination.PaginationParams, filters service.TicketListFilters) ([]service.Ticket, *pagination.PaginationResult, error) {
	args := []any{viewerID}
	where := []string{"$1::bigint > 0"}
	if !isAdmin {
		where = append(where, "t.user_id = $1")
	}
	add := func(clause string, value any) {
		args = append(args, value)
		where = append(where, strings.ReplaceAll(clause, "?", "$"+strconv.Itoa(len(args))))
	}
	if filters.Status != "" {
		add("t.status = ?", filters.Status)
	}
	if filters.Category != "" {
		add("t.category = ?", filters.Category)
	}
	if filters.Priority != "" {
		add("t.priority = ?", filters.Priority)
	}
	if isAdmin {
		switch filters.Assignee {
		case "mine":
			where = append(where, "t.assignee_id = $1")
		case "unassigned":
			where = append(where, "t.assignee_id IS NULL")
		case "":
		default:
			if id, err := strconv.ParseInt(filters.Assignee, 10, 64); err == nil && id > 0 {
				add("t.assignee_id = ?", id)
			}
		}
		if filters.Search != "" {
			args = append(args, "%"+strings.ToLower(filters.Search)+"%")
			p := "$" + strconv.Itoa(len(args))
			where = append(where, fmt.Sprintf("(LOWER(t.subject) LIKE %s OR LOWER(u.email) LIKE %s OR LOWER(u.username) LIKE %s OR t.id::text = %s)", p, p, p, p))
		}
	}

	whereSQL := strings.Join(where, " AND ")
	var total int64
	countQuery := `SELECT COUNT(*) FROM support_tickets t JOIN users u ON u.id = t.user_id WHERE ` + whereSQL
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, nil, err
	}

	limit, offset := params.Limit(), params.Offset()
	args = append(args, limit, offset)
	order := "t.last_message_at DESC, t.id DESC"
	if isAdmin {
		order = `CASE t.status WHEN 'pending_admin' THEN 0 WHEN 'pending_user' THEN 1 ELSE 2 END,
			CASE t.priority WHEN 'urgent' THEN 0 WHEN 'high' THEN 1 ELSE 2 END,
			t.last_message_at ASC, t.id ASC`
	}
	query := fmt.Sprintf(`
		SELECT t.id, t.user_id, t.subject, t.category, t.status, t.priority, t.assignee_id, t.closed_by,
			t.closed_at, t.last_message_at, t.created_at, t.updated_at,
			u.email, u.username, COALESCE(a.email, ''), COALESCE(a.username, ''),
			EXISTS (
				SELECT 1 FROM support_ticket_messages m
				WHERE m.ticket_id = t.id AND m.id > COALESCE(tr.last_read_message_id, 0) %s
			) AS unread
		FROM support_tickets t
		JOIN users u ON u.id = t.user_id
		LEFT JOIN users a ON a.id = t.assignee_id
		LEFT JOIN support_ticket_reads tr ON tr.ticket_id = t.id AND tr.user_id = $1
		WHERE %s ORDER BY %s LIMIT $%d OFFSET $%d`,
		visibilitySQL(isAdmin), whereSQL, order, len(args)-1, len(args))
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	items := make([]service.Ticket, 0, limit)
	for rows.Next() {
		item, err := scanTicket(rows)
		if err != nil {
			return nil, nil, err
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}
	pages := 0
	if limit > 0 {
		pages = int((total + int64(limit) - 1) / int64(limit))
	}
	return items, &pagination.PaginationResult{Total: total, Page: params.Page, PageSize: limit, Pages: pages}, nil
}

func (r *supportTicketRepository) Get(ctx context.Context, ticketID, viewerID int64, isAdmin bool) (*service.Ticket, error) {
	ownerClause := ""
	if !isAdmin {
		ownerClause = " AND t.user_id = $3"
	}
	query := `
		SELECT t.id, t.user_id, t.subject, t.category, t.status, t.priority, t.assignee_id, t.closed_by,
			t.closed_at, t.last_message_at, t.created_at, t.updated_at,
			u.email, u.username, COALESCE(a.email, ''), COALESCE(a.username, ''),
			EXISTS (SELECT 1 FROM support_ticket_messages m WHERE m.ticket_id = t.id AND m.id > COALESCE(tr.last_read_message_id, 0) ` + visibilitySQL(isAdmin) + `) AS unread
		FROM support_tickets t
		JOIN users u ON u.id = t.user_id
		LEFT JOIN users a ON a.id = t.assignee_id
		LEFT JOIN support_ticket_reads tr ON tr.ticket_id = t.id AND tr.user_id = $2
		WHERE t.id = $1` + ownerClause
	args := []any{ticketID, viewerID}
	if !isAdmin {
		args = append(args, viewerID)
	}
	item, err := scanTicket(r.db.QueryRowContext(ctx, query, args...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, service.ErrTicketNotFound
		}
		return nil, err
	}
	messages, err := r.listMessages(ctx, ticketID, isAdmin)
	if err != nil {
		return nil, err
	}
	item.Messages = messages
	return item, nil
}

func (r *supportTicketRepository) AddReply(ctx context.Context, input service.AddTicketReplyRecordInput) (*service.Ticket, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var ownerID int64
	var oldStatus string
	if err := tx.QueryRowContext(ctx, `SELECT user_id, status FROM support_tickets WHERE id = $1 FOR UPDATE`, input.TicketID).Scan(&ownerID, &oldStatus); err != nil {
		return nil, service.ErrTicketNotFound
	}
	if !input.IsAdmin && ownerID != input.ActorID {
		return nil, service.ErrTicketNotFound
	}
	visibility := service.TicketVisibilityUser
	if input.Kind == service.TicketMessageKindInternal {
		if !input.IsAdmin {
			return nil, service.ErrTicketNotFound
		}
		visibility = service.TicketVisibilityAdmin
	}
	messageID, err := insertTicketMessage(ctx, tx, input.TicketID, &input.ActorID, input.Kind, visibility, input.Body, nil, input.Now)
	if err != nil {
		return nil, err
	}
	if err := insertTicketAttachments(ctx, tx, messageID, input.ActorID, input.Attachments, input.Now); err != nil {
		return nil, err
	}
	if input.Kind == service.TicketMessageKindPublic {
		status := service.TicketStatusPendingAdmin
		if input.IsAdmin {
			status = service.TicketStatusPendingUser
		}
		_, err = tx.ExecContext(ctx, `UPDATE support_tickets SET status = $2, closed_at = NULL, closed_by = NULL, last_message_at = $3, updated_at = $3 WHERE id = $1`, input.TicketID, status, input.Now)
		if err != nil {
			return nil, err
		}
		if !input.IsAdmin && oldStatus == service.TicketStatusClosed {
			if _, err = tx.ExecContext(ctx, `UPDATE support_ticket_attachments a SET delete_after = NULL FROM support_ticket_messages m WHERE a.message_id = m.id AND m.ticket_id = $1 AND a.deleted_at IS NULL`, input.TicketID); err != nil {
				return nil, err
			}
			metadata, _ := json.Marshal(map[string]any{"event": "reopened"})
			messageID, err = insertTicketMessage(ctx, tx, input.TicketID, nil, service.TicketMessageKindSystem, service.TicketVisibilityUser, "", metadata, input.Now)
			if err != nil {
				return nil, err
			}
		}
	} else {
		_, err = tx.ExecContext(ctx, `UPDATE support_tickets SET last_message_at = $2, updated_at = $2 WHERE id = $1`, input.TicketID, input.Now)
		if err != nil {
			return nil, err
		}
	}
	if err := upsertTicketRead(ctx, tx, input.TicketID, input.ActorID, messageID, input.Now); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return r.Get(ctx, input.TicketID, input.ActorID, input.IsAdmin)
}

func (r *supportTicketRepository) Update(ctx context.Context, input service.UpdateTicketRecordInput) (*service.Ticket, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	var oldPriority, oldStatus string
	var oldAssignee *int64
	if err := tx.QueryRowContext(ctx, `SELECT priority, status, assignee_id FROM support_tickets WHERE id = $1 FOR UPDATE`, input.TicketID).Scan(&oldPriority, &oldStatus, &oldAssignee); err != nil {
		return nil, service.ErrTicketNotFound
	}
	if input.SetAssignee && input.AssigneeID != nil {
		var valid bool
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND role = 'admin' AND status = 'active' AND deleted_at IS NULL)`, *input.AssigneeID).Scan(&valid); err != nil {
			return nil, err
		}
		if !valid {
			return nil, service.ErrTicketAssignee
		}
	}
	priority := oldPriority
	if input.Priority != nil {
		priority = *input.Priority
	}
	status := oldStatus
	var closedAt *time.Time
	var closedBy *int64
	if input.SetClosed {
		if input.Closed {
			status, closedAt, closedBy = service.TicketStatusClosed, &input.Now, &input.ActorID
		} else {
			status = service.TicketStatusPendingAdmin
		}
	} else if oldStatus == service.TicketStatusClosed {
		var existingClosedAt time.Time
		var existingClosedBy int64
		_ = tx.QueryRowContext(ctx, `SELECT closed_at, closed_by FROM support_tickets WHERE id = $1`, input.TicketID).Scan(&existingClosedAt, &existingClosedBy)
		closedAt, closedBy = &existingClosedAt, &existingClosedBy
	}
	assignee := oldAssignee
	if input.SetAssignee {
		assignee = input.AssigneeID
	}
	_, err = tx.ExecContext(ctx, `UPDATE support_tickets SET priority=$2, status=$3, assignee_id=$4, closed_at=$5, closed_by=$6, updated_at=$7 WHERE id=$1`, input.TicketID, priority, status, assignee, closedAt, closedBy, input.Now)
	if err != nil {
		return nil, err
	}
	metadata, _ := json.Marshal(map[string]any{
		"event":           "ticket_updated",
		"old_priority":    oldPriority,
		"priority":        priority,
		"old_status":      oldStatus,
		"status":          status,
		"old_assignee_id": oldAssignee,
		"assignee_id":     assignee,
	})
	messageID, err := insertTicketMessage(ctx, tx, input.TicketID, &input.ActorID, service.TicketMessageKindSystem, service.TicketVisibilityAdmin, "", metadata, input.Now)
	if err != nil {
		return nil, err
	}
	if input.SetClosed {
		publicMetadata, _ := json.Marshal(map[string]any{"event": map[bool]string{true: "closed", false: "reopened"}[input.Closed]})
		messageID, err = insertTicketMessage(ctx, tx, input.TicketID, &input.ActorID, service.TicketMessageKindSystem, service.TicketVisibilityUser, "", publicMetadata, input.Now)
		if err != nil {
			return nil, err
		}
		if input.Closed {
			_, err = tx.ExecContext(ctx, `UPDATE support_ticket_attachments a SET delete_after=$2 FROM support_ticket_messages m WHERE a.message_id=m.id AND m.ticket_id=$1 AND a.deleted_at IS NULL`, input.TicketID, input.DeleteAfter)
		} else {
			_, err = tx.ExecContext(ctx, `UPDATE support_ticket_attachments a SET delete_after=NULL FROM support_ticket_messages m WHERE a.message_id=m.id AND m.ticket_id=$1 AND a.deleted_at IS NULL`, input.TicketID)
		}
		if err != nil {
			return nil, err
		}
	}
	if _, err = tx.ExecContext(ctx, `UPDATE support_tickets SET last_message_at=$2 WHERE id=$1`, input.TicketID, input.Now); err != nil {
		return nil, err
	}
	if err := upsertTicketRead(ctx, tx, input.TicketID, input.ActorID, messageID, input.Now); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return r.Get(ctx, input.TicketID, input.ActorID, true)
}

func (r *supportTicketRepository) MarkRead(ctx context.Context, ticketID, viewerID int64, isAdmin bool, at time.Time) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	var ownerID int64
	if err := tx.QueryRowContext(ctx, `SELECT user_id FROM support_tickets WHERE id=$1`, ticketID).Scan(&ownerID); err != nil || (!isAdmin && ownerID != viewerID) {
		return service.ErrTicketNotFound
	}
	query := `SELECT COALESCE(MAX(id),0) FROM support_ticket_messages WHERE ticket_id=$1`
	if !isAdmin {
		query += ` AND visibility='user'`
	}
	var messageID int64
	if err := tx.QueryRowContext(ctx, query, ticketID).Scan(&messageID); err != nil {
		return err
	}
	if err := upsertTicketRead(ctx, tx, ticketID, viewerID, messageID, at); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *supportTicketRepository) UnreadCount(ctx context.Context, viewerID int64, isAdmin bool) (int64, error) {
	owner := "AND t.user_id=$1"
	if isAdmin {
		owner = ""
	}
	query := `SELECT COUNT(*) FROM support_tickets t LEFT JOIN support_ticket_reads tr ON tr.ticket_id=t.id AND tr.user_id=$1
		WHERE EXISTS (SELECT 1 FROM support_ticket_messages m WHERE m.ticket_id=t.id AND m.id>COALESCE(tr.last_read_message_id,0) ` + visibilitySQL(isAdmin) + `) ` + owner
	var count int64
	err := r.db.QueryRowContext(ctx, query, viewerID).Scan(&count)
	return count, err
}

func (r *supportTicketRepository) GetAttachment(ctx context.Context, attachmentID, viewerID int64, isAdmin bool) (*service.TicketAttachment, error) {
	query := `SELECT a.id,a.message_id,a.uploader_id,a.original_name,a.storage_key,a.media_type,a.size_bytes,a.delete_after,a.deleted_at,a.deleted_by,a.delete_reason,a.created_at
		FROM support_ticket_attachments a JOIN support_ticket_messages m ON m.id=a.message_id JOIN support_tickets t ON t.id=m.ticket_id
		WHERE a.id=$1 AND a.deleted_at IS NULL`
	args := []any{attachmentID}
	if !isAdmin {
		query += ` AND t.user_id=$2 AND m.visibility='user'`
		args = append(args, viewerID)
	}
	item, err := scanAttachment(r.db.QueryRowContext(ctx, query, args...))
	if err != nil {
		return nil, service.ErrTicketAttachmentGone
	}
	return item, nil
}

func (r *supportTicketRepository) DeleteAttachment(ctx context.Context, attachmentID, actorID int64, reason string, at time.Time) (*service.TicketAttachment, error) {
	item, err := r.GetAttachment(ctx, attachmentID, actorID, true)
	if err != nil {
		return nil, err
	}
	res, err := r.db.ExecContext(ctx, `UPDATE support_ticket_attachments SET deleted_at=$2,deleted_by=$3,delete_reason=$4 WHERE id=$1 AND deleted_at IS NULL`, attachmentID, at, actorID, reason)
	if err != nil {
		return nil, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, service.ErrTicketAttachmentGone
	}
	item.DeletedAt, item.DeletedBy, item.DeleteReason = &at, &actorID, &reason
	return item, nil
}

func (r *supportTicketRepository) ListAttachmentsDue(ctx context.Context, now time.Time, limit int) ([]service.TicketAttachment, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id,message_id,uploader_id,original_name,storage_key,media_type,size_bytes,delete_after,deleted_at,deleted_by,delete_reason,created_at FROM support_ticket_attachments WHERE delete_after <= $1 AND deleted_at IS NULL ORDER BY delete_after,id LIMIT $2`, now, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []service.TicketAttachment
	for rows.Next() {
		item, err := scanAttachment(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	return items, rows.Err()
}

func (r *supportTicketRepository) MarkAttachmentDeleted(ctx context.Context, attachmentID int64, reason string, at time.Time) error {
	_, err := r.db.ExecContext(ctx, `UPDATE support_ticket_attachments SET deleted_at=$2,delete_reason=$3 WHERE id=$1 AND deleted_at IS NULL`, attachmentID, at, reason)
	return err
}

func (r *supportTicketRepository) ListActiveAdmins(ctx context.Context) ([]service.TicketUserSummary, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id,email,username FROM users WHERE role='admin' AND status='active' AND deleted_at IS NULL ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []service.TicketUserSummary
	for rows.Next() {
		var item service.TicketUserSummary
		if err := rows.Scan(&item.ID, &item.Email, &item.Username); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *supportTicketRepository) GetUserSummary(ctx context.Context, userID int64) (*service.TicketUserSummary, error) {
	var item service.TicketUserSummary
	if err := r.db.QueryRowContext(ctx, `SELECT id,email,username FROM users WHERE id=$1 AND deleted_at IS NULL`, userID).Scan(&item.ID, &item.Email, &item.Username); err != nil {
		return nil, service.ErrUserNotFound
	}
	return &item, nil
}

func (r *supportTicketRepository) listMessages(ctx context.Context, ticketID int64, isAdmin bool) ([]service.TicketMessage, error) {
	query := `SELECT m.id,m.ticket_id,m.author_id,m.kind,m.visibility,m.body,m.metadata,m.created_at,COALESCE(u.email,''),COALESCE(u.username,''),COALESCE(u.role,'') FROM support_ticket_messages m LEFT JOIN users u ON u.id=m.author_id WHERE m.ticket_id=$1`
	if !isAdmin {
		query += ` AND m.visibility='user'`
	}
	query += ` ORDER BY m.id`
	rows, err := r.db.QueryContext(ctx, query, ticketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []service.TicketMessage{}
	messageIndex := map[int64]int{}
	for rows.Next() {
		var item service.TicketMessage
		var metadata []byte
		var email, username, role string
		if err := rows.Scan(&item.ID, &item.TicketID, &item.AuthorID, &item.Kind, &item.Visibility, &item.Body, &metadata, &item.CreatedAt, &email, &username, &role); err != nil {
			return nil, err
		}
		item.Metadata = metadata
		item.Attachments = []service.TicketAttachment{}
		if item.AuthorID != nil {
			item.Author = &service.TicketUserSummary{ID: *item.AuthorID, Email: email, Username: username}
			item.AuthorRole = role
		}
		messageIndex[item.ID] = len(items)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	attQuery := `SELECT a.id,a.message_id,a.uploader_id,a.original_name,a.storage_key,a.media_type,a.size_bytes,a.delete_after,a.deleted_at,a.deleted_by,a.delete_reason,a.created_at FROM support_ticket_attachments a JOIN support_ticket_messages m ON m.id=a.message_id WHERE m.ticket_id=$1`
	if !isAdmin {
		attQuery += ` AND m.visibility='user'`
	}
	attRows, err := r.db.QueryContext(ctx, attQuery, ticketID)
	if err != nil {
		return nil, err
	}
	defer attRows.Close()
	for attRows.Next() {
		item, err := scanAttachment(attRows)
		if err != nil {
			return nil, err
		}
		if index, ok := messageIndex[item.MessageID]; ok {
			items[index].Attachments = append(items[index].Attachments, *item)
		}
	}
	return items, attRows.Err()
}

type ticketRowScanner interface{ Scan(...any) error }

func scanTicket(row ticketRowScanner) (*service.Ticket, error) {
	var item service.Ticket
	var assigneeEmail, assigneeName string
	if err := row.Scan(&item.ID, &item.UserID, &item.Subject, &item.Category, &item.Status, &item.Priority, &item.AssigneeID, &item.ClosedBy, &item.ClosedAt, &item.LastMessageAt, &item.CreatedAt, &item.UpdatedAt, &item.User.Email, &item.User.Username, &assigneeEmail, &assigneeName, &item.Unread); err != nil {
		return nil, err
	}
	item.User.ID = item.UserID
	if item.AssigneeID != nil {
		item.Assignee = &service.TicketUserSummary{ID: *item.AssigneeID, Email: assigneeEmail, Username: assigneeName}
	}
	return &item, nil
}

func scanAttachment(row ticketRowScanner) (*service.TicketAttachment, error) {
	var item service.TicketAttachment
	err := row.Scan(&item.ID, &item.MessageID, &item.UploaderID, &item.OriginalName, &item.StorageKey, &item.MediaType, &item.SizeBytes, &item.DeleteAfter, &item.DeletedAt, &item.DeletedBy, &item.DeleteReason, &item.CreatedAt)
	return &item, err
}

func insertTicketMessage(ctx context.Context, tx *sql.Tx, ticketID int64, authorID *int64, kind, visibility, body string, metadata []byte, at time.Time) (int64, error) {
	var id int64
	err := tx.QueryRowContext(ctx, `INSERT INTO support_ticket_messages (ticket_id,author_id,kind,visibility,body,metadata,created_at) VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id`, ticketID, authorID, kind, visibility, body, nullableJSON(metadata), at).Scan(&id)
	return id, err
}

func insertTicketAttachments(ctx context.Context, tx *sql.Tx, messageID, uploaderID int64, items []service.TicketAttachment, at time.Time) error {
	for i := range items {
		item := items[i]
		_, err := tx.ExecContext(ctx, `INSERT INTO support_ticket_attachments (message_id,uploader_id,original_name,storage_key,media_type,size_bytes,created_at) VALUES ($1,$2,$3,$4,$5,$6,$7)`, messageID, uploaderID, item.OriginalName, item.StorageKey, item.MediaType, item.SizeBytes, at)
		if err != nil {
			return err
		}
	}
	return nil
}

func upsertTicketRead(ctx context.Context, tx *sql.Tx, ticketID, userID, messageID int64, at time.Time) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO support_ticket_reads (ticket_id,user_id,last_read_message_id,read_at) VALUES ($1,$2,$3,$4) ON CONFLICT (ticket_id,user_id) DO UPDATE SET last_read_message_id=GREATEST(support_ticket_reads.last_read_message_id,EXCLUDED.last_read_message_id),read_at=EXCLUDED.read_at`, ticketID, userID, messageID, at)
	return err
}

func visibilitySQL(isAdmin bool) string {
	if isAdmin {
		return ""
	}
	return "AND m.visibility='user'"
}

func nullableJSON(value []byte) any {
	if len(value) == 0 || !json.Valid(value) {
		return nil
	}
	return value
}
