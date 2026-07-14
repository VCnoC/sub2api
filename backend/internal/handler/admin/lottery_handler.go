// Package admin 提供抽奖奖池、奖品、邀请规则和审计记录管理接口。
package admin

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

type LotteryHandler struct{ service *service.LotteryService }

func NewLotteryHandler(lotteryService *service.LotteryService) *LotteryHandler {
	return &LotteryHandler{service: lotteryService}
}

type lotteryPoolRequest struct {
	Name         string     `json:"name" binding:"required,max=80"`
	Enabled      bool       `json:"enabled"`
	CycleType    string     `json:"cycle_type" binding:"required,oneof=daily weekly"`
	CycleChances int        `json:"cycle_chances" binding:"min=0,max=100"`
	StartsAt     *time.Time `json:"starts_at"`
	EndsAt       *time.Time `json:"ends_at"`
}

type lotteryPrizeRequest struct {
	PoolID         int64    `json:"pool_id" binding:"required,gt=0"`
	Name           string   `json:"name" binding:"required,max=120"`
	Description    string   `json:"description"`
	ImageData      string   `json:"image_data"`
	PrizeType      string   `json:"prize_type" binding:"required,oneof=balance subscription"`
	BalanceAmount  *float64 `json:"balance_amount"`
	GroupID        *int64   `json:"group_id"`
	ValidityDays   *int     `json:"validity_days"`
	ProbabilityPPM int      `json:"probability_ppm" binding:"min=0,max=1000000"`
	StockTotal     *int64   `json:"stock_total"`
	Enabled        bool     `json:"enabled"`
	SortOrder      int      `json:"sort_order"`
}

type lotteryRuleRequest struct {
	Name              string   `json:"name" binding:"required,max=120"`
	EventType         string   `json:"event_type" binding:"required,oneof=signup redeem recharge"`
	Beneficiary       string   `json:"beneficiary" binding:"required,oneof=inviter invitee"`
	NormalChances     int      `json:"normal_chances" binding:"min=0"`
	LuxuryChances     int      `json:"luxury_chances" binding:"min=0"`
	RechargeMode      *string  `json:"recharge_mode"`
	RechargeThreshold *float64 `json:"recharge_threshold"`
	Repeatable        bool     `json:"repeatable"`
	Enabled           bool     `json:"enabled"`
}

func (h *LotteryHandler) ListPools(c *gin.Context) {
	items, err := h.service.ListPools(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, items)
}

func (h *LotteryHandler) UpdatePool(c *gin.Context) {
	var req lotteryPoolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorFrom(c, service.ErrLotteryInvalidInput)
		return
	}
	executeAdminIdempotentJSON(c, "admin.lottery.pool.update", req, service.DefaultWriteIdempotencyTTL(), func(ctx context.Context) (any, error) {
		return h.service.UpdatePool(ctx, c.Param("key"), service.LotteryPoolUpdate{
			Name: req.Name, Enabled: req.Enabled, CycleType: req.CycleType, CycleChances: req.CycleChances,
			StartsAt: req.StartsAt, EndsAt: req.EndsAt,
		})
	})
}

func (h *LotteryHandler) ListPrizes(c *gin.Context) {
	poolID, ok := positiveInt64Query(c, "pool_id")
	if !ok {
		response.ErrorFrom(c, service.ErrLotteryInvalidInput)
		return
	}
	items, err := h.service.ListPrizes(c.Request.Context(), poolID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, items)
}

func (h *LotteryHandler) CreatePrize(c *gin.Context) {
	var req lotteryPrizeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorFrom(c, service.ErrLotteryInvalidInput)
		return
	}
	executeAdminIdempotentJSON(c, "admin.lottery.prize.create", req, service.DefaultWriteIdempotencyTTL(), func(ctx context.Context) (any, error) {
		return h.service.CreatePrize(ctx, prizeInput(req))
	})
}

func (h *LotteryHandler) UpdatePrize(c *gin.Context) {
	id, ok := positivePathID(c)
	if !ok {
		return
	}
	var req lotteryPrizeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorFrom(c, service.ErrLotteryInvalidInput)
		return
	}
	executeAdminIdempotentJSON(c, "admin.lottery.prize.update", req, service.DefaultWriteIdempotencyTTL(), func(ctx context.Context) (any, error) {
		return h.service.UpdatePrize(ctx, id, prizeInput(req))
	})
}

func (h *LotteryHandler) DeletePrize(c *gin.Context) {
	id, ok := positivePathID(c)
	if !ok {
		return
	}
	executeAdminIdempotentJSON(c, "admin.lottery.prize.delete", gin.H{"id": id}, service.DefaultWriteIdempotencyTTL(), func(ctx context.Context) (any, error) {
		if err := h.service.DeletePrize(ctx, id); err != nil {
			return nil, err
		}
		return gin.H{"message": "ok"}, nil
	})
}

func prizeInput(req lotteryPrizeRequest) service.LotteryPrizeInput {
	return service.LotteryPrizeInput{
		PoolID: req.PoolID, Name: req.Name, Description: req.Description, ImageData: req.ImageData,
		PrizeType: req.PrizeType, BalanceAmount: req.BalanceAmount, GroupID: req.GroupID,
		ValidityDays: req.ValidityDays, ProbabilityPPM: req.ProbabilityPPM, StockTotal: req.StockTotal,
		Enabled: req.Enabled, SortOrder: req.SortOrder,
	}
}

func (h *LotteryHandler) ListRules(c *gin.Context) {
	items, err := h.service.ListRules(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, items)
}

func (h *LotteryHandler) CreateRule(c *gin.Context) {
	var req lotteryRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorFrom(c, service.ErrLotteryInvalidInput)
		return
	}
	executeAdminIdempotentJSON(c, "admin.lottery.rule.create", req, service.DefaultWriteIdempotencyTTL(), func(ctx context.Context) (any, error) {
		return h.service.CreateRule(ctx, ruleInput(req))
	})
}

func (h *LotteryHandler) UpdateRule(c *gin.Context) {
	id, ok := positivePathID(c)
	if !ok {
		return
	}
	var req lotteryRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorFrom(c, service.ErrLotteryInvalidInput)
		return
	}
	executeAdminIdempotentJSON(c, "admin.lottery.rule.update", req, service.DefaultWriteIdempotencyTTL(), func(ctx context.Context) (any, error) {
		return h.service.UpdateRule(ctx, id, ruleInput(req))
	})
}

func (h *LotteryHandler) DeleteRule(c *gin.Context) {
	id, ok := positivePathID(c)
	if !ok {
		return
	}
	executeAdminIdempotentJSON(c, "admin.lottery.rule.delete", gin.H{"id": id}, service.DefaultWriteIdempotencyTTL(), func(ctx context.Context) (any, error) {
		if err := h.service.DeleteRule(ctx, id); err != nil {
			return nil, err
		}
		return gin.H{"message": "ok"}, nil
	})
}

func ruleInput(req lotteryRuleRequest) service.LotteryRuleInput {
	return service.LotteryRuleInput{
		Name: req.Name, EventType: req.EventType, Beneficiary: req.Beneficiary,
		NormalChances: req.NormalChances, LuxuryChances: req.LuxuryChances,
		RechargeMode: req.RechargeMode, RechargeThreshold: req.RechargeThreshold,
		Repeatable: req.Repeatable, Enabled: req.Enabled,
	}
}

func (h *LotteryHandler) ListDraws(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	var userID *int64
	if raw := strings.TrimSpace(c.Query("user_id")); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || value <= 0 {
			response.ErrorFrom(c, service.ErrLotteryInvalidInput)
			return
		}
		userID = &value
	}
	items, result, err := h.service.ListAdminDraws(c.Request.Context(), pagination.PaginationParams{Page: page, PageSize: pageSize}, userID, strings.TrimSpace(c.Query("pool")), strings.TrimSpace(c.Query("outcome")))
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, items, result.Total, page, pageSize)
}

func (h *LotteryHandler) ListLedger(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	var userID *int64
	if raw := strings.TrimSpace(c.Query("user_id")); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || value <= 0 {
			response.ErrorFrom(c, service.ErrLotteryInvalidInput)
			return
		}
		userID = &value
	}
	items, result, err := h.service.ListChanceLedger(c.Request.Context(), pagination.PaginationParams{Page: page, PageSize: pageSize}, userID, strings.TrimSpace(c.Query("pool")), strings.TrimSpace(c.Query("action")))
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, items, result.Total, page, pageSize)
}

func positivePathID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.ErrorFrom(c, service.ErrLotteryInvalidInput)
		return 0, false
	}
	return id, true
}

func positiveInt64Query(c *gin.Context, key string) (int64, bool) {
	value, err := strconv.ParseInt(strings.TrimSpace(c.Query(key)), 10, 64)
	return value, err == nil && value > 0
}
