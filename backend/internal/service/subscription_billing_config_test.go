package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeSubscriptionBillingConfig_RequestCount(t *testing.T) {
	mode, limit5h, limit1d, err := NormalizeSubscriptionBillingConfig(
		SubscriptionTypeSubscription,
		SubscriptionBillingModeRequestCount,
		20,
		50,
	)
	require.NoError(t, err)
	require.Equal(t, SubscriptionBillingModeRequestCount, mode)
	require.Equal(t, 20, limit5h)
	require.Equal(t, 50, limit1d)
}

func TestNormalizeSubscriptionBillingConfig_RejectsEmptyRequestLimits(t *testing.T) {
	_, _, _, err := NormalizeSubscriptionBillingConfig(
		SubscriptionTypeSubscription,
		SubscriptionBillingModeRequestCount,
		0,
		0,
	)
	require.Error(t, err)
}
