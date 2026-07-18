package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

const apiKeyGroupFailoverContextKey = "api_key_group_failover"

type apiKeyGroupActivationError struct {
	Status  int
	Code    string
	Message string
}

func (e *apiKeyGroupActivationError) Error() string { return e.Message }

// APIKeyGroupFailoverState 保存单个请求的候选组进度，不修改 API Key 的持久化顺序。
type APIKeyGroupFailoverState struct {
	apiKey              *service.APIKey
	groups              []*service.Group
	current             int
	subscription        *service.UserSubscription
	subscriptionService *service.SubscriptionService
	cfg                 *config.Config
	skipBilling         bool
	loadSubscription    bool
}

func newAPIKeyGroupFailoverState(apiKey *service.APIKey, subscriptionService *service.SubscriptionService, cfg *config.Config, skipBilling, loadSubscription bool) *APIKeyGroupFailoverState {
	groups := apiKey.Groups
	if len(groups) == 0 {
		groups = []*service.Group{apiKey.Group}
	}
	return &APIKeyGroupFailoverState{
		apiKey: apiKey, groups: groups, current: -1,
		subscriptionService: subscriptionService, cfg: cfg,
		skipBilling: skipBilling, loadSubscription: loadSubscription,
	}
}

func (s *APIKeyGroupFailoverState) activateFrom(c *gin.Context, start int) error {
	var firstErr error
	for index := start; index < len(s.groups); index++ {
		subscription, err := s.checkCandidate(c, s.groups[index])
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		s.activate(c, index, subscription)
		return nil
	}
	if firstErr != nil {
		return firstErr
	}
	return &apiKeyGroupActivationError{Status: http.StatusForbidden, Code: "GROUP_UNAVAILABLE", Message: "No available group found for this API key"}
}

func (s *APIKeyGroupFailoverState) checkCandidate(c *gin.Context, group *service.Group) (*service.UserSubscription, error) {
	if group != nil {
		if strings.EqualFold(group.Status, "deleted") {
			return nil, &apiKeyGroupActivationError{Status: http.StatusForbidden, Code: "GROUP_DELETED", Message: "API Key 所属分组已删除"}
		}
		if !group.IsActive() {
			return nil, &apiKeyGroupActivationError{Status: http.StatusForbidden, Code: "GROUP_DISABLED", Message: "API Key 所属分组已停用"}
		}
		if !group.IsSubscriptionType() && s.apiKey.User != nil && !s.apiKey.User.CanBindGroup(group.ID, group.IsExclusive) {
			return nil, &apiKeyGroupActivationError{Status: http.StatusForbidden, Code: "GROUP_NOT_ALLOWED", Message: "API Key 所属专属分组不再允许当前用户使用"}
		}
	}

	var subscription *service.UserSubscription
	if group != nil && group.IsSubscriptionType() && s.subscriptionService != nil && s.loadSubscription {
		sub, err := s.subscriptionService.GetActiveSubscription(c.Request.Context(), s.apiKey.User.ID, group.ID)
		if err != nil {
			if s.skipBilling {
				return nil, nil
			}
			if service.IsSubscriptionLimitError(err) {
				return nil, &apiKeyGroupActivationError{Status: http.StatusTooManyRequests, Code: "USAGE_LIMIT_EXCEEDED", Message: err.Error()}
			}
			return nil, &apiKeyGroupActivationError{Status: http.StatusForbidden, Code: "SUBSCRIPTION_NOT_FOUND", Message: "No active subscription found for this group"}
		}
		subscription = sub
	}

	if !s.skipBilling && subscription == nil && apiKeyBalanceBelowAuthThreshold(s.apiKey.User.Balance, s.cfg) {
		return nil, &apiKeyGroupActivationError{Status: http.StatusForbidden, Code: "INSUFFICIENT_BALANCE", Message: "Insufficient account balance"}
	}
	return subscription, nil
}

func (s *APIKeyGroupFailoverState) activate(c *gin.Context, index int, subscription *service.UserSubscription) {
	group := s.groups[index]
	s.current = index
	s.subscription = subscription
	s.apiKey.Group = group
	s.apiKey.GroupID = nil
	if group != nil {
		groupID := group.ID
		s.apiKey.GroupID = &groupID
		s.apiKey.User.UserGroupRPMOverride = s.apiKey.GroupRPMOverrides[groupID]
		setGroupContext(c, group)
	} else {
		s.apiKey.User.UserGroupRPMOverride = nil
	}
	if subscription != nil {
		c.Set(string(ContextKeySubscription), subscription)
	} else {
		c.Set(string(ContextKeySubscription), nil)
	}
	c.Set(string(ContextKeyAPIKey), s.apiKey)
}

func (s *APIKeyGroupFailoverState) Subscription() *service.UserSubscription { return s.subscription }

// AdvanceAPIKeyGroup 激活下一个具备基础计费资格的候选组。
func AdvanceAPIKeyGroup(c *gin.Context) bool {
	if c == nil || c.Request == nil || c.Request.Context().Err() != nil || c.Writer.Written() {
		return false
	}
	value, ok := c.Get(apiKeyGroupFailoverContextKey)
	if !ok {
		return false
	}
	state, ok := value.(*APIKeyGroupFailoverState)
	if !ok || state == nil || state.current < 0 {
		return false
	}
	return state.activateFrom(c, state.current+1) == nil
}

func abortAPIKeyGroupActivation(c *gin.Context, err error) {
	var activationErr *apiKeyGroupActivationError
	if errors.As(err, &activationErr) {
		service.MarkOpsClientBusinessLimited(c, service.OpsClientBusinessLimitedReasonAPIKeyGroupUnavailable)
		AbortWithError(c, activationErr.Status, activationErr.Code, activationErr.Message)
		return
	}
	AbortWithError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to validate API key group")
}
