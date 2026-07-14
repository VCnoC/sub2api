// Package service 中的抽奖领域契约，定义双奖池、规则、次数和审计数据。
package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

const (
	LotteryPoolNormal = "normal"
	LotteryPoolLuxury = "luxury"

	LotteryCycleDaily  = "daily"
	LotteryCycleWeekly = "weekly"

	LotteryPrizeBalance      = "balance"
	LotteryPrizeSubscription = "subscription"

	LotteryEventSignup   = "signup"
	LotteryEventRedeem   = "redeem"
	LotteryEventRecharge = "recharge"

	LotteryBeneficiaryInviter = "inviter"
	LotteryBeneficiaryInvitee = "invitee"

	LotteryRechargeSingle     = "single"
	LotteryRechargeCumulative = "cumulative"

	LotteryProbabilityScale       = 1_000_000
	LotterySystemRedeemNotePrefix = "[lottery] "
	lotteryImageMaxBytes          = 300 * 1024
	lotteryMaxRuleChances         = 100_000
	lotteryMaxMoney               = 1_000_000_000_000
	lotteryMaxDescription         = 2_000
)

var (
	ErrLotteryPoolNotFound  = infraerrors.NotFound("LOTTERY_POOL_NOT_FOUND", "lottery pool not found")
	ErrLotteryPrizeNotFound = infraerrors.NotFound("LOTTERY_PRIZE_NOT_FOUND", "lottery prize not found")
	ErrLotteryRuleNotFound  = infraerrors.NotFound("LOTTERY_RULE_NOT_FOUND", "lottery rule not found")
	ErrLotteryDrawNotFound  = infraerrors.NotFound("LOTTERY_DRAW_NOT_FOUND", "lottery draw not found")
	ErrLotteryInactive      = infraerrors.Forbidden("LOTTERY_INACTIVE", "lottery pool is not active")
	ErrLotteryNoChance      = infraerrors.BadRequest("LOTTERY_NO_CHANCE", "no lottery chance available")
	ErrLotteryInvalidInput  = infraerrors.BadRequest("LOTTERY_INVALID_INPUT", "invalid lottery configuration")
	ErrLotteryProbability   = infraerrors.BadRequest("LOTTERY_PROBABILITY_INVALID", "enabled prize probability exceeds 100 percent")
	ErrLotteryInvalidImage  = infraerrors.BadRequest("LOTTERY_IMAGE_INVALID", "invalid lottery prize image")
	ErrLotteryFulfillFailed = infraerrors.InternalServer("LOTTERY_FULFILL_FAILED", "failed to fulfill lottery prize")
	ErrLotteryAlreadyExists = infraerrors.Conflict("LOTTERY_ALREADY_EXISTS", "lottery record already exists")
	ErrLotteryRuleImmutable = infraerrors.Conflict("LOTTERY_RULE_IMMUTABLE", "lottery rule behavior cannot change after rewards have been issued")
)

type LotteryPool struct {
	ID           int64      `json:"id"`
	Key          string     `json:"key"`
	Name         string     `json:"name"`
	Enabled      bool       `json:"enabled"`
	CycleType    string     `json:"cycle_type"`
	CycleChances int        `json:"cycle_chances"`
	StartsAt     *time.Time `json:"starts_at,omitempty"`
	EndsAt       *time.Time `json:"ends_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

func (p LotteryPool) ActiveAt(now time.Time) bool {
	if !p.Enabled {
		return false
	}
	if p.StartsAt != nil && now.Before(*p.StartsAt) {
		return false
	}
	return p.EndsAt == nil || now.Before(*p.EndsAt)
}

type LotteryPrize struct {
	ID             int64     `json:"id"`
	PoolID         int64     `json:"pool_id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	ImageData      string    `json:"image_data,omitempty"`
	PrizeType      string    `json:"prize_type"`
	BalanceAmount  *float64  `json:"balance_amount,omitempty"`
	GroupID        *int64    `json:"group_id,omitempty"`
	ValidityDays   *int      `json:"validity_days,omitempty"`
	ProbabilityPPM int       `json:"probability_ppm"`
	StockTotal     *int64    `json:"stock_total,omitempty"`
	StockUsed      int64     `json:"stock_used"`
	Enabled        bool      `json:"enabled"`
	SortOrder      int       `json:"sort_order"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (p LotteryPrize) InStock() bool {
	return p.StockTotal == nil || p.StockUsed < *p.StockTotal
}

type LotteryRule struct {
	ID                int64     `json:"id"`
	Name              string    `json:"name"`
	EventType         string    `json:"event_type"`
	Beneficiary       string    `json:"beneficiary"`
	NormalChances     int       `json:"normal_chances"`
	LuxuryChances     int       `json:"luxury_chances"`
	RechargeMode      *string   `json:"recharge_mode,omitempty"`
	RechargeThreshold *float64  `json:"recharge_threshold,omitempty"`
	Repeatable        bool      `json:"repeatable"`
	Enabled           bool      `json:"enabled"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type LotteryChanceAccount struct {
	UserID         int64     `json:"user_id"`
	PoolID         int64     `json:"pool_id"`
	PeriodKey      string    `json:"period_key"`
	BaseRemaining  int       `json:"base_remaining"`
	ExtraRemaining int64     `json:"extra_remaining"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type LotteryDraw struct {
	ID             int64          `json:"id"`
	UserID         int64          `json:"user_id,omitempty"`
	PoolID         int64          `json:"pool_id"`
	PoolKey        string         `json:"pool_key,omitempty"`
	Outcome        string         `json:"outcome"`
	ChanceSource   string         `json:"chance_source"`
	PrizeID        *int64         `json:"prize_id,omitempty"`
	RedeemCodeID   *int64         `json:"redeem_code_id,omitempty"`
	RandomRoll     int            `json:"-"`
	PrizeSnapshot  map[string]any `json:"prize,omitempty"`
	BaseRemaining  int            `json:"base_remaining,omitempty"`
	ExtraRemaining int64          `json:"extra_remaining,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
}

type LotteryChanceLedgerEntry struct {
	ID           int64          `json:"id"`
	UserID       int64          `json:"user_id"`
	PoolID       int64          `json:"pool_id"`
	PoolKey      string         `json:"pool_key,omitempty"`
	Action       string         `json:"action"`
	BaseDelta    int            `json:"base_delta"`
	ExtraDelta   int64          `json:"extra_delta"`
	RuleID       *int64         `json:"rule_id,omitempty"`
	SourceType   string         `json:"source_type"`
	SourceID     string         `json:"source_id"`
	SourceUserID *int64         `json:"source_user_id,omitempty"`
	TierNo       int            `json:"tier_no"`
	Metadata     map[string]any `json:"metadata,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
}

type LotteryPoolSummary struct {
	Pool           LotteryPool    `json:"pool"`
	Prizes         []LotteryPrize `json:"prizes"`
	BaseRemaining  int            `json:"base_remaining"`
	ExtraRemaining int64          `json:"extra_remaining"`
	PeriodKey      string         `json:"period_key"`
	Active         bool           `json:"active"`
}

type LotterySummary struct {
	Pools []LotteryPoolSummary `json:"pools"`
}

type LotteryPoolUpdate struct {
	Name         string
	Enabled      bool
	CycleType    string
	CycleChances int
	StartsAt     *time.Time
	EndsAt       *time.Time
}

type LotteryPrizeInput struct {
	PoolID         int64
	Name           string
	Description    string
	ImageData      string
	PrizeType      string
	BalanceAmount  *float64
	GroupID        *int64
	ValidityDays   *int
	ProbabilityPPM int
	StockTotal     *int64
	Enabled        bool
	SortOrder      int
}

type LotteryRuleInput struct {
	Name              string
	EventType         string
	Beneficiary       string
	NormalChances     int
	LuxuryChances     int
	RechargeMode      *string
	RechargeThreshold *float64
	Repeatable        bool
	Enabled           bool
}

type LotteryChanceGrant struct {
	UserID       int64
	PoolKey      string
	Chances      int64
	RuleID       int64
	SourceType   string
	SourceID     string
	SourceUserID int64
	TierNo       int
	DedupeKey    string
	Metadata     map[string]any
}

type LotteryRepository interface {
	ListPools(ctx context.Context) ([]LotteryPool, error)
	GetPoolByKey(ctx context.Context, key string) (*LotteryPool, error)
	UpdatePool(ctx context.Context, key string, input LotteryPoolUpdate) (*LotteryPool, error)
	ListPrizes(ctx context.Context, poolID int64, includeDisabled bool) ([]LotteryPrize, error)
	GetPrize(ctx context.Context, id int64, forUpdate bool) (*LotteryPrize, error)
	CreatePrize(ctx context.Context, input LotteryPrizeInput) (*LotteryPrize, error)
	UpdatePrize(ctx context.Context, id int64, input LotteryPrizeInput) (*LotteryPrize, error)
	DisablePrize(ctx context.Context, id int64) error
	EnabledProbabilityTotal(ctx context.Context, poolID, excludePrizeID int64) (int, error)
	ClaimPrizeStock(ctx context.Context, prizeID int64) (bool, error)
	CreditBalance(ctx context.Context, userID int64, amount float64) error

	ListRules(ctx context.Context, eventType string, includeDisabled bool) ([]LotteryRule, error)
	GetRule(ctx context.Context, id int64) (*LotteryRule, error)
	CreateRule(ctx context.Context, input LotteryRuleInput) (*LotteryRule, error)
	UpdateRule(ctx context.Context, id int64, input LotteryRuleInput) (*LotteryRule, error)
	DisableRule(ctx context.Context, id int64) error

	LockChanceAccount(ctx context.Context, userID int64, pool LotteryPool, periodKey string) (*LotteryChanceAccount, error)
	GrantExtraChance(ctx context.Context, input LotteryChanceGrant) (bool, error)
	ReverseExtraChance(ctx context.Context, grantDedupeKey, reversalDedupeKey string, metadata map[string]any) (int64, bool, error)
	ConsumeChance(ctx context.Context, account LotteryChanceAccount, poolID int64, drawDedupeKey string) (string, *LotteryChanceAccount, error)
	CreateDraw(ctx context.Context, userID int64, poolID int64, idempotencyKey string, outcome, chanceSource string, prizeID, redeemCodeID *int64, randomRoll int, snapshot map[string]any) (*LotteryDraw, error)
	GetDrawByIdempotencyKey(ctx context.Context, userID, poolID int64, key string) (*LotteryDraw, error)
	ListUserDraws(ctx context.Context, userID int64, params pagination.PaginationParams, poolKey string) ([]LotteryDraw, *pagination.PaginationResult, error)
	ListDraws(ctx context.Context, params pagination.PaginationParams, userID *int64, poolKey, outcome string) ([]LotteryDraw, *pagination.PaginationResult, error)
	ListChanceLedger(ctx context.Context, params pagination.PaginationParams, userID *int64, poolKey, action string) ([]LotteryChanceLedgerEntry, *pagination.PaginationResult, error)

	GetInviterID(ctx context.Context, userID int64) (*int64, error)
	HasPriorRedeem(ctx context.Context, userID, excludingCodeID int64) (bool, error)
	NetCompletedRecharge(ctx context.Context, userID int64, excludingOrderID int64) (float64, error)
	GrantMatchesSource(ctx context.Context, dedupeKey, sourceType, sourceID string) (bool, error)
	ListActiveCumulativeGrantKeys(ctx context.Context, userID, ruleID, sourceUserID int64, aboveTier int) ([]string, error)
	RuleHasLedger(ctx context.Context, ruleID int64) (bool, error)
}

func lotteryPeriodKey(now time.Time, cycleType string) string {
	local := now.In(time.Local)
	if cycleType == LotteryCycleWeekly {
		weekday := (int(local.Weekday()) + 6) % 7
		monday := local.AddDate(0, 0, -weekday)
		return "w:" + monday.Format("2006-01-02")
	}
	return "d:" + local.Format("2006-01-02")
}

func validateLotteryPoolUpdate(input LotteryPoolUpdate) error {
	name := strings.TrimSpace(input.Name)
	if name == "" || utf8.RuneCountInString(name) > 80 || (input.CycleType != LotteryCycleDaily && input.CycleType != LotteryCycleWeekly) || input.CycleChances < 0 || input.CycleChances > 100 {
		return ErrLotteryInvalidInput
	}
	if input.StartsAt != nil && input.EndsAt != nil && !input.EndsAt.After(*input.StartsAt) {
		return ErrLotteryInvalidInput
	}
	return nil
}

func validateLotteryPrizeInput(input LotteryPrizeInput) error {
	name := strings.TrimSpace(input.Name)
	if input.PoolID <= 0 || name == "" || utf8.RuneCountInString(name) > 120 || utf8.RuneCountInString(strings.TrimSpace(input.Description)) > lotteryMaxDescription || input.ProbabilityPPM < 0 || input.ProbabilityPPM > LotteryProbabilityScale {
		return ErrLotteryInvalidInput
	}
	if input.StockTotal != nil && *input.StockTotal < 0 {
		return ErrLotteryInvalidInput
	}
	switch input.PrizeType {
	case LotteryPrizeBalance:
		if input.BalanceAmount == nil || !validLotteryMoney(*input.BalanceAmount) || input.GroupID != nil || input.ValidityDays != nil {
			return ErrLotteryInvalidInput
		}
	case LotteryPrizeSubscription:
		if input.GroupID == nil || *input.GroupID <= 0 || input.ValidityDays == nil || *input.ValidityDays <= 0 || *input.ValidityDays > MaxValidityDays || input.BalanceAmount != nil {
			return ErrLotteryInvalidInput
		}
	default:
		return ErrLotteryInvalidInput
	}
	return validateLotteryImage(input.ImageData)
}

func validateLotteryRuleInput(input LotteryRuleInput) error {
	name := strings.TrimSpace(input.Name)
	if name == "" || utf8.RuneCountInString(name) > 120 || input.NormalChances < 0 || input.NormalChances > lotteryMaxRuleChances || input.LuxuryChances < 0 || input.LuxuryChances > lotteryMaxRuleChances || input.NormalChances+input.LuxuryChances <= 0 {
		return ErrLotteryInvalidInput
	}
	if input.Beneficiary != LotteryBeneficiaryInviter && input.Beneficiary != LotteryBeneficiaryInvitee {
		return ErrLotteryInvalidInput
	}
	switch input.EventType {
	case LotteryEventSignup:
		if input.RechargeMode != nil || input.RechargeThreshold != nil || input.Repeatable {
			return ErrLotteryInvalidInput
		}
	case LotteryEventRedeem:
		if input.Beneficiary != LotteryBeneficiaryInviter || input.RechargeMode != nil || input.RechargeThreshold != nil || input.Repeatable {
			return ErrLotteryInvalidInput
		}
	case LotteryEventRecharge:
		if input.Beneficiary != LotteryBeneficiaryInviter || input.RechargeMode == nil || (*input.RechargeMode != LotteryRechargeSingle && *input.RechargeMode != LotteryRechargeCumulative) || input.RechargeThreshold == nil || !validLotteryMoney(*input.RechargeThreshold) {
			return ErrLotteryInvalidInput
		}
	default:
		return ErrLotteryInvalidInput
	}
	return nil
}

func validateLotteryImage(value string) error {
	if value == "" {
		return nil
	}
	comma := strings.IndexByte(value, ',')
	if comma <= 0 {
		return ErrLotteryInvalidImage
	}
	header := value[:comma]
	expectedType := map[string]string{
		"data:image/png;base64":  "image/png",
		"data:image/jpeg;base64": "image/jpeg",
		"data:image/webp;base64": "image/webp",
	}[header]
	if expectedType == "" {
		return ErrLotteryInvalidImage
	}
	encoded := value[comma+1:]
	if len(encoded) > base64.StdEncoding.EncodedLen(lotteryImageMaxBytes) {
		return ErrLotteryInvalidImage
	}
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil || len(decoded) == 0 || len(decoded) > lotteryImageMaxBytes || http.DetectContentType(decoded) != expectedType {
		return ErrLotteryInvalidImage
	}
	return nil
}

func validLotteryMoney(value float64) bool {
	return value > 0 && value <= lotteryMaxMoney && !math.IsNaN(value) && !math.IsInf(value, 0)
}

func lotteryGrantDedupeKey(ruleID int64, sourceType, sourceID string, userID int64, poolKey string, tier int) string {
	return fmt.Sprintf("rule:%d:%s:%s:user:%d:pool:%s:tier:%d", ruleID, sourceType, sourceID, userID, poolKey, tier)
}
