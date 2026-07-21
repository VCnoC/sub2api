//go:build integration

package repository

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

func newRequestCountFixture(t *testing.T, limit5h, limit1d int) (*userSubscriptionRepository, *service.User, *service.Group, *service.APIKey) {
	t.Helper()
	client := testEntClient(t)
	user := mustCreateUser(t, client, &service.User{
		Email:        fmt.Sprintf("request-count-%d@example.com", time.Now().UnixNano()),
		PasswordHash: "hash",
	})
	group := mustCreateGroup(t, client, &service.Group{
		Name:                    fmt.Sprintf("request-count-%d", time.Now().UnixNano()),
		SubscriptionType:        service.SubscriptionTypeSubscription,
		SubscriptionBillingMode: service.SubscriptionBillingModeRequestCount,
		RequestLimit5h:          limit5h,
		RequestLimit1d:          limit1d,
	})
	apiKey := mustCreateApiKey(t, client, &service.APIKey{
		UserID:  user.ID,
		GroupID: &group.ID,
		Key:     fmt.Sprintf("sk-request-count-%d", time.Now().UnixNano()),
		Name:    "request-count",
	})
	return NewUserSubscriptionRepository(client).(*userSubscriptionRepository), user, group, apiKey
}

func TestRequestCountReservationLifecycle(t *testing.T) {
	ctx := context.Background()
	repo, user, group, apiKey := newRequestCountFixture(t, 2, 3)
	sub := mustCreateSubscription(t, repo.client, &service.UserSubscription{UserID: user.ID, GroupID: group.ID})

	reservation, err := repo.ReserveRequestCount(ctx, "request-lifecycle", apiKey.ID, user.ID, group.ID, time.Now().Add(time.Minute))
	require.NoError(t, err)
	duplicate, err := repo.ReserveRequestCount(ctx, "request-lifecycle", apiKey.ID, user.ID, group.ID, time.Now().Add(time.Minute))
	require.NoError(t, err)
	require.Equal(t, reservation.ID, duplicate.ID)

	current, err := repo.GetByID(ctx, sub.ID)
	require.NoError(t, err)
	require.Equal(t, 1, current.RequestUsage5h)
	require.Equal(t, 1, current.RequestUsage1d)
	require.NotNil(t, current.RequestWindow5hStart)
	require.NotNil(t, current.RequestWindow1dStart)

	require.NoError(t, repo.ReleaseRequestCount(ctx, reservation.ID))
	require.NoError(t, repo.ReleaseRequestCount(ctx, reservation.ID))
	current, err = repo.GetByID(ctx, sub.ID)
	require.NoError(t, err)
	require.Zero(t, current.RequestUsage5h)
	require.Zero(t, current.RequestUsage1d)
	require.Nil(t, current.RequestWindow5hStart)
	require.Nil(t, current.RequestWindow1dStart)
}

func TestRequestCountReservationCommitStartsSuccessfulWindow(t *testing.T) {
	ctx := context.Background()
	repo, user, group, apiKey := newRequestCountFixture(t, 2, 2)
	sub := mustCreateSubscription(t, repo.client, &service.UserSubscription{UserID: user.ID, GroupID: group.ID})
	reservation, err := repo.ReserveRequestCount(ctx, "request-commit", apiKey.ID, user.ID, group.ID, time.Now().Add(time.Minute))
	require.NoError(t, err)
	require.NotNil(t, reservation.Window5hStart)

	billingRepo := NewUsageBillingRepository(repo.client, integrationDB)
	result, err := billingRepo.Apply(ctx, &service.UsageBillingCommand{
		RequestID:            "request-commit",
		APIKeyID:             apiKey.ID,
		UserID:               user.ID,
		SubscriptionID:       &sub.ID,
		RequestReservationID: reservation.ID,
	})
	require.NoError(t, err)
	require.True(t, result.Applied)

	var status string
	require.NoError(t, integrationDB.QueryRowContext(ctx, "SELECT status FROM subscription_request_reservations WHERE id = $1", reservation.ID).Scan(&status))
	require.Equal(t, service.SubscriptionRequestReservationCommitted, status)
	current, err := repo.GetByID(ctx, sub.ID)
	require.NoError(t, err)
	require.Equal(t, 1, current.RequestUsage5h)
	require.Equal(t, 1, current.RequestUsage1d)
	require.NotNil(t, current.RequestWindow5hStart)
	require.False(t, current.RequestWindow5hStart.Before(*reservation.Window5hStart))
	require.NoError(t, repo.ReleaseRequestCount(ctx, reservation.ID))
	current, err = repo.GetByID(ctx, sub.ID)
	require.NoError(t, err)
	require.Equal(t, 1, current.RequestUsage5h)
}

func TestRequestCountReservationConcurrentLimit(t *testing.T) {
	ctx := context.Background()
	repo, user, group, apiKey := newRequestCountFixture(t, 2, 0)
	mustCreateSubscription(t, repo.client, &service.UserSubscription{UserID: user.ID, GroupID: group.ID})

	const requests = 6
	var wg sync.WaitGroup
	errs := make(chan error, requests)
	for i := 0; i < requests; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, err := repo.ReserveRequestCount(ctx, fmt.Sprintf("request-concurrent-%d", i), apiKey.ID, user.ID, group.ID, time.Now().Add(time.Minute))
			errs <- err
		}(i)
	}
	wg.Wait()
	close(errs)

	successes := 0
	exhausted := 0
	for err := range errs {
		switch {
		case err == nil:
			successes++
		case err == service.ErrRequestCountLimitExceeded:
			exhausted++
		default:
			require.NoError(t, err)
		}
	}
	require.Equal(t, 2, successes)
	require.Equal(t, requests-2, exhausted)
}

func TestRequestCountReservationUsesEarliestExpiringSubscription(t *testing.T) {
	ctx := context.Background()
	repo, user, group, apiKey := newRequestCountFixture(t, 1, 0)
	late := mustCreateSubscription(t, repo.client, &service.UserSubscription{UserID: user.ID, GroupID: group.ID, ExpiresAt: time.Now().Add(48 * time.Hour)})
	early := mustCreateSubscription(t, repo.client, &service.UserSubscription{UserID: user.ID, GroupID: group.ID, ExpiresAt: time.Now().Add(24 * time.Hour)})

	first, err := repo.ReserveRequestCount(ctx, "request-card-1", apiKey.ID, user.ID, group.ID, time.Now().Add(time.Minute))
	require.NoError(t, err)
	require.Equal(t, early.ID, first.SubscriptionID)
	second, err := repo.ReserveRequestCount(ctx, "request-card-2", apiKey.ID, user.ID, group.ID, time.Now().Add(time.Minute))
	require.NoError(t, err)
	require.Equal(t, late.ID, second.SubscriptionID)
}

func TestExpiredRequestCountReservationClearsUnusedWindow(t *testing.T) {
	ctx := context.Background()
	repo, user, group, apiKey := newRequestCountFixture(t, 2, 2)
	sub := mustCreateSubscription(t, repo.client, &service.UserSubscription{UserID: user.ID, GroupID: group.ID})
	reservation, err := repo.ReserveRequestCount(ctx, "request-expired", apiKey.ID, user.ID, group.ID, time.Now().Add(-time.Minute))
	require.NoError(t, err)

	_, err = repo.client.ExecContext(ctx, releaseExpiredRequestCountReservationsSQL, user.ID, group.ID)
	require.NoError(t, err)
	current, err := repo.GetByID(ctx, sub.ID)
	require.NoError(t, err)
	require.Zero(t, current.RequestUsage5h)
	require.Zero(t, current.RequestUsage1d)
	require.Nil(t, current.RequestWindow5hStart)
	require.Nil(t, current.RequestWindow1dStart)

	var status string
	require.NoError(t, integrationDB.QueryRowContext(ctx, "SELECT status FROM subscription_request_reservations WHERE id = $1", reservation.ID).Scan(&status))
	require.Equal(t, service.SubscriptionRequestReservationReleased, status)
}
