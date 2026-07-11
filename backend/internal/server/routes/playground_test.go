// Package routes 测试对话广场视频路由注册。
package routes

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/handler"
	servermiddleware "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestPlaygroundVideoRoutesAreRegistered(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	v1 := router.Group("/api/v1")
	RegisterPlaygroundRoutes(
		v1,
		&handler.Handlers{
			Gateway:                &handler.GatewayHandler{},
			OpenAIGateway:          &handler.OpenAIGatewayHandler{},
			Playground:             &handler.PlaygroundHandler{},
			PlaygroundConversation: &handler.PlaygroundConversationHandler{},
		},
		servermiddleware.JWTAuthMiddleware(func(c *gin.Context) { c.Next() }),
		nil,
		nil,
		nil,
		nil,
		nil,
		&config.Config{},
	)

	routes := make(map[string]bool)
	for _, route := range router.Routes() {
		routes[route.Method+" "+route.Path] = true
	}
	require.True(t, routes["POST /api/v1/playground/videos"])
	require.True(t, routes["GET /api/v1/playground/videos/:request_id"])
}
