package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/timezone"
	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
)

type AccountWindowStatsQuery struct {
	Key       string
	AccountID int64
	StartTime time.Time
}

type BatchUsageError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type BatchUsageSnapshot struct {
	Usage      map[int64]*UsageInfo       `json:"usage"`
	TodayStats map[int64]*WindowStats     `json:"today_stats"`
	Errors     map[int64]*BatchUsageError `json:"errors"`
}

type accountWindowStatsByStartBatchReader interface {
	GetAccountWindowStatsByStartBatch(ctx context.Context, queries []AccountWindowStatsQuery) (map[string]*usagestats.AccountStats, error)
}

type geminiUsageTotalsBatchReader interface {
	GetGeminiUsageTotalsBatch(ctx context.Context, accountIDs []int64, startTime, endTime time.Time) (map[int64]GeminiUsageTotals, error)
}

// GetUsageBatchPassive returns local snapshots only. It deliberately does not
// call any quota fetcher or the single-account usage methods.
func (s *AccountUsageService) GetUsageBatchPassive(ctx context.Context, accountIDs []int64) (*BatchUsageSnapshot, error) {
	result := newBatchUsageSnapshot(accountIDs)
	if len(accountIDs) == 0 {
		return result, nil
	}
	accounts, err := s.accountRepo.GetByIDs(ctx, accountIDs)
	if err != nil {
		return nil, fmt.Errorf("batch get accounts: %w", err)
	}

	byID := make(map[int64]*Account, len(accounts))
	foundIDs := make([]int64, 0, len(accounts))
	for _, account := range accounts {
		if account == nil {
			continue
		}
		byID[account.ID] = account
		foundIDs = append(foundIDs, account.ID)
	}
	for _, id := range accountIDs {
		if byID[id] == nil {
			result.Errors[id] = batchUsageError("not_found", "account not found")
		}
	}

	today, err := s.getTodayStatsBatchStrict(ctx, foundIDs)
	if err != nil {
		return nil, err
	}
	for id, stats := range today {
		result.TodayStats[id] = stats
	}

	now := time.Now()
	windowQueries := make([]AccountWindowStatsQuery, 0, len(accounts)*2)
	geminiIDs := make([]int64, 0, len(accounts))
	geminiQuotas := make(map[int64]GeminiQuota)
	for _, account := range accounts {
		usage, usageErr := s.buildPassiveUsageBase(ctx, account, now)
		result.Usage[account.ID] = usage
		if usageErr != nil {
			result.Errors[account.ID] = usageErr
		}
		windowQueries = append(windowQueries, usageWindowQueries(account, usage, now)...)
		if account.Platform == PlatformGemini && s.geminiQuotaService != nil {
			if quota, ok := s.geminiQuotaService.QuotaForAccount(ctx, account); ok {
				geminiIDs = append(geminiIDs, account.ID)
				geminiQuotas[account.ID] = quota
			}
		}
	}

	windowStats, err := s.getWindowStatsByStartBatch(ctx, windowQueries)
	if err != nil {
		return nil, err
	}
	applyBatchWindowStats(accounts, result.Usage, windowStats)

	if err := s.applyGeminiBatchUsage(ctx, now, geminiIDs, geminiQuotas, result); err != nil {
		return nil, err
	}
	for _, account := range accounts {
		if account.Platform == PlatformGrok && result.Usage[account.ID] != nil {
			result.Usage[account.ID].GrokLocalUsage = result.TodayStats[account.ID]
		}
	}
	return result, nil
}

func newBatchUsageSnapshot(ids []int64) *BatchUsageSnapshot {
	result := &BatchUsageSnapshot{
		Usage:      make(map[int64]*UsageInfo, len(ids)),
		TodayStats: make(map[int64]*WindowStats, len(ids)),
		Errors:     make(map[int64]*BatchUsageError),
	}
	for _, id := range ids {
		result.Usage[id] = nil
	}
	return result
}

func (s *AccountUsageService) buildPassiveUsageBase(ctx context.Context, account *Account, now time.Time) (*UsageInfo, *BatchUsageError) {
	if account == nil {
		return nil, batchUsageError("not_found", "account not found")
	}
	switch account.Platform {
	case PlatformAnthropic:
		if !account.IsAnthropicOAuthOrSetupToken() {
			return nil, batchUsageError("unsupported_account", "account type does not support usage snapshots")
		}
		usage := s.estimateSetupTokenUsage(account)
		usage.Source = "passive"
		usage.SevenDay = buildPassiveUsageWindow(account.Extra, "passive_usage_7d_utilization", "passive_usage_7d_reset")
		usage.SevenDayFable = buildPassiveUsageWindow(account.Extra, "passive_usage_7d_oi_utilization", "passive_usage_7d_oi_reset")
		usage.UpdatedAt = passiveSampledAt(account.Extra)
		return usage, nil
	case PlatformOpenAI:
		if !account.IsOpenAIOAuth() {
			return nil, batchUsageError("unsupported_account", "account type does not support usage snapshots")
		}
		usage := &UsageInfo{Source: "passive"}
		applyExtraToUsage(usage, account.Extra, now)
		usage.UpdatedAt = extraTimestamp(account.Extra, "codex_usage_updated_at")
		if usage.FiveHour == nil && usage.SevenDay == nil {
			return usage, batchUsageError("snapshot_unavailable", "local usage snapshot unavailable")
		}
		return usage, nil
	case PlatformGemini:
		return &UsageInfo{Source: "passive", UpdatedAt: &now}, nil
	case PlatformAntigravity:
		usage := s.antigravityUsageFromCache(account.ID)
		if usage == nil {
			return nil, batchUsageError("snapshot_unavailable", "local usage snapshot unavailable")
		}
		usage.Source = "passive"
		enrichUsageWithAccountError(usage, account)
		return usage, nil
	case PlatformGrok:
		usage := NewGrokQuotaFetcher().BuildUsageInfo(account)
		if usage.ErrorCode == "quota_unknown" {
			return usage, batchUsageError("snapshot_unavailable", "local usage snapshot unavailable")
		}
		enrichUsageWithAccountError(usage, account)
		return usage, nil
	default:
		return nil, batchUsageError("unsupported_platform", "platform does not support usage snapshots")
	}
}

func (s *AccountUsageService) getTodayStatsBatchStrict(ctx context.Context, ids []int64) (map[int64]*WindowStats, error) {
	result := make(map[int64]*WindowStats, len(ids))
	if len(ids) == 0 {
		return result, nil
	}
	reader, ok := s.usageLogRepo.(accountWindowStatsBatchReader)
	if !ok {
		return nil, fmt.Errorf("usage repository does not support batch today stats")
	}
	stats, err := reader.GetAccountWindowStatsBatch(ctx, ids, timezone.Today())
	if err != nil {
		return nil, fmt.Errorf("batch get today stats: %w", err)
	}
	for _, id := range ids {
		result[id] = windowStatsFromAccountStats(stats[id])
	}
	return result, nil
}

func (s *AccountUsageService) getWindowStatsByStartBatch(ctx context.Context, queries []AccountWindowStatsQuery) (map[string]*usagestats.AccountStats, error) {
	if len(queries) == 0 {
		return map[string]*usagestats.AccountStats{}, nil
	}
	reader, ok := s.usageLogRepo.(accountWindowStatsByStartBatchReader)
	if !ok {
		return nil, fmt.Errorf("usage repository does not support batch window stats")
	}
	stats, err := reader.GetAccountWindowStatsByStartBatch(ctx, queries)
	if err != nil {
		return nil, fmt.Errorf("batch get window stats: %w", err)
	}
	return stats, nil
}

func usageWindowQueries(account *Account, usage *UsageInfo, now time.Time) []AccountWindowStatsQuery {
	if account == nil || usage == nil {
		return nil
	}
	switch account.Platform {
	case PlatformAnthropic:
		return []AccountWindowStatsQuery{{Key: usageWindowKey(account.ID, "5h"), AccountID: account.ID, StartTime: account.GetCurrentWindowStartTime()}}
	case PlatformOpenAI:
		return []AccountWindowStatsQuery{
			{Key: usageWindowKey(account.ID, "5h"), AccountID: account.ID, StartTime: codexWindowStatsStart(usage.FiveHour, 5*time.Hour, now)},
			{Key: usageWindowKey(account.ID, "7d"), AccountID: account.ID, StartTime: codexWindowStatsStart(usage.SevenDay, 7*24*time.Hour, now)},
		}
	default:
		return nil
	}
}

func applyBatchWindowStats(accounts []*Account, usageByID map[int64]*UsageInfo, stats map[string]*usagestats.AccountStats) {
	for _, account := range accounts {
		usage := usageByID[account.ID]
		if usage == nil {
			continue
		}
		if account.Platform == PlatformAnthropic || account.Platform == PlatformOpenAI {
			if usage.FiveHour == nil {
				usage.FiveHour = &UsageProgress{}
			}
			usage.FiveHour.WindowStats = windowStatsFromAccountStats(stats[usageWindowKey(account.ID, "5h")])
		}
		if account.Platform == PlatformOpenAI {
			if usage.SevenDay == nil {
				usage.SevenDay = &UsageProgress{}
			}
			usage.SevenDay.WindowStats = windowStatsFromAccountStats(stats[usageWindowKey(account.ID, "7d")])
		}
	}
}

func (s *AccountUsageService) applyGeminiBatchUsage(ctx context.Context, now time.Time, ids []int64, quotas map[int64]GeminiQuota, result *BatchUsageSnapshot) error {
	if len(ids) == 0 {
		return nil
	}
	reader, ok := s.usageLogRepo.(geminiUsageTotalsBatchReader)
	if !ok {
		return fmt.Errorf("usage repository does not support batch gemini stats")
	}
	dayStart, minuteStart := geminiDailyWindowStart(now), now.Truncate(time.Minute)
	dayTotals, err := reader.GetGeminiUsageTotalsBatch(ctx, ids, dayStart, now)
	if err != nil {
		return fmt.Errorf("batch get gemini daily stats: %w", err)
	}
	minuteTotals, err := reader.GetGeminiUsageTotalsBatch(ctx, ids, minuteStart, now)
	if err != nil {
		return fmt.Errorf("batch get gemini minute stats: %w", err)
	}
	for _, id := range ids {
		result.Usage[id] = buildGeminiUsageFromTotals(quotas[id], dayTotals[id], minuteTotals[id], now)
	}
	return nil
}

func buildGeminiUsageFromTotals(quota GeminiQuota, daily, minute GeminiUsageTotals, now time.Time) *UsageInfo {
	usage := &UsageInfo{Source: "passive", UpdatedAt: &now}
	dailyReset, minuteReset := geminiDailyResetTime(now), now.Truncate(time.Minute).Add(time.Minute)
	if quota.SharedRPD > 0 {
		usage.GeminiSharedDaily = buildGeminiUsageProgress(daily.ProRequests+daily.FlashRequests, quota.SharedRPD, dailyReset, daily.ProTokens+daily.FlashTokens, daily.ProCost+daily.FlashCost, now)
	} else {
		usage.GeminiProDaily = buildGeminiUsageProgress(daily.ProRequests, quota.ProRPD, dailyReset, daily.ProTokens, daily.ProCost, now)
		usage.GeminiFlashDaily = buildGeminiUsageProgress(daily.FlashRequests, quota.FlashRPD, dailyReset, daily.FlashTokens, daily.FlashCost, now)
	}
	if quota.SharedRPM > 0 {
		usage.GeminiSharedMinute = buildGeminiUsageProgress(minute.ProRequests+minute.FlashRequests, quota.SharedRPM, minuteReset, minute.ProTokens+minute.FlashTokens, minute.ProCost+minute.FlashCost, now)
	} else {
		usage.GeminiProMinute = buildGeminiUsageProgress(minute.ProRequests, quota.ProRPM, minuteReset, minute.ProTokens, minute.ProCost, now)
		usage.GeminiFlashMinute = buildGeminiUsageProgress(minute.FlashRequests, quota.FlashRPM, minuteReset, minute.FlashTokens, minute.FlashCost, now)
	}
	return usage
}

func (s *AccountUsageService) antigravityUsageFromCache(accountID int64) *UsageInfo {
	if s == nil || s.cache == nil {
		return nil
	}
	raw, ok := s.cache.antigravityCache.Load(accountID)
	if !ok {
		return nil
	}
	entry, ok := raw.(*antigravityUsageCache)
	if !ok || entry == nil || entry.usageInfo == nil {
		return nil
	}
	data, err := json.Marshal(entry.usageInfo)
	if err != nil {
		return nil
	}
	var clone UsageInfo
	if json.Unmarshal(data, &clone) != nil {
		return nil
	}
	recalcAntigravityRemainingSeconds(&clone)
	return &clone
}

func passiveSampledAt(extra map[string]any) *time.Time {
	return extraTimestamp(extra, "passive_usage_sampled_at")
}

func extraTimestamp(extra map[string]any, key string) *time.Time {
	raw, ok := extra[key]
	if !ok {
		return nil
	}
	parsed, err := parseTime(fmt.Sprint(raw))
	if err != nil {
		return nil
	}
	return &parsed
}

func usageWindowKey(accountID int64, window string) string {
	return fmt.Sprintf("%d:%s", accountID, window)
}

func batchUsageError(code, message string) *BatchUsageError {
	return &BatchUsageError{Code: code, Message: message}
}
