// Package service 中的抽奖机会服务，将邀请、兑换和充值事件转换为幂等次数流水。
package service

import (
	"context"
	"fmt"
	"math"
	"strconv"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

type LotteryChanceService struct {
	repo LotteryRepository
}

func NewLotteryChanceService(repo LotteryRepository) *LotteryChanceService {
	return &LotteryChanceService{repo: repo}
}

func (s *LotteryChanceService) GrantSignup(ctx context.Context, inviterID, inviteeID int64) error {
	if s == nil || s.repo == nil || inviterID <= 0 || inviteeID <= 0 || inviterID == inviteeID {
		return nil
	}
	rules, err := s.repo.ListRules(ctx, LotteryEventSignup, false)
	if err != nil {
		return err
	}
	for _, rule := range rules {
		userID := inviterID
		if rule.Beneficiary == LotteryBeneficiaryInvitee {
			userID = inviteeID
		}
		if err := s.grantRule(ctx, rule, userID, inviteeID, LotteryEventSignup, strconv.FormatInt(inviteeID, 10), 0); err != nil {
			return err
		}
	}
	return nil
}

func (s *LotteryChanceService) GrantFirstRedeem(ctx context.Context, inviteeID, codeID int64, systemGenerated bool) error {
	if s == nil || s.repo == nil || inviteeID <= 0 || codeID <= 0 || systemGenerated {
		return nil
	}
	hasPrior, err := s.repo.HasPriorRedeem(ctx, inviteeID, codeID)
	if err != nil || hasPrior {
		return err
	}
	inviterID, err := s.repo.GetInviterID(ctx, inviteeID)
	if err != nil || inviterID == nil || *inviterID <= 0 {
		return err
	}
	rules, err := s.repo.ListRules(ctx, LotteryEventRedeem, false)
	if err != nil {
		return err
	}
	sourceID := strconv.FormatInt(codeID, 10)
	dedupeSourceID := "first:" + strconv.FormatInt(inviteeID, 10)
	for _, rule := range rules {
		if err := s.grantRuleWithDedupeSource(ctx, rule, *inviterID, inviteeID, LotteryEventRedeem, sourceID, dedupeSourceID, 0); err != nil {
			return fmt.Errorf("grant first redeem chance for code %d: %w", codeID, err)
		}
	}
	return nil
}

// GrantRedeemLotteryChance 将兑换码面值次数发放到指定奖池（写入 extra，rule_id 为空）。
func (s *LotteryChanceService) GrantRedeemLotteryChance(ctx context.Context, userID, codeID int64, poolKey string, chances int64) error {
	if s == nil || s.repo == nil {
		return fmt.Errorf("lottery chance service not configured")
	}
	if userID <= 0 || codeID <= 0 || chances <= 0 {
		return infraerrors.BadRequest("REDEEM_CODE_INVALID", "invalid lottery chance redeem code")
	}
	if poolKey != LotteryPoolNormal && poolKey != LotteryPoolLuxury {
		return infraerrors.BadRequest("REDEEM_CODE_INVALID", "invalid lottery chance pool_key")
	}
	sourceID := strconv.FormatInt(codeID, 10)
	dedupe := fmt.Sprintf("redeem_code:%d:%s", codeID, poolKey)
	_, err := s.repo.GrantExtraChance(ctx, LotteryChanceGrant{
		UserID:     userID,
		PoolKey:    poolKey,
		Chances:    chances,
		RuleID:     0,
		SourceType: "redeem_code",
		SourceID:   sourceID,
		DedupeKey:  dedupe,
		Metadata: map[string]any{
			"redeem_code_id": codeID,
			"pool_key":       poolKey,
		},
	})
	if err != nil {
		return fmt.Errorf("grant redeem lottery chance for code %d: %w", codeID, err)
	}
	return nil
}

func (s *LotteryChanceService) GrantRecharge(ctx context.Context, inviteeID, orderID int64, amount float64) error {
	if s == nil || s.repo == nil || inviteeID <= 0 || orderID <= 0 || amount <= 0 {
		return nil
	}
	inviterID, err := s.repo.GetInviterID(ctx, inviteeID)
	if err != nil || inviterID == nil || *inviterID <= 0 {
		return err
	}
	previousTotal, err := s.repo.NetCompletedRecharge(ctx, inviteeID, orderID)
	if err != nil {
		return err
	}
	rules, err := s.repo.ListRules(ctx, LotteryEventRecharge, false)
	if err != nil {
		return err
	}
	for _, rule := range rules {
		if rule.RechargeMode == nil || rule.RechargeThreshold == nil || *rule.RechargeThreshold <= 0 {
			continue
		}
		switch *rule.RechargeMode {
		case LotteryRechargeSingle:
			if amount+1e-9 < *rule.RechargeThreshold {
				continue
			}
			sourceID := strconv.FormatInt(orderID, 10)
			dedupeSourceID := sourceID
			if !rule.Repeatable {
				dedupeSourceID = "first:" + strconv.FormatInt(inviteeID, 10)
			}
			if err := s.grantRuleWithDedupeSource(ctx, rule, *inviterID, inviteeID, "recharge_single", sourceID, dedupeSourceID, 1); err != nil {
				return err
			}
		case LotteryRechargeCumulative:
			previousTier := int(math.Floor((previousTotal + 1e-9) / *rule.RechargeThreshold))
			currentTier := int(math.Floor((previousTotal + amount + 1e-9) / *rule.RechargeThreshold))
			if !rule.Repeatable {
				if previousTier > 0 {
					continue
				}
				if currentTier > 1 {
					currentTier = 1
				}
			}
			for tier := previousTier + 1; tier <= currentTier; tier++ {
				sourceID := strconv.FormatInt(orderID, 10)
				dedupeSourceID := "cumulative:" + strconv.FormatInt(inviteeID, 10)
				if err := s.grantRuleWithDedupeSource(ctx, rule, *inviterID, inviteeID, "recharge_cumulative", sourceID, dedupeSourceID, tier); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (s *LotteryChanceService) ReverseRecharge(ctx context.Context, inviteeID, orderID int64, originalAmount, refundAmount float64) error {
	if s == nil || s.repo == nil || inviteeID <= 0 || orderID <= 0 || refundAmount <= 0 {
		return nil
	}
	inviterID, err := s.repo.GetInviterID(ctx, inviteeID)
	if err != nil || inviterID == nil || *inviterID <= 0 {
		return err
	}
	otherTotal, err := s.repo.NetCompletedRecharge(ctx, inviteeID, orderID)
	if err != nil {
		return err
	}
	remainingOrderAmount := math.Max(0, originalAmount-refundAmount)
	netTotal := otherTotal + remainingOrderAmount
	rules, err := s.repo.ListRules(ctx, LotteryEventRecharge, true)
	if err != nil {
		return err
	}
	for _, rule := range rules {
		if rule.RechargeMode == nil || rule.RechargeThreshold == nil || *rule.RechargeThreshold <= 0 {
			continue
		}
		if *rule.RechargeMode == LotteryRechargeSingle {
			if remainingOrderAmount+1e-9 >= *rule.RechargeThreshold {
				continue
			}
			sourceID := strconv.FormatInt(orderID, 10)
			dedupeSourceID := sourceID
			if !rule.Repeatable {
				dedupeSourceID = "first:" + strconv.FormatInt(inviteeID, 10)
			}
			if err := s.reverseRuleGrant(ctx, rule, *inviterID, inviteeID, "recharge_single", sourceID, dedupeSourceID, 1, orderID); err != nil {
				return err
			}
			continue
		}

		allowedTier := int(math.Floor((netTotal + 1e-9) / *rule.RechargeThreshold))
		if !rule.Repeatable && allowedTier > 1 {
			allowedTier = 1
		}
		keys, err := s.repo.ListActiveCumulativeGrantKeys(ctx, *inviterID, rule.ID, inviteeID, allowedTier)
		if err != nil {
			return err
		}
		for _, key := range keys {
			if _, _, err := s.repo.ReverseExtraChance(ctx, key, "reverse:"+key, map[string]any{"refund_order_id": orderID}); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *LotteryChanceService) grantRule(ctx context.Context, rule LotteryRule, userID, sourceUserID int64, sourceType, sourceID string, tier int) error {
	return s.grantRuleWithDedupeSource(ctx, rule, userID, sourceUserID, sourceType, sourceID, sourceID, tier)
}

func (s *LotteryChanceService) grantRuleWithDedupeSource(ctx context.Context, rule LotteryRule, userID, sourceUserID int64, sourceType, sourceID, dedupeSourceID string, tier int) error {
	targets := []struct {
		poolKey string
		chances int
	}{
		{LotteryPoolNormal, rule.NormalChances},
		{LotteryPoolLuxury, rule.LuxuryChances},
	}
	for _, target := range targets {
		if target.chances <= 0 {
			continue
		}
		dedupe := lotteryGrantDedupeKey(rule.ID, sourceType, dedupeSourceID, userID, target.poolKey, tier)
		_, err := s.repo.GrantExtraChance(ctx, LotteryChanceGrant{
			UserID:       userID,
			PoolKey:      target.poolKey,
			Chances:      int64(target.chances),
			RuleID:       rule.ID,
			SourceType:   sourceType,
			SourceID:     sourceID,
			SourceUserID: sourceUserID,
			TierNo:       tier,
			DedupeKey:    dedupe,
			Metadata: map[string]any{
				"rule_name": rule.Name, "event_type": rule.EventType,
				"recharge_mode": rule.RechargeMode, "recharge_threshold": rule.RechargeThreshold,
				"repeatable": rule.Repeatable,
			},
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *LotteryChanceService) reverseRuleGrant(ctx context.Context, rule LotteryRule, userID, sourceUserID int64, sourceType, sourceID, dedupeSourceID string, tier int, refundOrderID int64) error {
	for _, target := range []struct {
		poolKey string
		chances int
	}{{LotteryPoolNormal, rule.NormalChances}, {LotteryPoolLuxury, rule.LuxuryChances}} {
		if target.chances <= 0 {
			continue
		}
		key := lotteryGrantDedupeKey(rule.ID, sourceType, dedupeSourceID, userID, target.poolKey, tier)
		matches, err := s.repo.GrantMatchesSource(ctx, key, sourceType, sourceID)
		if err != nil {
			return err
		}
		if !matches {
			continue
		}
		if _, _, err := s.repo.ReverseExtraChance(ctx, key, "reverse:"+key, map[string]any{
			"refund_order_id": refundOrderID,
			"source_user_id":  sourceUserID,
		}); err != nil {
			return err
		}
	}
	return nil
}
