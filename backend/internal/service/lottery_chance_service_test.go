// Package service 的机会规则测试覆盖双奖池、首次事件、充值档位和退款冲正。
package service

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type lotteryChanceRepoStub struct {
	LotteryRepository
	rules         []LotteryRule
	inviterID     *int64
	hasPrior      bool
	previousTotal float64
	grants        []LotteryChanceGrant
	grantByKey    map[string]LotteryChanceGrant
	reversals     []string
	activeKeys    []string
}

func (r *lotteryChanceRepoStub) ListRules(context.Context, string, bool) ([]LotteryRule, error) {
	return append([]LotteryRule(nil), r.rules...), nil
}

func (r *lotteryChanceRepoStub) GetInviterID(context.Context, int64) (*int64, error) {
	return r.inviterID, nil
}

func (r *lotteryChanceRepoStub) HasPriorRedeem(context.Context, int64, int64) (bool, error) {
	return r.hasPrior, nil
}

func (r *lotteryChanceRepoStub) NetCompletedRecharge(context.Context, int64, int64) (float64, error) {
	return r.previousTotal, nil
}

func (r *lotteryChanceRepoStub) GrantExtraChance(_ context.Context, input LotteryChanceGrant) (bool, error) {
	if r.grantByKey == nil {
		r.grantByKey = make(map[string]LotteryChanceGrant)
	}
	if _, exists := r.grantByKey[input.DedupeKey]; exists {
		return false, nil
	}
	r.grantByKey[input.DedupeKey] = input
	r.grants = append(r.grants, input)
	return true, nil
}

func (r *lotteryChanceRepoStub) GrantMatchesSource(_ context.Context, key, sourceType, sourceID string) (bool, error) {
	grant, ok := r.grantByKey[key]
	return ok && grant.SourceType == sourceType && grant.SourceID == sourceID, nil
}

func (r *lotteryChanceRepoStub) ReverseExtraChance(_ context.Context, key, reversalKey string, _ map[string]any) (int64, bool, error) {
	if _, exists := r.grantByKey[key]; !exists {
		return 0, false, nil
	}
	r.reversals = append(r.reversals, reversalKey)
	return r.grantByKey[key].Chances, true, nil
}

func (r *lotteryChanceRepoStub) ListActiveCumulativeGrantKeys(context.Context, int64, int64, int64, int) ([]string, error) {
	return append([]string(nil), r.activeKeys...), nil
}

func TestLotteryChanceSignupRulesCanTargetBothPoolsAndUsers(t *testing.T) {
	inviterID := int64(10)
	repo := &lotteryChanceRepoStub{inviterID: &inviterID, rules: []LotteryRule{
		{ID: 1, Name: "inviter", EventType: LotteryEventSignup, Beneficiary: LotteryBeneficiaryInviter, NormalChances: 2, LuxuryChances: 1},
		{ID: 2, Name: "invitee", EventType: LotteryEventSignup, Beneficiary: LotteryBeneficiaryInvitee, NormalChances: 3},
	}}

	require.NoError(t, NewLotteryChanceService(repo).GrantSignup(context.Background(), inviterID, 20))
	require.Len(t, repo.grants, 3)
	require.EqualValues(t, []int64{10, 10, 20}, []int64{repo.grants[0].UserID, repo.grants[1].UserID, repo.grants[2].UserID})
	require.Equal(t, []string{LotteryPoolNormal, LotteryPoolLuxury, LotteryPoolNormal}, []string{repo.grants[0].PoolKey, repo.grants[1].PoolKey, repo.grants[2].PoolKey})
}

func TestLotteryChanceFirstRedeemUsesHistoryAndStableDedupe(t *testing.T) {
	inviterID := int64(10)
	rule := LotteryRule{ID: 3, Name: "redeem", EventType: LotteryEventRedeem, Beneficiary: LotteryBeneficiaryInviter, NormalChances: 1}
	repo := &lotteryChanceRepoStub{inviterID: &inviterID, hasPrior: true, rules: []LotteryRule{rule}}
	service := NewLotteryChanceService(repo)

	require.NoError(t, service.GrantFirstRedeem(context.Background(), 20, 91, false))
	require.Empty(t, repo.grants)

	repo.hasPrior = false
	require.NoError(t, service.GrantFirstRedeem(context.Background(), 20, 91, false))
	require.Len(t, repo.grants, 1)
	require.Equal(t, "91", repo.grants[0].SourceID)
	require.Contains(t, repo.grants[0].DedupeKey, "first:20")
}

func TestLotteryChanceRechargeUsesOrderAuditAndLifetimeTierDedupe(t *testing.T) {
	inviterID := int64(10)
	singleMode, cumulativeMode := LotteryRechargeSingle, LotteryRechargeCumulative
	singleThreshold, cumulativeThreshold := 50.0, 100.0
	repo := &lotteryChanceRepoStub{inviterID: &inviterID, previousTotal: 90, rules: []LotteryRule{
		{ID: 4, Name: "single", EventType: LotteryEventRecharge, Beneficiary: LotteryBeneficiaryInviter, NormalChances: 2, RechargeMode: &singleMode, RechargeThreshold: &singleThreshold},
		{ID: 5, Name: "cumulative", EventType: LotteryEventRecharge, Beneficiary: LotteryBeneficiaryInviter, LuxuryChances: 1, RechargeMode: &cumulativeMode, RechargeThreshold: &cumulativeThreshold, Repeatable: true},
	}}

	require.NoError(t, NewLotteryChanceService(repo).GrantRecharge(context.Background(), 20, 55, 220))
	require.Len(t, repo.grants, 4)
	require.Equal(t, "55", repo.grants[0].SourceID)
	require.Contains(t, repo.grants[0].DedupeKey, "first:20")
	for _, grant := range repo.grants[1:] {
		require.Equal(t, "55", grant.SourceID)
		require.Contains(t, grant.DedupeKey, "cumulative:20")
	}
	require.Equal(t, []int{1, 2, 3}, []int{repo.grants[1].TierNo, repo.grants[2].TierNo, repo.grants[3].TierNo})
}

func TestLotteryChanceRefundOnlyReversesGrantFromRefundedSingleOrder(t *testing.T) {
	inviterID := int64(10)
	mode, threshold := LotteryRechargeSingle, 50.0
	rule := LotteryRule{ID: 6, Name: "first", EventType: LotteryEventRecharge, Beneficiary: LotteryBeneficiaryInviter, NormalChances: 2, RechargeMode: &mode, RechargeThreshold: &threshold}
	repo := &lotteryChanceRepoStub{inviterID: &inviterID, rules: []LotteryRule{rule}}
	service := NewLotteryChanceService(repo)

	require.NoError(t, service.GrantRecharge(context.Background(), 20, 55, 100))
	require.NoError(t, service.ReverseRecharge(context.Background(), 20, 66, 100, 100))
	require.Empty(t, repo.reversals, "后续未获奖订单的退款不能冲正首单奖励")

	require.NoError(t, service.ReverseRecharge(context.Background(), 20, 55, 100, 100))
	require.Len(t, repo.reversals, 1)
	require.True(t, strings.HasPrefix(repo.reversals[0], "reverse:"))
}

func TestGrantRedeemLotteryChanceWritesIdempotentExtraGrant(t *testing.T) {
	repo := &lotteryChanceRepoStub{}
	service := NewLotteryChanceService(repo)

	require.NoError(t, service.GrantRedeemLotteryChance(context.Background(), 20, 91, LotteryPoolLuxury, 3))
	require.Len(t, repo.grants, 1)
	require.Equal(t, int64(20), repo.grants[0].UserID)
	require.Equal(t, LotteryPoolLuxury, repo.grants[0].PoolKey)
	require.EqualValues(t, 3, repo.grants[0].Chances)
	require.Equal(t, int64(0), repo.grants[0].RuleID)
	require.Equal(t, "redeem_code", repo.grants[0].SourceType)
	require.Equal(t, "91", repo.grants[0].SourceID)
	require.Equal(t, "redeem_code:91:luxury", repo.grants[0].DedupeKey)

	require.NoError(t, service.GrantRedeemLotteryChance(context.Background(), 20, 91, LotteryPoolLuxury, 3))
	require.Len(t, repo.grants, 1, "相同 dedupe_key 不应重复发放")
}

func TestGrantRedeemLotteryChanceRejectsInvalidPool(t *testing.T) {
	repo := &lotteryChanceRepoStub{}
	err := NewLotteryChanceService(repo).GrantRedeemLotteryChance(context.Background(), 20, 91, "invalid", 1)
	require.Error(t, err)
	require.Empty(t, repo.grants)
}
