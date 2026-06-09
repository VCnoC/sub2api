package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestAssignOrExtendSubscription_UnexpiredRenewResetsUsageAndWindows 验证 bug 修复：
// 用户的月度额度已耗尽（但订阅日期未过期），用激活码续期相同套餐时，
// 应当重置 daily/weekly/monthly 用量计数和窗口，否则用户付了钱仍然无法使用。
func TestAssignOrExtendSubscription_UnexpiredRenewResetsUsageAndWindows(t *testing.T) {
	groupRepo := &subscriptionGroupRepoStub{
		group: &Group{ID: 1, SubscriptionType: SubscriptionTypeSubscription},
	}
	subRepo := newSubscriptionUserSubRepoStub()

	originalStart := time.Now().AddDate(0, 0, -15)
	originalExpiresAt := originalStart.AddDate(0, 0, 30)
	originalWindowStart := startOfDay(originalStart)
	subRepo.seed(&UserSubscription{
		ID:                 100,
		UserID:             200,
		GroupID:            1,
		StartsAt:           originalStart,
		ExpiresAt:          originalExpiresAt,
		Status:             SubscriptionStatusActive,
		DailyWindowStart:   &originalWindowStart,
		WeeklyWindowStart:  &originalWindowStart,
		MonthlyWindowStart: &originalWindowStart,
		DailyUsageUSD:      5,
		WeeklyUsageUSD:     35,
		MonthlyUsageUSD:    100, // 月度额度耗尽
		Notes:              "first",
	})
	svc := NewSubscriptionService(groupRepo, subRepo, nil, nil, nil)

	renewed, reused, err := svc.AssignOrExtendSubscription(context.Background(), &AssignSubscriptionInput{
		UserID:       200,
		GroupID:      1,
		ValidityDays: 30,
		Notes:        "renew",
	})

	require.NoError(t, err)
	require.True(t, reused, "应识别为续期场景")
	require.Equal(t, SubscriptionStatusActive, renewed.Status)
	require.Equal(t, originalStart, renewed.StartsAt, "未过期续期应保留原 StartsAt")
	require.True(t, renewed.ExpiresAt.After(originalExpiresAt), "ExpiresAt 应当延长（叠加 30 天）")

	expectedExpiresAt := originalExpiresAt.AddDate(0, 0, 30)
	require.WithinDuration(t, expectedExpiresAt, renewed.ExpiresAt, time.Second)

	// 核心断言：用量窗口和计数必须被重置
	require.Equal(t, 0.0, renewed.DailyUsageUSD, "续期必须重置日用量")
	require.Equal(t, 0.0, renewed.WeeklyUsageUSD, "续期必须重置周用量")
	require.Equal(t, 0.0, renewed.MonthlyUsageUSD, "续期必须重置月用量（修复 bug 的关键）")

	require.NotNil(t, renewed.DailyWindowStart)
	require.NotNil(t, renewed.WeeklyWindowStart)
	require.NotNil(t, renewed.MonthlyWindowStart)
	require.True(t, renewed.MonthlyWindowStart.After(originalWindowStart),
		"月度窗口必须刷新为新的窗口起点，否则用户仍被限额限制")

	require.Equal(t, "first\nrenew", renewed.Notes)
}

// TestAssignOrResetSubscription_UnexpiredAssignmentDiscardsRemainingDays 验证：
// 管理员"重新分配"语义 = ExpiresAt = now + ValidityDays，
// 原剩余天数被作废（与 AssignOrExtendSubscription 的"叠加"语义对比）。
func TestAssignOrResetSubscription_UnexpiredAssignmentDiscardsRemainingDays(t *testing.T) {
	groupRepo := &subscriptionGroupRepoStub{
		group: &Group{ID: 1, SubscriptionType: SubscriptionTypeSubscription},
	}
	subRepo := newSubscriptionUserSubRepoStub()

	originalStart := time.Now().AddDate(0, 0, -15)
	originalExpiresAt := time.Now().AddDate(0, 0, 20) // 还剩 20 天
	originalWindowStart := startOfDay(originalStart)
	subRepo.seed(&UserSubscription{
		ID:                 200,
		UserID:             300,
		GroupID:            1,
		StartsAt:           originalStart,
		ExpiresAt:          originalExpiresAt,
		Status:             SubscriptionStatusActive,
		DailyWindowStart:   &originalWindowStart,
		WeeklyWindowStart:  &originalWindowStart,
		MonthlyWindowStart: &originalWindowStart,
		DailyUsageUSD:      8,
		WeeklyUsageUSD:     50,
		MonthlyUsageUSD:    100,
		Notes:              "old",
	})
	svc := NewSubscriptionService(groupRepo, subRepo, nil, nil, nil)

	beforeAssign := time.Now()
	renewed, reused, err := svc.AssignOrResetSubscription(context.Background(), &AssignSubscriptionInput{
		UserID:       300,
		GroupID:      1,
		ValidityDays: 30,
		Notes:        "reassign",
	})

	require.NoError(t, err)
	require.True(t, reused)
	require.Equal(t, SubscriptionStatusActive, renewed.Status)

	// 核心：原剩余 20 天作废，新 ExpiresAt = now + 30 天（约 30 天后）
	expectedExpiresAt := beforeAssign.AddDate(0, 0, 30)
	require.WithinDuration(t, expectedExpiresAt, renewed.ExpiresAt, 5*time.Second,
		"重新分配应当从 now 开始计算 ExpiresAt，而不是叠加原剩余天数")
	require.True(t, renewed.ExpiresAt.Before(originalExpiresAt.AddDate(0, 0, 30)),
		"ExpiresAt 应当比'叠加语义'短（原剩 20 天 + 30 天 = 50 天后）")

	// StartsAt 重置为 now
	require.True(t, renewed.StartsAt.After(originalStart), "StartsAt 应被重置")

	// 用量全部重置
	require.Equal(t, 0.0, renewed.DailyUsageUSD)
	require.Equal(t, 0.0, renewed.WeeklyUsageUSD)
	require.Equal(t, 0.0, renewed.MonthlyUsageUSD)
}

// TestAssignOrResetSubscription_NoExistingCreatesNew 验证：用户没有订阅时，行为与 createSubscription 一致。
func TestAssignOrResetSubscription_NoExistingCreatesNew(t *testing.T) {
	groupRepo := &subscriptionGroupRepoStub{
		group: &Group{ID: 1, SubscriptionType: SubscriptionTypeSubscription},
	}
	subRepo := newSubscriptionUserSubRepoStub()
	svc := NewSubscriptionService(groupRepo, subRepo, nil, nil, nil)

	sub, reused, err := svc.AssignOrResetSubscription(context.Background(), &AssignSubscriptionInput{
		UserID:       400,
		GroupID:      1,
		ValidityDays: 30,
		Notes:        "first time",
	})

	require.NoError(t, err)
	require.False(t, reused, "不存在订阅时应是新建场景")
	require.NotZero(t, sub.ID)
	require.Equal(t, SubscriptionStatusActive, sub.Status)
	require.Equal(t, 0.0, sub.MonthlyUsageUSD)
}

// TestAssignOrExtendSubscription_SuspendedSubscriptionRestoredAndUsageReset 验证：
// 暂停状态的订阅续期时，应被恢复为 active 并重置用量。
func TestAssignOrExtendSubscription_SuspendedSubscriptionRestoredAndUsageReset(t *testing.T) {
	groupRepo := &subscriptionGroupRepoStub{
		group: &Group{ID: 1, SubscriptionType: SubscriptionTypeSubscription},
	}
	subRepo := newSubscriptionUserSubRepoStub()

	originalStart := time.Now().AddDate(0, 0, -5)
	originalExpiresAt := originalStart.AddDate(0, 0, 30)
	originalWindowStart := startOfDay(originalStart)
	subRepo.seed(&UserSubscription{
		ID:                 101,
		UserID:             201,
		GroupID:            1,
		StartsAt:           originalStart,
		ExpiresAt:          originalExpiresAt,
		Status:             SubscriptionStatusSuspended,
		DailyWindowStart:   &originalWindowStart,
		WeeklyWindowStart:  &originalWindowStart,
		MonthlyWindowStart: &originalWindowStart,
		DailyUsageUSD:      7,
		WeeklyUsageUSD:     0,
		MonthlyUsageUSD:    50,
	})
	svc := NewSubscriptionService(groupRepo, subRepo, nil, nil, nil)

	renewed, reused, err := svc.AssignOrExtendSubscription(context.Background(), &AssignSubscriptionInput{
		UserID:       201,
		GroupID:      1,
		ValidityDays: 15,
		Notes:        "reactivate",
	})

	require.NoError(t, err)
	require.True(t, reused)
	require.Equal(t, SubscriptionStatusActive, renewed.Status, "暂停订阅续期后应恢复 active")
	require.Equal(t, 0.0, renewed.DailyUsageUSD)
	require.Equal(t, 0.0, renewed.MonthlyUsageUSD)
}
