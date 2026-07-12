// Package handler 提供登录用户的工单 HTTP 接口。
package handler

import (
	"mime"
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
	items, result, err := h.service.List(c.Request.Context(), subject.UserID, false, pagination.PaginationParams{Page: page, PageSize: pageSize}, service.TicketListFilters{Status: strings.TrimSpace(c.Query("status")), Category: strings.TrimSpace(c.Query("category"))})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, items, result.Total, page, pageSize)
}

func (h *SupportTicketHandler) Create(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}
	files, ok := parseTicketMultipart(c)
	if !ok {
		return
	}
	item, err := h.service.Create(c.Request.Context(), service.CreateTicketInput{
		UserID: subject.UserID, Subject: c.PostForm("subject"), Category: c.PostForm("category"), Body: c.PostForm("body"), Files: files,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, item)
}

func (h *SupportTicketHandler) Get(c *gin.Context) {
	subject, ticketID, ok := ticketSubjectAndID(c)
	if !ok {
		return
	}
	item, err := h.service.Get(c.Request.Context(), ticketID, subject.UserID, false)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, item)
}

func (h *SupportTicketHandler) Reply(c *gin.Context) {
	subject, ticketID, ok := ticketSubjectAndID(c)
	if !ok {
		return
	}
	files, ok := parseTicketMultipart(c)
	if !ok {
		return
	}
	item, err := h.service.Reply(c.Request.Context(), service.ReplyTicketInput{TicketID: ticketID, ActorID: subject.UserID, Body: c.PostForm("body"), Files: files})
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
	count, err := h.service.UnreadCount(c.Request.Context(), subject.UserID, false)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"count": count})
}

func (h *SupportTicketHandler) DownloadAttachment(c *gin.Context) {
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
	role, _ := middleware2.GetUserRoleFromContext(c)
	item, file, err := h.service.OpenAttachment(c.Request.Context(), attachmentID, subject.UserID, role == service.RoleAdmin)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	defer file.Close()
	stat, err := file.Stat()
	if err != nil {
		response.ErrorFrom(c, service.ErrTicketAttachmentGone)
		return
	}
	disposition := "attachment"
	if strings.HasPrefix(item.MediaType, "image/") {
		disposition = "inline"
	}
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Content-Disposition", mime.FormatMediaType(disposition, map[string]string{"filename": item.OriginalName}))
	c.DataFromReader(http.StatusOK, stat.Size(), item.MediaType, file, nil)
}

func parseTicketMultipart(c *gin.Context) ([]*multipart.FileHeader, bool) {
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

func ticketSubjectAndID(c *gin.Context) (middleware2.AuthSubject, int64, bool) {
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
