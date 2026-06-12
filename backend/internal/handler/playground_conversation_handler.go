package handler

import (
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// PlaygroundConversationHandler 处理对话广场会话 CRUD 请求。
//
// 所有操作均要求 JWT 鉴权（jwtAuth 中间件），并从 JWT subject 中取 userID，
// 保证用户只能操作自己的会话（越权防护双保险：handler 层取 userID + repo 层 SQL WHERE）。
type PlaygroundConversationHandler struct {
	conversationService *service.PlaygroundConversationService
}

// NewPlaygroundConversationHandler 创建 PlaygroundConversationHandler 实例。
func NewPlaygroundConversationHandler(
	conversationService *service.PlaygroundConversationService,
) *PlaygroundConversationHandler {
	return &PlaygroundConversationHandler{
		conversationService: conversationService,
	}
}

// ListConversations 返回当前用户的所有会话摘要列表（不含 messages）。
//
// GET /api/v1/playground/conversations
func (h *PlaygroundConversationHandler) ListConversations(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	items, err := h.conversationService.List(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	out := make([]*dto.PlaygroundConversationSummaryDTO, 0, len(items))
	for i := range items {
		out = append(out, dto.ConversationSummaryFromService(&items[i]))
	}
	response.Success(c, out)
}

// GetConversation 获取指定会话的完整数据（含 messages）。
//
// GET /api/v1/playground/conversations/:id
func (h *PlaygroundConversationHandler) GetConversation(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	id, err := parseConversationID(c)
	if err != nil {
		response.BadRequest(c, "Invalid conversation ID")
		return
	}

	conv, err := h.conversationService.Get(c.Request.Context(), id, subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dto.ConversationDetailFromService(conv))
}

// CreateConversation 新建会话。
//
// POST /api/v1/playground/conversations
func (h *PlaygroundConversationHandler) CreateConversation(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req dto.CreateConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	conv, err := h.conversationService.Create(
		c.Request.Context(),
		subject.UserID,
		req.Title,
		req.Model,
		req.GroupName,
		req.Messages,
	)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Created(c, dto.ConversationDetailFromService(conv))
}

// UpdateConversation 部分更新会话（支持更新 title/model/group_name/messages）。
//
// PUT /api/v1/playground/conversations/:id
func (h *PlaygroundConversationHandler) UpdateConversation(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	id, err := parseConversationID(c)
	if err != nil {
		response.BadRequest(c, "Invalid conversation ID")
		return
	}

	var req dto.UpdateConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	if err := h.conversationService.Update(
		c.Request.Context(),
		id,
		subject.UserID,
		req.Title,
		req.Model,
		req.GroupName,
		req.Messages,
	); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{"message": "ok"})
}

// DeleteConversation 删除指定会话。
//
// DELETE /api/v1/playground/conversations/:id
func (h *PlaygroundConversationHandler) DeleteConversation(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	id, err := parseConversationID(c)
	if err != nil {
		response.BadRequest(c, "Invalid conversation ID")
		return
	}

	if err := h.conversationService.Delete(c.Request.Context(), id, subject.UserID); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{"message": "ok"})
}

// parseConversationID 从路径参数 :id 解析会话 ID。
// 非法（非正整数）时返回错误。
func parseConversationID(c *gin.Context) (int64, error) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, err
	}
	return id, nil
}
