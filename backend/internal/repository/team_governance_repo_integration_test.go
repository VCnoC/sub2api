//go:build integration

package repository

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestTeamGovernanceRepositoryWorkflow(t *testing.T) {
	dsn := os.Getenv("TEAM_GOVERNANCE_TEST_DSN")
	if dsn == "" {
		t.Skip("TEAM_GOVERNANCE_TEST_DSN is not set")
	}
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	ctx := context.Background()
	repo := NewTeamGovernanceRepository(db)

	var ownerID, adminID, memberID int64
	require.NoError(t, db.QueryRowContext(ctx, `SELECT id FROM users WHERE email = 'trigger@test.local'`).Scan(&ownerID))
	require.NoError(t, db.QueryRowContext(ctx, `INSERT INTO users (email, password_hash, role) VALUES ('admin@test.local', 'x', 'admin') RETURNING id`).Scan(&adminID))
	require.NoError(t, db.QueryRowContext(ctx, `INSERT INTO users (email, password_hash) VALUES ('member@test.local', 'x') RETURNING id`).Scan(&memberID))
	_, err = repo.UpdateSettings(ctx, adminID, service.TeamGovernanceSettings{Levels: []service.TeamLevelRequirement{{Limit: 5, Mode: "and"}, {Limit: 15, Mode: "and"}, {Limit: 40, Mode: "and"}}})
	require.NoError(t, err)

	application, err := repo.SubmitCreateApplication(ctx, ownerID, "Integration Team", "test", "")
	require.NoError(t, err)
	application, err = repo.ReviewApplication(ctx, application.ID, adminID, service.ReviewTeamApplicationInput{Approve: true}, "INTEGR8")
	require.NoError(t, err)
	require.NotNil(t, application.CreatedTeamID)

	joinRequest, err := repo.SubmitJoinRequest(ctx, memberID, "INTEGR8", "join")
	require.NoError(t, err)
	joinRequest, err = repo.ReviewJoinRequest(ctx, ownerID, joinRequest.ID, true, "")
	require.NoError(t, err)
	require.Equal(t, service.TeamRequestApproved, joinRequest.Status)
	state, err := repo.UpgradeTeam(ctx, ownerID, *application.CreatedTeamID)
	require.NoError(t, err)
	require.Equal(t, 40, state.Level)
	require.Equal(t, 2, state.MemberCount)

	require.NoError(t, repo.DepositTeamFund(ctx, ownerID, *application.CreatedTeamID, 1))
	detail, err := repo.GetAdminTeam(ctx, *application.CreatedTeamID)
	require.NoError(t, err)
	require.Len(t, detail.Members, 2)
	require.Len(t, detail.FundLedger, 1)
	require.InDelta(t, 1, detail.Team.Balance, 0.000001)

	teams, total, err := repo.ListAdminTeams(ctx, "Integration", "active", 1, 20)
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Len(t, teams, 1)
}
