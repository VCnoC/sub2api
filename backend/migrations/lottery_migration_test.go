// Package migrations 的抽奖迁移回归测试锁定核心表、约束和固定奖池种子。
package migrations

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMigration176DefinesLotteryIntegrityAndFixedPools(t *testing.T) {
	content, err := FS.ReadFile("176_lottery_system.sql")
	require.NoError(t, err)
	sql := string(content)

	for _, table := range []string{
		"lottery_pools", "lottery_prizes", "lottery_rules",
		"lottery_user_chances", "lottery_chance_ledger", "lottery_draws",
	} {
		require.Contains(t, sql, "CREATE TABLE IF NOT EXISTS "+table)
	}
	require.Contains(t, sql, "key IN ('normal', 'luxury')")
	require.Contains(t, sql, "normal_chances BETWEEN 0 AND 100000")
	require.Contains(t, sql, "extra_remaining >= 0")
	require.Contains(t, sql, "dedupe_key VARCHAR(255) NOT NULL UNIQUE")
	require.Contains(t, sql, "UNIQUE (user_id, pool_id, idempotency_key)")
	require.Contains(t, sql, "ON CONFLICT (key) DO NOTHING")
}
