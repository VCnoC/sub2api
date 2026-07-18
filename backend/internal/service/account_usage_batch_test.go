package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
	"github.com/stretchr/testify/require"
)

type batchUsageAccountRepoStub struct {
	AccountRepository
	accounts      map[int64]*Account
	getByIDsCalls int
}

func (r *batchUsageAccountRepoStub) GetByIDs(_ context.Context, ids []int64) ([]*Account, error) {
	r.getByIDsCalls++
	result := make([]*Account, 0, len(ids))
	for _, id := range ids {
		if account := r.accounts[id]; account != nil {
			result = append(result, account)
		}
	}
	return result, nil
}

func (r *batchUsageAccountRepoStub) GetByID(context.Context, int64) (*Account, error) {
	panic("batch usage must not call GetByID")
}

type batchUsageLogRepoStub struct {
	UsageLogRepository
	todayCalls  int
	windowCalls int
	geminiCalls int
}

func (r *batchUsageLogRepoStub) GetAccountWindowStatsBatch(_ context.Context, ids []int64, _ time.Time) (map[int64]*usagestats.AccountStats, error) {
	r.todayCalls++
	result := make(map[int64]*usagestats.AccountStats, len(ids))
	for _, id := range ids {
		result[id] = &usagestats.AccountStats{Requests: id, Tokens: id * 10}
	}
	return result, nil
}

func (r *batchUsageLogRepoStub) GetAccountWindowStatsByStartBatch(_ context.Context, queries []AccountWindowStatsQuery) (map[string]*usagestats.AccountStats, error) {
	r.windowCalls++
	result := make(map[string]*usagestats.AccountStats, len(queries))
	for _, query := range queries {
		result[query.Key] = &usagestats.AccountStats{Requests: query.AccountID + 1}
	}
	return result, nil
}

func (r *batchUsageLogRepoStub) GetGeminiUsageTotalsBatch(_ context.Context, ids []int64, _, _ time.Time) (map[int64]GeminiUsageTotals, error) {
	r.geminiCalls++
	result := make(map[int64]GeminiUsageTotals, len(ids))
	for _, id := range ids {
		result[id] = GeminiUsageTotals{ProRequests: 2, FlashRequests: 3}
	}
	return result, nil
}

type panicClaudeUsageFetcher struct{}

func (panicClaudeUsageFetcher) FetchUsage(context.Context, string, string) (*ClaudeUsageResponse, error) {
	panic("batch usage must not fetch upstream")
}

func (panicClaudeUsageFetcher) FetchUsageWithOptions(context.Context, *ClaudeUsageFetchOptions) (*ClaudeUsageResponse, error) {
	panic("batch usage must not fetch upstream")
}

func TestAccountUsageServiceGetUsageBatchPassiveUsesConstantBatchCalls(t *testing.T) {
	for _, count := range []int{50, 100} {
		t.Run(fmt.Sprintf("%d accounts", count), func(t *testing.T) {
			now := time.Now()
			accounts := make(map[int64]*Account, count)
			ids := make([]int64, 0, count)
			for i := 1; i <= count; i++ {
				id := int64(i)
				ids = append(ids, id)
				accounts[id] = &Account{
					ID: id, Platform: PlatformAnthropic, Type: AccountTypeSetupToken,
					SessionWindowStart: &now, SessionWindowEnd: batchPtrTime(now.Add(5 * time.Hour)),
					Extra: map[string]any{"session_window_utilization": 0.25},
				}
			}
			accountRepo := &batchUsageAccountRepoStub{accounts: accounts}
			usageRepo := &batchUsageLogRepoStub{}
			svc := NewAccountUsageService(accountRepo, usageRepo, panicClaudeUsageFetcher{}, nil, nil, nil, nil, NewUsageCache(), nil, nil)

			result, err := svc.GetUsageBatchPassive(context.Background(), ids)
			require.NoError(t, err)
			require.Len(t, result.Usage, count)
			require.Len(t, result.TodayStats, count)
			require.Empty(t, result.Errors)
			require.Equal(t, 1, accountRepo.getByIDsCalls)
			require.Equal(t, 1, usageRepo.todayCalls)
			require.Equal(t, 1, usageRepo.windowCalls)
			require.Zero(t, usageRepo.geminiCalls)
		})
	}
}

func TestAccountUsageServiceGetUsageBatchPassivePartialFailuresAndGemini(t *testing.T) {
	now := time.Now()
	accounts := map[int64]*Account{
		1: {ID: 1, Platform: PlatformOpenAI, Type: AccountTypeOAuth, Extra: map[string]any{
			"codex_5h_used_percent": 10.0, "codex_5h_reset_at": now.Add(time.Hour).Format(time.RFC3339),
			"codex_7d_used_percent": 20.0, "codex_7d_reset_at": now.Add(24 * time.Hour).Format(time.RFC3339),
		}},
		2: {ID: 2, Platform: PlatformGemini, Type: AccountTypeAPIKey, Credentials: map[string]any{"tier_id": GeminiTierAIStudioFree}},
		3: {ID: 3, Platform: PlatformVideo, Type: AccountTypeOAuth},
		4: {ID: 4, Platform: PlatformAntigravity, Type: AccountTypeOAuth},
	}
	accountRepo := &batchUsageAccountRepoStub{accounts: accounts}
	usageRepo := &batchUsageLogRepoStub{}
	svc := NewAccountUsageService(accountRepo, usageRepo, panicClaudeUsageFetcher{}, NewGeminiQuotaService(nil, nil), nil, nil, nil, NewUsageCache(), nil, nil)

	result, err := svc.GetUsageBatchPassive(context.Background(), []int64{1, 2, 3, 4, 999})
	require.NoError(t, err)
	require.Equal(t, "passive", result.Usage[1].Source)
	require.NotNil(t, result.Usage[1].FiveHour.WindowStats)
	require.NotNil(t, result.Usage[2].GeminiProDaily)
	require.Equal(t, "unsupported_platform", result.Errors[3].Code)
	require.Equal(t, "snapshot_unavailable", result.Errors[4].Code)
	require.Equal(t, "not_found", result.Errors[999].Code)
	require.Equal(t, 2, usageRepo.geminiCalls)
	require.Equal(t, 1, accountRepo.getByIDsCalls)
}

func TestAccountUsageServiceGetUsageBatchPassiveClonesAntigravityCache(t *testing.T) {
	reset := time.Now().Add(time.Hour)
	cached := &UsageInfo{FiveHour: &UsageProgress{ResetsAt: &reset, RemainingSeconds: 1}}
	cache := NewUsageCache()
	cache.antigravityCache.Store(int64(7), &antigravityUsageCache{usageInfo: cached, timestamp: time.Now()})
	accountRepo := &batchUsageAccountRepoStub{accounts: map[int64]*Account{
		7: {ID: 7, Platform: PlatformAntigravity, Type: AccountTypeOAuth},
	}}
	svc := NewAccountUsageService(accountRepo, &batchUsageLogRepoStub{}, panicClaudeUsageFetcher{}, nil, nil, nil, nil, cache, nil, nil)

	result, err := svc.GetUsageBatchPassive(context.Background(), []int64{7})
	require.NoError(t, err)
	require.NotSame(t, cached, result.Usage[7])
	require.Equal(t, 1, cached.FiveHour.RemainingSeconds)
	require.Greater(t, result.Usage[7].FiveHour.RemainingSeconds, 1)
}

func batchPtrTime(value time.Time) *time.Time { return &value }
