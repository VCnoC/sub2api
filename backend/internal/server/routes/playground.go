package routes

import (
	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// RegisterPlaygroundRoutes 注册对话广场（Playground）路由。
//
// 路由分组：/api/v1/playground
//   - GET  /models                 列出指定分组下用户可用的模型
//   - POST /chat/completions       OpenAI Chat Completions 兼容转发
//
// 鉴权：JWT（用户 Web 会话），区别于 /v1/chat/completions 的 API Key 鉴权。
//
// 转发链路：jwtAuth → BackendModeUserGuard → PlaygroundContextMiddleware
//
//	→ bodyLimit/clientRequestID/opsErrorLogger/endpointNorm
//	→ 按 groupPlatform dispatch 到现有 OpenAIGateway / Gateway handler
//
// 计费/限流/记账：100% 复用现有 chat completions handler 的内置链路
// （billingCacheService → submitUsageRecordTask → BalanceNotifyService）。
func RegisterPlaygroundRoutes(
	v1 *gin.RouterGroup,
	h *handler.Handlers,
	jwtAuth middleware.JWTAuthMiddleware,
	apiKeyService *service.APIKeyService,
	userService *service.UserService,
	subscriptionService *service.SubscriptionService,
	opsService *service.OpsService,
	settingService *service.SettingService,
	cfg *config.Config,
) {
	pg := v1.Group("/playground")
	pg.Use(gin.HandlerFunc(jwtAuth))
	pg.Use(middleware.BackendModeUserGuard(settingService))
	{
		// 用户可用模型查询（按 group）
		pg.GET("/models", h.Playground.AvailableModels)

		// Chat Completions 转发：复用现有 OpenAI/Anthropic gateway 链路
		// 通过 PlaygroundContextMiddleware 注入虚拟 APIKey + Group + Subscription 上下文
		chat := pg.Group("")
		chat.Use(middleware.RequestBodyLimit(cfg.Gateway.MaxBodySize))
		chat.Use(middleware.ClientRequestID())
		chat.Use(handler.OpsErrorLoggerMiddleware(opsService))
		chat.Use(handler.InboundEndpointMiddleware())
		chat.Use(middleware.NewPlaygroundContextMiddleware(apiKeyService, userService, subscriptionService))
		{
			chat.POST("/chat/completions", func(c *gin.Context) {
				// 按 Group.Platform 自动路由到 OpenAI 或 Anthropic gateway handler
				if getGroupPlatform(c) == service.PlatformOpenAI {
					h.OpenAIGateway.ChatCompletions(c)
					return
				}
				h.Gateway.ChatCompletions(c)
			})
		}
	}
}
