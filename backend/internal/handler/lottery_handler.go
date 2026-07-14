// Package handler 提供登录用户的抽奖摘要、执行和历史接口。
package handler

import (
	"context"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

type LotteryHandler struct{ service *service.LotteryService }

func NewLotteryHandler(lotteryService *service.LotteryService) *LotteryHandler {
	return &LotteryHandler{service: lotteryService}
}

func (h *LotteryHandler) Summary(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}
	result, err := h.service.Summary(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}

func (h *LotteryHandler) Draw(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}
	poolKey := strings.TrimSpace(c.Param("key"))
	payload := gin.H{"pool_key": poolKey}
	executeUserIdempotentJSON(c, "user.lottery.draw."+poolKey, payload, service.DefaultWriteIdempotencyTTL(), func(ctx context.Context) (any, error) {
		return h.service.Draw(ctx, subject.UserID, poolKey, c.GetHeader("Idempotency-Key"))
	})
}

func (h *LotteryHandler) History(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}
	page, pageSize := response.ParsePagination(c)
	items, result, err := h.service.ListUserDraws(c.Request.Context(), subject.UserID, pagination.PaginationParams{Page: page, PageSize: pageSize}, strings.TrimSpace(c.Query("pool")))
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, items, result.Total, page, pageSize)
}
