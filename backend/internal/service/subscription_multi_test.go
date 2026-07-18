package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type multiSubscriptionCandidateRepo struct {
	userSubRepoNoop
	subs []UserSubscription
}

func (r *multiSubscriptionCandidateRepo) ListActiveByUserID(context.Context, int64) ([]UserSubscription, error) {
	return append([]UserSubscription(nil), r.subs...), nil
}

func TestIssueSubscriptionCreatesIndependentEntitlement(t *testing.T) {
	groupRepo := &subscriptionGroupRepoStub{group: &Group{ID: 1, SubscriptionType: SubscriptionTypeSubscription}}
	subRepo := newSubscriptionUserSubRepoStub()
	subRepo.seed(&UserSubscription{ID: 10, UserID: 7, GroupID: 1, StartsAt: time.Now(), ExpiresAt: time.Now().Add(24 * time.Hour)})
	svc := NewSubscriptionService(groupRepo, subRepo, nil, nil, nil)

	issued, err := svc.IssueSubscription(context.Background(), &AssignSubscriptionInput{UserID: 7, GroupID: 1, ValidityDays: 1})

	require.NoError(t, err)
	require.NotEqual(t, int64(10), issued.ID)
	require.Equal(t, 1, subRepo.createCalls)
	require.Len(t, subRepo.byID, 2)
}

func TestGetActiveSubscriptionSkipsExhaustedEarlierEntitlement(t *testing.T) {
	now := time.Now()
	windowStart := now.Add(-time.Hour)
	limit := 10.0
	group := &Group{ID: 1, SubscriptionType: SubscriptionTypeSubscription, DailyLimitUSD: &limit}
	repo := &multiSubscriptionCandidateRepo{subs: []UserSubscription{
		{ID: 1, UserID: 7, GroupID: 1, Status: SubscriptionStatusActive, ExpiresAt: now.Add(24 * time.Hour), DailyWindowStart: &windowStart, DailyUsageUSD: limit},
		{ID: 2, UserID: 7, GroupID: 1, Status: SubscriptionStatusActive, ExpiresAt: now.Add(48 * time.Hour), DailyWindowStart: &windowStart, DailyUsageUSD: 1},
	}}
	svc := NewSubscriptionService(&subscriptionGroupRepoStub{group: group}, repo, nil, nil, nil)

	sub, err := svc.GetActiveSubscription(context.Background(), 7, 1)

	require.NoError(t, err)
	require.Equal(t, int64(2), sub.ID)
}

func TestGetActiveSubscriptionReturnsLimitErrorWhenAllEntitlementsExhausted(t *testing.T) {
	now := time.Now()
	windowStart := now.Add(-time.Hour)
	limit := 10.0
	group := &Group{ID: 1, SubscriptionType: SubscriptionTypeSubscription, DailyLimitUSD: &limit}
	repo := &multiSubscriptionCandidateRepo{subs: []UserSubscription{
		{ID: 1, UserID: 7, GroupID: 1, Status: SubscriptionStatusActive, ExpiresAt: now.Add(24 * time.Hour), DailyWindowStart: &windowStart, DailyUsageUSD: limit},
		{ID: 2, UserID: 7, GroupID: 1, Status: SubscriptionStatusActive, ExpiresAt: now.Add(48 * time.Hour), DailyWindowStart: &windowStart, DailyUsageUSD: limit + 1},
	}}
	svc := NewSubscriptionService(&subscriptionGroupRepoStub{group: group}, repo, nil, nil, nil)

	_, err := svc.GetActiveSubscription(context.Background(), 7, 1)

	require.True(t, errors.Is(err, ErrDailyLimitExceeded))
}

func TestReduceSubscriptionConsumesEarliestEntitlementsFirst(t *testing.T) {
	now := time.Now()
	repo := newSubscriptionUserSubRepoStub()
	repo.seed(&UserSubscription{ID: 1, UserID: 7, GroupID: 1, Status: SubscriptionStatusActive, ExpiresAt: now.Add(24 * time.Hour)})
	repo.seed(&UserSubscription{ID: 2, UserID: 7, GroupID: 1, Status: SubscriptionStatusActive, ExpiresAt: now.Add(72 * time.Hour)})
	subscriptionSvc := NewSubscriptionService(nil, repo, nil, nil, nil)
	redeemSvc := &RedeemService{subscriptionService: subscriptionSvc}

	err := redeemSvc.reduceOrCancelSubscription(context.Background(), 7, 1, 2, "REFUND")

	require.NoError(t, err)
	require.Equal(t, SubscriptionStatusExpired, repo.byID[1].Status)
	require.WithinDuration(t, now, repo.byID[1].ExpiresAt, time.Second)
	require.WithinDuration(t, now.Add(48*time.Hour), repo.byID[2].ExpiresAt, time.Second)
}
