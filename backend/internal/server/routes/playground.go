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
//   - GET  /models                           列出指定分组下用户可用的模型
//   - POST /chat/completions                 OpenAI Chat Completions 兼容转发
//   - POST /videos                          创建视频任务
//   - GET  /videos/:request_id              查询视频任务
//   - GET  /conversations                    列出当前用户的所有会话摘要
//   - POST /conversations                    新建会话
//   - GET  /conversations/:id               获取会话详情（含 messages）
//   - PUT  /conversations/:id               更新会话（全量覆盖 messages，body 上限 50MB）
//   - DELETE /conversations/:id             删除会话
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
			chat.POST("/videos", func(c *gin.Context) {
				if getGroupPlatform(c) != service.PlatformVideo {
					playgroundVideoNotSupported(c)
					return
				}
				h.OpenAIGateway.GrokVideoGeneration(c)
			})
			chat.GET("/videos/:request_id", func(c *gin.Context) {
				if getGroupPlatform(c) != service.PlatformVideo {
					playgroundVideoNotSupported(c)
					return
				}
				h.OpenAIGateway.GrokVideoStatus(c)
			})
		}

		// 会话 CRUD：使用独立子组并放宽 body 上限至 52MB（messages 上限 50MB + 包装开销）
		// 注意：不复用 chat 子组，避免引入 OpsErrorLogger/EndpointMiddleware 等无关中间件
		conversations := pg.Group("/conversations")
		conversations.Use(middleware.RequestBodyLimit(52 << 20)) // 52MB
		{
			conversations.GET("", h.PlaygroundConversation.ListConversations)
			conversations.POST("", h.PlaygroundConversation.CreateConversation)
			conversations.GET("/:id", h.PlaygroundConversation.GetConversation)
			conversations.PUT("/:id", h.PlaygroundConversation.UpdateConversation)
			conversations.DELETE("/:id", h.PlaygroundConversation.DeleteConversation)
		}
	}
}

func playgroundVideoNotSupported(c *gin.Context) {
	service.MarkOpsClientBusinessLimited(c, service.OpsClientBusinessLimitedReasonLocalFeatureGate)
	c.JSON(404, gin.H{
		"error": gin.H{
			"type":    "not_found_error",
			"message": "Videos API is only supported for video groups",
		},
	})
}
