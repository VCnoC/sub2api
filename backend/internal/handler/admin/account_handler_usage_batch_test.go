package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type usageBatchAccountRepoStub struct {
	service.AccountRepository
	accounts      map[int64]*service.Account
	getByIDsCalls int
}

func (r *usageBatchAccountRepoStub) GetByIDs(_ context.Context, ids []int64) ([]*service.Account, error) {
	r.getByIDsCalls++
	result := make([]*service.Account, 0, len(ids))
	for _, id := range ids {
		if account := r.accounts[id]; account != nil {
			result = append(result, account)
		}
	}
	return result, nil
}

type usageBatchLogRepoStub struct{ service.UsageLogRepository }

func (usageBatchLogRepoStub) GetAccountWindowStatsBatch(_ context.Context, ids []int64, _ time.Time) (map[int64]*usagestats.AccountStats, error) {
	result := make(map[int64]*usagestats.AccountStats, len(ids))
	for _, id := range ids {
		result[id] = &usagestats.AccountStats{Requests: 3}
	}
	return result, nil
}

func (usageBatchLogRepoStub) GetAccountWindowStatsByStartBatch(_ context.Context, queries []service.AccountWindowStatsQuery) (map[string]*usagestats.AccountStats, error) {
	result := make(map[string]*usagestats.AccountStats, len(queries))
	for _, query := range queries {
		result[query.Key] = &usagestats.AccountStats{Requests: 4}
	}
	return result, nil
}

func setupUsageBatchRouter() (*gin.Engine, *usageBatchAccountRepoStub) {
	gin.SetMode(gin.TestMode)
	accountUsageBatchCache = newSnapshotCache(30 * time.Second)
	end := time.Now().Add(5 * time.Hour)
	accountRepo := &usageBatchAccountRepoStub{accounts: map[int64]*service.Account{
		1: {ID: 1, Platform: service.PlatformAnthropic, Type: service.AccountTypeSetupToken, SessionWindowEnd: &end},
	}}
	usageSvc := service.NewAccountUsageService(accountRepo, usageBatchLogRepoStub{}, nil, nil, nil, nil, nil, nil, service.NewUsageCache(), nil, nil)
	handler := NewAccountHandler(nil, nil, nil, nil, nil, nil, nil, usageSvc, nil, nil, nil, nil, nil, nil)
	router := gin.New()
	router.POST("/api/v1/admin/accounts/usage/batch", handler.GetUsageBatch)
	return router, accountRepo
}

func TestAccountHandlerGetUsageBatchValidation(t *testing.T) {
	router, _ := setupUsageBatchRouter()
	cases := []struct {
		name string
		body any
	}{
		{name: "missing account_ids", body: map[string]any{}},
		{name: "non positive id", body: map[string]any{"account_ids": []int64{1, 0}}},
		{name: "over unique limit", body: map[string]any{"account_ids": usageBatchIDs(101)}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			response := performUsageBatchRequest(t, router, tc.body, "")
			require.Equal(t, http.StatusBadRequest, response.Code)
		})
	}
}

func TestAccountHandlerGetUsageBatchEmptyAndPartialSuccess(t *testing.T) {
	router, accountRepo := setupUsageBatchRouter()
	empty := performUsageBatchRequest(t, router, map[string]any{"account_ids": []int64{}}, "")
	require.Equal(t, http.StatusOK, empty.Code)
	var emptyEnvelope struct {
		Data service.BatchUsageSnapshot `json:"data"`
	}
	require.NoError(t, json.Unmarshal(empty.Body.Bytes(), &emptyEnvelope))
	require.Empty(t, emptyEnvelope.Data.Usage)

	first := performUsageBatchRequest(t, router, map[string]any{"account_ids": []int64{1, 1, 999}}, "")
	require.Equal(t, http.StatusOK, first.Code)
	require.Equal(t, "miss", first.Header().Get("X-Snapshot-Cache"))
	require.NotEmpty(t, first.Header().Get("ETag"))
	var envelope struct {
		Data service.BatchUsageSnapshot `json:"data"`
	}
	require.NoError(t, json.Unmarshal(first.Body.Bytes(), &envelope))
	require.Equal(t, "not_found", envelope.Data.Errors[999].Code)
	require.Equal(t, 1, accountRepo.getByIDsCalls)

	second := performUsageBatchRequest(t, router, map[string]any{"account_ids": []int64{999, 1}}, first.Header().Get("ETag"))
	require.Equal(t, http.StatusNotModified, second.Code)
	require.Equal(t, 1, accountRepo.getByIDsCalls)
}

func performUsageBatchRequest(t *testing.T, router http.Handler, body any, etag string) *httptest.ResponseRecorder {
	t.Helper()
	raw, err := json.Marshal(body)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/accounts/usage/batch", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	if etag != "" {
		req.Header.Set("If-None-Match", etag)
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	return recorder
}

func usageBatchIDs(count int) []int64 {
	ids := make([]int64, count)
	for i := range ids {
		ids[i] = int64(i + 1)
	}
	return ids
}
