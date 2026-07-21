package migrations

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRequestCountSubscriptionsMigrationDefinesLimitsAndReservations(t *testing.T) {
	migration, err := FS.ReadFile("186_request_count_subscriptions.sql")
	require.NoError(t, err)
	sql := string(migration)

	require.Contains(t, sql, "subscription_billing_mode VARCHAR(20) NOT NULL DEFAULT 'usd'")
	require.Contains(t, sql, "request_limit_5h INTEGER NOT NULL DEFAULT 0")
	require.Contains(t, sql, "request_usage_1d INTEGER NOT NULL DEFAULT 0")
	require.Contains(t, sql, "CREATE TABLE IF NOT EXISTS subscription_request_reservations")
	require.Contains(t, sql, "UNIQUE (request_id, subscription_id)")
	require.Contains(t, sql, "WHERE status = 'pending'")
}
