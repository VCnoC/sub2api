// Package middleware 测试对话广场请求上下文与余额边界。
package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type playgroundUserRepositoryStub struct {
	service.UserRepository
	user *service.User
}

func (s *playgroundUserRepositoryStub) GetByID(context.Context, int64) (*service.User, error) {
	return s.user, nil
}

func (s *playgroundUserRepositoryStub) GetUserAvatar(context.Context, int64) (*service.UserAvatar, error) {
	return nil, nil
}

type playgroundGroupRepositoryStub struct {
	service.GroupRepository
	group *service.Group
}

func (s *playgroundGroupRepositoryStub) GetByID(context.Context, int64) (*service.Group, error) {
	return s.group, nil
}

func (s *playgroundGroupRepositoryStub) ListActive(context.Context) ([]service.Group, error) {
	return []service.Group{*s.group}, nil
}

type playgroundSubscriptionRepositoryStub struct {
	service.UserSubscriptionRepository
}

func (s *playgroundSubscriptionRepositoryStub) ListActiveByUserID(context.Context, int64) ([]service.UserSubscription, error) {
	return nil, nil
}

type playgroundAPIKeyRepositoryStub struct {
	service.APIKeyRepository
	key service.APIKey
}

func (s *playgroundAPIKeyRepositoryStub) SearchAPIKeys(context.Context, int64, string, int) ([]service.APIKey, error) {
	return []service.APIKey{s.key}, nil
}

func TestPlaygroundContextAllowsZeroBalanceVideoStatusQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	groupID := int64(7)
	user := &service.User{ID: 11, Status: service.StatusActive, Balance: 0}
	group := &service.Group{
		ID:               groupID,
		Name:             "video",
		Platform:         service.PlatformVideo,
		Status:           service.StatusActive,
		Hydrated:         true,
		SubscriptionType: service.SubscriptionTypeStandard,
	}
	userRepo := &playgroundUserRepositoryStub{user: user}
	apiKeyService := service.NewAPIKeyService(
		&playgroundAPIKeyRepositoryStub{key: service.APIKey{
			ID:      17,
			UserID:  user.ID,
			Name:    service.PlaygroundInternalKeyName,
			GroupID: &groupID,
			Status:  service.StatusActive,
		}},
		userRepo,
		&playgroundGroupRepositoryStub{group: group},
		&playgroundSubscriptionRepositoryStub{},
		nil,
		nil,
		&config.Config{},
	)
	userService := service.NewUserService(userRepo, nil, nil, nil)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(string(ContextKeyUser), AuthSubject{UserID: user.ID})
		c.Next()
	})
	router.Use(NewPlaygroundContextMiddleware(apiKeyService, userService, nil))
	router.Any("/videos/:request_id", func(c *gin.Context) {
		_, ok := GetAPIKeyFromContext(c)
		require.True(t, ok)
		c.Status(http.StatusNoContent)
	})

	statusRequest := httptest.NewRequest(http.MethodGet, "/videos/task-1?group=video", nil)
	statusResponse := httptest.NewRecorder()
	router.ServeHTTP(statusResponse, statusRequest)
	require.Equal(t, http.StatusNoContent, statusResponse.Code)

	createRequest := httptest.NewRequest(
		http.MethodPost,
		"/videos/task-1",
		strings.NewReader(`{"model":"grok-imagine-video","group":"video"}`),
	)
	createRequest.Header.Set("Content-Type", "application/json")
	createResponse := httptest.NewRecorder()
	router.ServeHTTP(createResponse, createRequest)
	require.Equal(t, http.StatusForbidden, createResponse.Code)
	require.Contains(t, createResponse.Body.String(), "insufficient_quota")
}
