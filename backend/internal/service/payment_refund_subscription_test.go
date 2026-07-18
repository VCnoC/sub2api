package service

import (
	"context"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/stretchr/testify/require"
)

func TestPrepDeductUsesSubscriptionLinkedToOrder(t *testing.T) {
	groupID := int64(7)
	days := 30
	linkedID := int64(22)
	repo := newSubscriptionUserSubRepoStub()
	repo.seed(&UserSubscription{ID: 11, UserID: 5, GroupID: groupID, Status: SubscriptionStatusActive, ExpiresAt: time.Now().Add(24 * time.Hour)})
	repo.seed(&UserSubscription{ID: linkedID, UserID: 5, GroupID: groupID, Status: SubscriptionStatusActive, ExpiresAt: time.Now().Add(48 * time.Hour)})
	svc := &PaymentService{subscriptionSvc: NewSubscriptionService(nil, repo, nil, nil, nil)}
	order := &dbent.PaymentOrder{UserID: 5, OrderType: payment.OrderTypeSubscription, SubscriptionGroupID: &groupID, SubscriptionDays: &days, SubscriptionID: &linkedID}
	plan := &RefundPlan{}

	result := svc.prepDeduct(context.Background(), order, plan, false)

	require.Nil(t, result)
	require.Equal(t, linkedID, plan.SubscriptionID)
}

func TestPrepDeductRequiresForceForLegacySubscriptionOrder(t *testing.T) {
	groupID := int64(7)
	days := 30
	svc := &PaymentService{subscriptionSvc: NewSubscriptionService(nil, newSubscriptionUserSubRepoStub(), nil, nil, nil)}
	order := &dbent.PaymentOrder{UserID: 5, OrderType: payment.OrderTypeSubscription, SubscriptionGroupID: &groupID, SubscriptionDays: &days}

	result := svc.prepDeduct(context.Background(), order, &RefundPlan{}, false)

	require.NotNil(t, result)
	require.True(t, result.RequireForce)
}
