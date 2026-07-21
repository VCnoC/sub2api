package migrations

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTeamGovernanceMigrationDefinesApprovalAndTransferableBalanceGuards(t *testing.T) {
	migration, err := FS.ReadFile("187_team_governance.sql")
	require.NoError(t, err)
	sql := string(migration)

	require.Contains(t, sql, "CREATE TABLE IF NOT EXISTS team_applications")
	require.Contains(t, sql, "configured BOOLEAN NOT NULL DEFAULT FALSE")
	require.Contains(t, sql, "CREATE TABLE IF NOT EXISTS team_join_requests")
	require.Contains(t, sql, "WHERE application_type = 'create' AND status = 'pending'")
	require.Contains(t, sql, "WHERE status = 'pending'")
	require.Contains(t, sql, "CREATE TABLE IF NOT EXISTS team_transferable_balances")
	require.Contains(t, sql, "SELECT id, GREATEST(balance, 0) FROM users")
	require.Contains(t, sql, "sync_team_transferable_on_balance_decrease")
	require.Contains(t, sql, "sync_team_transferable_on_redeem")
	require.Contains(t, sql, "COALESCE(NEW.notes, '') NOT LIKE '[lottery] %'")
	require.Contains(t, sql, "NEW.type IN ('balance', 'admin_balance')")
}
