package handler

import (
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/claude"
	"github.com/Wei-Shaw/sub2api/internal/pkg/geminicli"
	"github.com/Wei-Shaw/sub2api/internal/pkg/openai"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// PlaygroundHandler 对话广场专用 handler。
//
// 设计原则：handler 极简 —— Chat Completions 转发由 PlaygroundContextMiddleware
// 注入虚拟 APIKey 后直接 dispatch 给现有 Gateway / OpenAIGateway handler 复用整套
// 计费/限流/记账链路，故此处仅提供模型查询端点。
type PlaygroundHandler struct {
	apiKeyService  *service.APIKeyService
	gatewayService *service.GatewayService
}

// NewPlaygroundHandler 构造 PlaygroundHandler
func NewPlaygroundHandler(
	apiKeyService *service.APIKeyService,
	gatewayService *service.GatewayService,
) *PlaygroundHandler {
	return &PlaygroundHandler{
		apiKeyService:  apiKeyService,
		gatewayService: gatewayService,
	}
}

// AvailableModels 返回指定分组下用户可用的模型列表。
//
// GET /api/v1/playground/models?group={groupName}
//
// 行为：
//  1. 校验用户已认证（jwtAuth 已注入 subject）
//  2. 校验 group 属于用户可用分组列表（防越权暴露其他分组的模型）
//  3. 通过 gatewayService.GetAvailableModels 查询白名单模型
//  4. 若白名单为空，回退到 platform 默认模型集合
func (h *PlaygroundHandler) AvailableModels(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	groupName := strings.TrimSpace(c.Query("group"))
	if groupName == "" {
		response.BadRequest(c, "group is required")
		return
	}

	ctx := c.Request.Context()

	availableGroups, err := h.apiKeyService.GetAvailableGroups(ctx, subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	var targetGroup *service.Group
	for i := range availableGroups {
		if availableGroups[i].Name == groupName {
			targetGroup = &availableGroups[i]
			break
		}
	}
	if targetGroup == nil {
		response.Forbidden(c, "Selected group is not available for current user")
		return
	}

	// 1. 优先从账号白名单读
	whitelist := h.gatewayService.GetAvailableModels(ctx, &targetGroup.ID, targetGroup.Platform)

	modelIDs := whitelist
	if len(modelIDs) == 0 {
		// 2. Fallback 到平台默认模型列表
		modelIDs = defaultModelsForPlatform(targetGroup.Platform)
	}

	result := make([]dto.PlaygroundAvailableModel, 0, len(modelIDs))
	for _, modelID := range modelIDs {
		modelID = strings.TrimSpace(modelID)
		if modelID == "" {
			continue
		}
		result = append(result, dto.PlaygroundAvailableModel{
			ID:        modelID,
			Platform:  targetGroup.Platform,
			GroupID:   targetGroup.ID,
			GroupName: targetGroup.Name,
		})
	}

	response.Success(c, dto.PlaygroundAvailableModelsResponse{Models: result})
}

// defaultModelsForPlatform 返回指定平台的默认模型 ID 列表。
// 与现有 GatewayHandler.Models 的 fallback 行为保持一致。
func defaultModelsForPlatform(platform string) []string {
	switch platform {
	case service.PlatformOpenAI:
		ids := make([]string, 0, len(openai.DefaultModels))
		for _, m := range openai.DefaultModels {
			ids = append(ids, m.ID)
		}
		return ids
	case service.PlatformGemini:
		ids := make([]string, 0, len(geminicli.DefaultModels))
		for _, m := range geminicli.DefaultModels {
			ids = append(ids, m.ID)
		}
		return ids
	default:
		// Anthropic / Antigravity 等回退到 claude 默认列表
		ids := make([]string, 0, len(claude.DefaultModels))
		for _, m := range claude.DefaultModels {
			ids = append(ids, m.ID)
		}
		return ids
	}
}
