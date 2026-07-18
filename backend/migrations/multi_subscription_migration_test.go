package migrations

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMultiSubscriptionMigrationsDefineIndependentEntitlements(t *testing.T) {
	linkMigration, err := FS.ReadFile("177_add_payment_order_subscription_link.sql")
	require.NoError(t, err)
	linkSQL := string(linkMigration)
	require.Contains(t, linkSQL, "ADD COLUMN IF NOT EXISTS subscription_id BIGINT")
	require.Contains(t, linkSQL, "REFERENCES user_subscriptions(id)")
	require.Contains(t, linkSQL, "ON DELETE SET NULL")

	indexMigration, err := FS.ReadFile("178_multi_subscription_candidate_indexes_notx.sql")
	require.NoError(t, err)
	indexSQL := string(indexMigration)
	require.Contains(t, indexSQL, "DROP INDEX CONCURRENTLY IF EXISTS user_subscriptions_user_group_unique_active")
	require.Contains(t, indexSQL, "idx_user_subscriptions_candidate_order")
	require.Contains(t, indexSQL, "user_id, group_id, status, expires_at, id")
	require.Contains(t, indexSQL, "WHERE deleted_at IS NULL")
}
