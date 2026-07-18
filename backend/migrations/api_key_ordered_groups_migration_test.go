package migrations

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAPIKeyOrderedGroupsMigrationDefinesOrderAndBackfill(t *testing.T) {
	migration, err := FS.ReadFile("185_api_key_ordered_groups.sql")
	require.NoError(t, err)
	sql := string(migration)

	require.Contains(t, sql, "CREATE TABLE IF NOT EXISTS api_key_groups")
	require.Contains(t, sql, "api_key_groups_api_key_group_key UNIQUE (api_key_id, group_id)")
	require.Contains(t, sql, "UNIQUE (api_key_id, priority)")
	require.Contains(t, sql, "priority >= 0 AND priority < 5")
	require.Contains(t, sql, "SELECT id, group_id, 0")
	require.Contains(t, sql, "WHERE group_id IS NOT NULL AND deleted_at IS NULL")
}
