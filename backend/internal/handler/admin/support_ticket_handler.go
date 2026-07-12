// Package admin 提供管理员工单管理 HTTP 接口。
package admin

import (
	"encoding/json"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

type SupportTicketHandler struct{ service *service.TicketService }

func NewSupportTicketHandler(service *service.TicketService) *SupportTicketHandler {
	return &SupportTicketHandler{service: service}
}

func (h *SupportTicketHandler) List(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}
	page, pageSize := response.ParsePagination(c)
	items, result, err := h.service.List(c.Request.Context(), subject.UserID, true, pagination.PaginationParams{Page: page, PageSize: pageSize}, service.TicketListFilters{
		Status: strings.TrimSpace(c.Query("status")), Category: strings.TrimSpace(c.Query("category")), Priority: strings.TrimSpace(c.Query("priority")), Assignee: strings.TrimSpace(c.Query("assignee")), Search: strings.TrimSpace(c.Query("search")),
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, items, result.Total, page, pageSize)
}

func (h *SupportTicketHandler) Get(c *gin.Context) {
	subject, ticketID, ok := adminTicketSubjectAndID(c)
	if !ok {
		return
	}
	item, err := h.service.Get(c.Request.Context(), ticketID, subject.UserID, true)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, item)
}

func (h *SupportTicketHandler) Reply(c *gin.Context) {
	subject, ticketID, ok := adminTicketSubjectAndID(c)
	if !ok {
		return
	}
	files, ok := parseAdminTicketMultipart(c)
	if !ok {
		return
	}
	internal, _ := strconv.ParseBool(c.PostForm("internal"))
	item, err := h.service.Reply(c.Request.Context(), service.ReplyTicketInput{TicketID: ticketID, ActorID: subject.UserID, IsAdmin: true, Internal: internal, Body: c.PostForm("body"), Files: files})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, item)
}

type updateSupportTicketRequest struct {
	Priority   *string         `json:"priority"`
	AssigneeID json.RawMessage `json:"assignee_id"`
	Closed     *bool           `json:"closed"`
}

func (h *SupportTicketHandler) Update(c *gin.Context) {
	subject, ticketID, ok := adminTicketSubjectAndID(c)
	if !ok {
		return
	}
	var req updateSupportTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorFrom(c, service.ErrTicketInvalidInput)
		return
	}
	input := service.UpdateTicketInput{TicketID: ticketID, ActorID: subject.UserID, Priority: req.Priority, Closed: req.Closed}
	if len(req.AssigneeID) > 0 {
		input.SetAssignee = true
		if string(req.AssigneeID) != "null" {
			var id int64
			if err := json.Unmarshal(req.AssigneeID, &id); err != nil || id <= 0 {
				response.ErrorFrom(c, service.ErrTicketAssignee)
				return
			}
			input.AssigneeID = &id
		}
	}
	item, err := h.service.Update(c.Request.Context(), input)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, item)
}

func (h *SupportTicketHandler) UnreadCount(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}
	count, err := h.service.UnreadCount(c.Request.Context(), subject.UserID, true)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"count": count})
}

type deleteSupportTicketAttachmentRequest struct {
	Reason string `json:"reason"`
}

func (h *SupportTicketHandler) DeleteAttachment(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}
	attachmentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || attachmentID <= 0 {
		response.ErrorFrom(c, service.ErrTicketAttachmentGone)
		return
	}
	var req deleteSupportTicketAttachmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorFrom(c, service.ErrTicketInvalidInput)
		return
	}
	if err := h.service.DeleteAttachment(c.Request.Context(), attachmentID, subject.UserID, req.Reason); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"message": "ok"})
}

func parseAdminTicketMultipart(c *gin.Context) ([]*multipart.FileHeader, bool) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, domain.TicketMaxReplyFileBytes+(1<<20))
	if err := c.Request.ParseMultipartForm(domain.TicketMaxReplyFileBytes); err != nil {
		response.ErrorFrom(c, service.ErrTicketInvalidFile)
		return nil, false
	}
	if c.Request.MultipartForm == nil {
		return nil, true
	}
	return c.Request.MultipartForm.File["files"], true
}

func adminTicketSubjectAndID(c *gin.Context) (middleware2.AuthSubject, int64, bool) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return middleware2.AuthSubject{}, 0, false
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.ErrorFrom(c, service.ErrTicketNotFound)
		return middleware2.AuthSubject{}, 0, false
	}
	return subject, id, true
}
