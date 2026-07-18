package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestUsageLogRepositoryGetAccountWindowStatsByStartBatch(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	repo := newUsageLogRepositoryWithSQL(nil, db)

	mock.ExpectQuery(`(?s)WITH windows AS .*FROM unnest.*LEFT JOIN usage_logs`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"query_key", "requests", "tokens", "cost", "standard_cost", "user_cost"}).
			AddRow("1:5h", int64(2), int64(100), 1.5, 1.25, 1.1).
			AddRow("1:7d", int64(8), int64(900), 9.5, 8.25, 7.1))

	queries := []service.AccountWindowStatsQuery{
		{Key: "1:5h", AccountID: 1, StartTime: time.Now().Add(-5 * time.Hour)},
		{Key: "1:7d", AccountID: 1, StartTime: time.Now().Add(-7 * 24 * time.Hour)},
	}
	stats, err := repo.GetAccountWindowStatsByStartBatch(context.Background(), queries)
	require.NoError(t, err)
	require.Equal(t, int64(2), stats["1:5h"].Requests)
	require.Equal(t, int64(8), stats["1:7d"].Requests)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUsageLogRepositoryGetAccountWindowStatsByStartBatchEmpty(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	repo := newUsageLogRepositoryWithSQL(nil, db)

	stats, err := repo.GetAccountWindowStatsByStartBatch(context.Background(), nil)
	require.NoError(t, err)
	require.Empty(t, stats)
	require.NoError(t, mock.ExpectationsWereMet())
}
