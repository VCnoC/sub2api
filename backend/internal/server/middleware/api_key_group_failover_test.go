package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

func TestAPIKeyGroupFailoverStateUsesOrderedAvailableGroups(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/v1/messages", nil)

	groups := []*service.Group{
		{ID: 1, Platform: service.PlatformAnthropic, Status: "inactive", SubscriptionType: service.SubscriptionTypeStandard},
		{ID: 2, Platform: service.PlatformAnthropic, Status: service.StatusActive, SubscriptionType: service.SubscriptionTypeStandard},
		{ID: 3, Platform: service.PlatformAnthropic, Status: service.StatusActive, SubscriptionType: service.SubscriptionTypeStandard},
	}
	apiKey := &service.APIKey{
		User:   &service.User{ID: 7, Status: service.StatusActive, Balance: 10},
		Groups: groups, GroupIDs: []int64{1, 2, 3},
	}
	state := newAPIKeyGroupFailoverState(apiKey, nil, &config.Config{}, false, true)
	if err := state.activateFrom(c, 0); err != nil {
		t.Fatalf("activateFrom() error = %v", err)
	}
	if apiKey.GroupID == nil || *apiKey.GroupID != 2 {
		t.Fatalf("active group = %v, want 2", apiKey.GroupID)
	}
	c.Set(apiKeyGroupFailoverContextKey, state)
	if !AdvanceAPIKeyGroup(c) {
		t.Fatal("AdvanceAPIKeyGroup() = false, want true")
	}
	if apiKey.GroupID == nil || *apiKey.GroupID != 3 {
		t.Fatalf("advanced group = %v, want 3", apiKey.GroupID)
	}
}

func TestAdvanceAPIKeyGroupStopsAfterResponseWrite(t *testing.T) {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest("POST", "/v1/messages", nil)
	apiKey := &service.APIKey{
		User: &service.User{ID: 7, Status: service.StatusActive, Balance: 10},
		Groups: []*service.Group{
			{ID: 1, Platform: service.PlatformAnthropic, Status: service.StatusActive, SubscriptionType: service.SubscriptionTypeStandard},
			{ID: 2, Platform: service.PlatformAnthropic, Status: service.StatusActive, SubscriptionType: service.SubscriptionTypeStandard},
		},
	}
	state := newAPIKeyGroupFailoverState(apiKey, nil, &config.Config{}, false, true)
	if err := state.activateFrom(c, 0); err != nil {
		t.Fatal(err)
	}
	c.Set(apiKeyGroupFailoverContextKey, state)
	c.String(200, "started")
	if AdvanceAPIKeyGroup(c) {
		t.Fatal("AdvanceAPIKeyGroup() = true after response write")
	}
}
