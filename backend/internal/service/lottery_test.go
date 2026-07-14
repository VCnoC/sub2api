// Package service 的抽奖单元测试覆盖周期、概率和管理输入边界。
package service

import (
	"encoding/base64"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLotteryPeriodKeyUsesSiteTimezone(t *testing.T) {
	previous := time.Local
	location, err := time.LoadLocation("Asia/Shanghai")
	require.NoError(t, err)
	time.Local = location
	t.Cleanup(func() { time.Local = previous })

	now := time.Date(2026, 7, 12, 16, 30, 0, 0, time.UTC)
	require.Equal(t, "d:2026-07-13", lotteryPeriodKey(now, LotteryCycleDaily))
	require.Equal(t, "w:2026-07-13", lotteryPeriodKey(now, LotteryCycleWeekly))
}

func TestSelectLotteryPrizeKeepsEmptyAndExhaustedIntervals(t *testing.T) {
	stock := int64(1)
	prizes := []LotteryPrize{
		{ID: 1, Enabled: true, ProbabilityPPM: 200_000},
		{ID: 2, Enabled: true, ProbabilityPPM: 300_000, StockTotal: &stock, StockUsed: 1},
	}

	require.EqualValues(t, 1, selectLotteryPrize(prizes, 0).ID)
	require.EqualValues(t, 1, selectLotteryPrize(prizes, 199_999).ID)
	require.Nil(t, selectLotteryPrize(prizes, 200_000), "库存耗尽的区间应落为未中奖")
	require.Nil(t, selectLotteryPrize(prizes, 500_000), "概率总和不足 100% 的剩余区间应落为未中奖")
}

func TestLotteryValidationRejectsUnsafeAndOversizedValues(t *testing.T) {
	validPNG := "data:image/png;base64," + base64.StdEncoding.EncodeToString([]byte("\x89PNG\r\n\x1a\n"))
	require.NoError(t, validateLotteryImage(validPNG))
	require.ErrorIs(t, validateLotteryImage("data:image/png;base64,"+base64.StdEncoding.EncodeToString([]byte("<script>"))), ErrLotteryInvalidImage)
	require.ErrorIs(t, validateLotteryImage("data:image/gif;base64,R0lGODlh"), ErrLotteryInvalidImage)
	require.ErrorIs(t, validateLotteryImage("data:image/png;base64,"+strings.Repeat("A", base64.StdEncoding.EncodedLen(lotteryImageMaxBytes)+1)), ErrLotteryInvalidImage)

	amount := math.Inf(1)
	require.ErrorIs(t, validateLotteryPrizeInput(LotteryPrizeInput{
		PoolID: 1, Name: "balance", PrizeType: LotteryPrizeBalance,
		BalanceAmount: &amount, ProbabilityPPM: 1,
	}), ErrLotteryInvalidInput)

	mode, threshold := LotteryRechargeSingle, 1.0
	require.ErrorIs(t, validateLotteryRuleInput(LotteryRuleInput{
		Name: "too many", EventType: LotteryEventRecharge, Beneficiary: LotteryBeneficiaryInviter,
		NormalChances: lotteryMaxRuleChances + 1, RechargeMode: &mode, RechargeThreshold: &threshold,
	}), ErrLotteryInvalidInput)
}

func TestLotteryPoolEndTimeIsExclusive(t *testing.T) {
	end := time.Date(2026, 7, 12, 12, 0, 0, 0, time.UTC)
	pool := LotteryPool{Enabled: true, EndsAt: &end}
	require.True(t, pool.ActiveAt(end.Add(-time.Nanosecond)))
	require.False(t, pool.ActiveAt(end))
}
