package handler

import (
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

// advanceAPIKeyGroup 推进候选组，并用现有计费服务重新检查该组资格。
func (h *GatewayHandler) advanceAPIKeyGroup(c *gin.Context, state *FailoverState) (*service.APIKey, *service.UserSubscription, bool) {
	return advanceAPIKeyGroupWithBilling(c, state, h.billingCacheService)
}

func (h *OpenAIGatewayHandler) advanceAPIKeyGroup(c *gin.Context, state *FailoverState) (*service.APIKey, *service.UserSubscription, bool) {
	return advanceAPIKeyGroupWithBilling(c, state, h.billingCacheService)
}

func advanceAPIKeyGroupWithBilling(c *gin.Context, state *FailoverState, billingCacheService *service.BillingCacheService) (*service.APIKey, *service.UserSubscription, bool) {
	for state.AdvanceGroup(c) {
		apiKey, ok := middleware.GetAPIKeyFromContext(c)
		if !ok || apiKey == nil {
			return nil, nil, false
		}
		if apiKey.Group != nil && apiKey.Group.ClaudeCodeOnly && c.Request.URL.Path != "/v1/messages" {
			continue
		}
		subscription, _ := middleware.GetSubscriptionFromContext(c)
		if billingCacheService == nil || billingCacheService.CheckBillingEligibility(
			c.Request.Context(), apiKey.User, apiKey, apiKey.Group, subscription, service.QuotaPlatform(c.Request.Context(), apiKey),
		) == nil {
			return apiKey, subscription, true
		}
	}
	return nil, nil, false
}
