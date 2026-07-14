// Package service 中的抽奖服务负责用户查询、原子抽奖、自动发奖和管理配置。
package service

import (
	"context"
	cryptorand "crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

type LotteryService struct {
	repo                 LotteryRepository
	redeemService        *RedeemService
	groupRepo            GroupRepository
	entClient            *dbent.Client
	authCacheInvalidator APIKeyAuthCacheInvalidator
	billingCacheService  *BillingCacheService
}

func NewLotteryService(
	repo LotteryRepository,
	redeemService *RedeemService,
	groupRepo GroupRepository,
	entClient *dbent.Client,
	authCacheInvalidator APIKeyAuthCacheInvalidator,
	billingCacheService *BillingCacheService,
) *LotteryService {
	return &LotteryService{
		repo: repo, redeemService: redeemService, groupRepo: groupRepo, entClient: entClient,
		authCacheInvalidator: authCacheInvalidator, billingCacheService: billingCacheService,
	}
}

func (s *LotteryService) Summary(ctx context.Context, userID int64) (*LotterySummary, error) {
	if userID <= 0 {
		return nil, ErrLotteryInvalidInput
	}
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	txCtx := dbent.NewTxContext(ctx, tx)
	pools, err := s.repo.ListPools(txCtx)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	result := &LotterySummary{Pools: make([]LotteryPoolSummary, 0, len(pools))}
	for _, pool := range pools {
		account, err := s.repo.LockChanceAccount(txCtx, userID, pool, lotteryPeriodKey(now, pool.CycleType))
		if err != nil {
			return nil, err
		}
		prizes, err := s.repo.ListPrizes(txCtx, pool.ID, false)
		if err != nil {
			return nil, err
		}
		result.Pools = append(result.Pools, LotteryPoolSummary{
			Pool: pool, Prizes: prizes, BaseRemaining: account.BaseRemaining,
			ExtraRemaining: account.ExtraRemaining, PeriodKey: account.PeriodKey, Active: pool.ActiveAt(now),
		})
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *LotteryService) Draw(ctx context.Context, userID int64, poolKey, idempotencyKey string) (*LotteryDraw, error) {
	poolKey = strings.TrimSpace(poolKey)
	idempotencyKey = strings.TrimSpace(idempotencyKey)
	if userID <= 0 || (poolKey != LotteryPoolNormal && poolKey != LotteryPoolLuxury) || idempotencyKey == "" || len(idempotencyKey) > 128 {
		return nil, ErrLotteryInvalidInput
	}
	pool, err := s.repo.GetPoolByKey(ctx, poolKey)
	if err != nil {
		return nil, err
	}
	existing, existingErr := s.repo.GetDrawByIdempotencyKey(ctx, userID, pool.ID, idempotencyKey)
	if existingErr == nil {
		return existing, nil
	}
	if !errors.Is(existingErr, ErrLotteryDrawNotFound) {
		return nil, existingErr
	}

	draw, fulfilledCode, balanceChanged, err := s.drawInTransaction(ctx, userID, poolKey, idempotencyKey)
	if errors.Is(err, ErrLotteryAlreadyExists) {
		return s.repo.GetDrawByIdempotencyKey(ctx, userID, pool.ID, idempotencyKey)
	}
	if err != nil {
		return nil, err
	}
	if fulfilledCode != nil {
		s.redeemService.FinalizeSystemRedeem(ctx, userID, fulfilledCode)
	}
	if balanceChanged {
		if s.authCacheInvalidator != nil {
			s.authCacheInvalidator.InvalidateAuthCacheByUserID(ctx, userID)
		}
		if s.billingCacheService != nil {
			_ = s.billingCacheService.InvalidateUserBalance(ctx, userID)
		}
	}
	return draw, nil
}

func (s *LotteryService) drawInTransaction(ctx context.Context, userID int64, poolKey, idempotencyKey string) (*LotteryDraw, *RedeemCode, bool, error) {
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return nil, nil, false, err
	}
	defer func() { _ = tx.Rollback() }()
	txCtx := dbent.NewTxContext(ctx, tx)

	pool, err := s.repo.GetPoolByKey(txCtx, poolKey)
	if err != nil {
		return nil, nil, false, err
	}
	now := time.Now()
	if !pool.ActiveAt(now) {
		return nil, nil, false, ErrLotteryInactive
	}
	existing, existingErr := s.repo.GetDrawByIdempotencyKey(txCtx, userID, pool.ID, idempotencyKey)
	if existingErr == nil {
		return existing, nil, false, nil
	}
	if !errors.Is(existingErr, ErrLotteryDrawNotFound) {
		return nil, nil, false, existingErr
	}
	account, err := s.repo.LockChanceAccount(txCtx, userID, *pool, lotteryPeriodKey(now, pool.CycleType))
	if err != nil {
		return nil, nil, false, err
	}
	// 相同用户和奖池的并发请求会在次数账户行锁处串行；加锁后再次读取即可复用先提交的结果。
	existing, existingErr = s.repo.GetDrawByIdempotencyKey(txCtx, userID, pool.ID, idempotencyKey)
	if existingErr == nil {
		return existing, nil, false, nil
	}
	if !errors.Is(existingErr, ErrLotteryDrawNotFound) {
		return nil, nil, false, existingErr
	}
	prizes, err := s.repo.ListPrizes(txCtx, pool.ID, false)
	if err != nil {
		return nil, nil, false, err
	}
	roll, err := secureLotteryRoll()
	if err != nil {
		return nil, nil, false, err
	}
	selected := selectLotteryPrize(prizes, roll)
	chanceSource, account, err := s.repo.ConsumeChance(txCtx, *account, pool.ID, idempotencyKey)
	if err != nil {
		return nil, nil, false, err
	}

	outcome := "none"
	var prizeID *int64
	var redeemCodeID *int64
	var fulfilledCode *RedeemCode
	balanceChanged := false
	snapshot := map[string]any{}
	if selected != nil {
		locked, lockErr := s.repo.GetPrize(txCtx, selected.ID, true)
		if lockErr != nil && !errors.Is(lockErr, ErrLotteryPrizeNotFound) {
			return nil, nil, false, lockErr
		}
		if lockErr == nil && locked.Enabled && locked.InStock() {
			claimed, claimErr := s.repo.ClaimPrizeStock(txCtx, locked.ID)
			if claimErr != nil {
				return nil, nil, false, claimErr
			}
			if claimed {
				outcome = "win"
				prizeID = &locked.ID
				snapshot = lotteryPrizeSnapshot(*locked)
				switch locked.PrizeType {
				case LotteryPrizeBalance:
					if locked.BalanceAmount == nil || *locked.BalanceAmount <= 0 {
						return nil, nil, false, ErrLotteryFulfillFailed
					}
					if err := s.repo.CreditBalance(txCtx, userID, *locked.BalanceAmount); err != nil {
						return nil, nil, false, fmt.Errorf("credit lottery balance: %w", err)
					}
					balanceChanged = true
				case LotteryPrizeSubscription:
					if locked.GroupID == nil || locked.ValidityDays == nil || s.redeemService == nil {
						return nil, nil, false, ErrLotteryFulfillFailed
					}
					fulfilledCode, err = s.redeemService.CreateAndRedeemSystemSubscriptionInTx(
						txCtx, userID, *locked.GroupID, *locked.ValidityDays,
						fmt.Sprintf("抽奖奖品，奖池=%s，幂等键=%s", pool.Key, idempotencyKey),
					)
					if err != nil {
						return nil, nil, false, fmt.Errorf("redeem lottery subscription: %w", err)
					}
					redeemCodeID = &fulfilledCode.ID
				default:
					return nil, nil, false, ErrLotteryFulfillFailed
				}
			}
		}
	}

	draw, err := s.repo.CreateDraw(txCtx, userID, pool.ID, idempotencyKey, outcome, chanceSource, prizeID, redeemCodeID, roll, snapshot)
	if err != nil {
		return nil, nil, false, err
	}
	draw.BaseRemaining = account.BaseRemaining
	draw.ExtraRemaining = account.ExtraRemaining
	if err := tx.Commit(); err != nil {
		return nil, nil, false, err
	}
	return draw, fulfilledCode, balanceChanged, nil
}

func secureLotteryRoll() (int, error) {
	value, err := cryptorand.Int(cryptorand.Reader, big.NewInt(LotteryProbabilityScale))
	if err != nil {
		return 0, fmt.Errorf("generate lottery roll: %w", err)
	}
	return int(value.Int64()), nil
}

func selectLotteryPrize(prizes []LotteryPrize, roll int) *LotteryPrize {
	cursor := 0
	for i := range prizes {
		prize := &prizes[i]
		if !prize.Enabled || prize.ProbabilityPPM <= 0 {
			continue
		}
		cursor += prize.ProbabilityPPM
		if roll < cursor {
			if !prize.InStock() {
				return nil
			}
			return prize
		}
	}
	return nil
}

func lotteryPrizeSnapshot(prize LotteryPrize) map[string]any {
	return map[string]any{
		"id": prize.ID, "name": prize.Name, "description": prize.Description,
		"prize_type": prize.PrizeType, "balance_amount": prize.BalanceAmount,
		"group_id": prize.GroupID, "validity_days": prize.ValidityDays,
		"probability_ppm": prize.ProbabilityPPM,
	}
}

func (s *LotteryService) ListUserDraws(ctx context.Context, userID int64, params pagination.PaginationParams, poolKey string) ([]LotteryDraw, *pagination.PaginationResult, error) {
	return s.repo.ListUserDraws(ctx, userID, params, poolKey)
}

func (s *LotteryService) ListPools(ctx context.Context) ([]LotteryPool, error) {
	return s.repo.ListPools(ctx)
}

func (s *LotteryService) UpdatePool(ctx context.Context, key string, input LotteryPoolUpdate) (*LotteryPool, error) {
	if key != LotteryPoolNormal && key != LotteryPoolLuxury {
		return nil, ErrLotteryPoolNotFound
	}
	if err := validateLotteryPoolUpdate(input); err != nil {
		return nil, err
	}
	return s.repo.UpdatePool(ctx, key, input)
}

func (s *LotteryService) ListPrizes(ctx context.Context, poolID int64) ([]LotteryPrize, error) {
	return s.repo.ListPrizes(ctx, poolID, true)
}

func (s *LotteryService) CreatePrize(ctx context.Context, input LotteryPrizeInput) (*LotteryPrize, error) {
	if err := s.validatePrize(ctx, input, 0); err != nil {
		return nil, err
	}
	return s.repo.CreatePrize(ctx, input)
}

func (s *LotteryService) UpdatePrize(ctx context.Context, id int64, input LotteryPrizeInput) (*LotteryPrize, error) {
	if id <= 0 {
		return nil, ErrLotteryPrizeNotFound
	}
	if err := s.validatePrize(ctx, input, id); err != nil {
		return nil, err
	}
	return s.repo.UpdatePrize(ctx, id, input)
}

func (s *LotteryService) validatePrize(ctx context.Context, input LotteryPrizeInput, excludeID int64) error {
	if err := validateLotteryPrizeInput(input); err != nil {
		return err
	}
	pools, err := s.repo.ListPools(ctx)
	if err != nil {
		return err
	}
	found := false
	for _, pool := range pools {
		if pool.ID == input.PoolID {
			found = true
			break
		}
	}
	if !found {
		return ErrLotteryPoolNotFound
	}
	if input.PrizeType == LotteryPrizeSubscription {
		group, err := s.groupRepo.GetByID(ctx, *input.GroupID)
		if err != nil || group == nil || !group.IsSubscriptionType() {
			return ErrGroupNotSubscriptionType
		}
	}
	if input.Enabled {
		total, err := s.repo.EnabledProbabilityTotal(ctx, input.PoolID, excludeID)
		if err != nil {
			return err
		}
		if total+input.ProbabilityPPM > LotteryProbabilityScale {
			return ErrLotteryProbability
		}
	}
	return nil
}

func (s *LotteryService) DeletePrize(ctx context.Context, id int64) error {
	return s.repo.DisablePrize(ctx, id)
}

func (s *LotteryService) ListRules(ctx context.Context) ([]LotteryRule, error) {
	return s.repo.ListRules(ctx, "", true)
}

func (s *LotteryService) CreateRule(ctx context.Context, input LotteryRuleInput) (*LotteryRule, error) {
	if err := validateLotteryRuleInput(input); err != nil {
		return nil, err
	}
	return s.repo.CreateRule(ctx, input)
}

func (s *LotteryService) UpdateRule(ctx context.Context, id int64, input LotteryRuleInput) (*LotteryRule, error) {
	if id <= 0 {
		return nil, ErrLotteryRuleNotFound
	}
	if err := validateLotteryRuleInput(input); err != nil {
		return nil, err
	}
	existing, err := s.repo.GetRule(ctx, id)
	if err != nil {
		return nil, err
	}
	hasLedger, err := s.repo.RuleHasLedger(ctx, id)
	if err != nil {
		return nil, err
	}
	if hasLedger && !lotteryRuleBehaviorEqual(*existing, input) {
		return nil, ErrLotteryRuleImmutable
	}
	return s.repo.UpdateRule(ctx, id, input)
}

func lotteryRuleBehaviorEqual(rule LotteryRule, input LotteryRuleInput) bool {
	if rule.EventType != input.EventType || rule.Beneficiary != input.Beneficiary ||
		rule.NormalChances != input.NormalChances || rule.LuxuryChances != input.LuxuryChances ||
		rule.Repeatable != input.Repeatable {
		return false
	}
	if (rule.RechargeMode == nil) != (input.RechargeMode == nil) || (rule.RechargeThreshold == nil) != (input.RechargeThreshold == nil) {
		return false
	}
	if rule.RechargeMode != nil && *rule.RechargeMode != *input.RechargeMode {
		return false
	}
	return rule.RechargeThreshold == nil || *rule.RechargeThreshold == *input.RechargeThreshold
}

func (s *LotteryService) DeleteRule(ctx context.Context, id int64) error {
	return s.repo.DisableRule(ctx, id)
}

func (s *LotteryService) ListAdminDraws(ctx context.Context, params pagination.PaginationParams, userID *int64, poolKey, outcome string) ([]LotteryDraw, *pagination.PaginationResult, error) {
	return s.repo.ListDraws(ctx, params, userID, poolKey, outcome)
}

func (s *LotteryService) ListChanceLedger(ctx context.Context, params pagination.PaginationParams, userID *int64, poolKey, action string) ([]LotteryChanceLedgerEntry, *pagination.PaginationResult, error) {
	return s.repo.ListChanceLedger(ctx, params, userID, poolKey, action)
}
