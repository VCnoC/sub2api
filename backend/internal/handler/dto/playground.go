package dto

import (
	"encoding/json"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

// PlaygroundChatRequest 对话广场聊天请求的关键字段提取结构。
// 仅用于参数校验与中间件分发，其余 OpenAI 标准字段（temperature/top_p/...）
// 由上游 ChatCompletions handler 通过 gjson 直接读取请求体处理，无需在此声明。
//
// 注意：本结构与 OpenAI ChatCompletion 请求体共存于同一 JSON，
// 反序列化时未声明字段会被忽略，原始请求体仍可完整透传到上游。
type PlaygroundChatRequest struct {
	// Model 用户选择的模型 ID（必填）
	Model string `json:"model"`
	// Group 用户选择的分组名（必填，需属于用户可用分组列表）
	Group string `json:"group"`
}

// PlaygroundAvailableModel 对话广场可用模型条目
type PlaygroundAvailableModel struct {
	// ID 模型唯一标识（如 "claude-sonnet-4-20250514"）
	ID string `json:"id"`
	// Platform 模型所属平台（"openai" / "anthropic" / "gemini" / "antigravity"）
	Platform string `json:"platform"`
	// GroupID 模型来源分组 ID
	GroupID int64 `json:"group_id"`
	// GroupName 模型来源分组名
	GroupName string `json:"group_name"`
}

// PlaygroundAvailableModelsResponse 可用模型列表响应
type PlaygroundAvailableModelsResponse struct {
	// Models 当前查询分组下用户可用的模型列表
	Models []PlaygroundAvailableModel `json:"models"`
}

// ---- 会话 CRUD DTO ----

// PlaygroundConversationSummaryDTO 会话列表条目（不含 messages 大字段）。
// 用于 GET /conversations 列表接口。
type PlaygroundConversationSummaryDTO struct {
	// ID 会话主键
	ID int64 `json:"id"`
	// Title 会话标题
	Title string `json:"title"`
	// Model 最后使用的模型标识（可选）
	Model *string `json:"model,omitempty"`
	// GroupName 所属分组名称（可选）
	GroupName *string `json:"group_name,omitempty"`
	// LastActivityAt 最后活动时间
	LastActivityAt time.Time `json:"last_activity_at"`
	// CreatedAt 创建时间
	CreatedAt time.Time `json:"created_at"`
}

// PlaygroundConversationDetailDTO 会话详情（含 messages）。
// 用于 GET /conversations/:id 接口。
type PlaygroundConversationDetailDTO struct {
	// ID 会话主键
	ID int64 `json:"id"`
	// Title 会话标题
	Title string `json:"title"`
	// Model 最后使用的模型标识（可选）
	Model *string `json:"model,omitempty"`
	// GroupName 所属分组名称（可选）
	GroupName *string `json:"group_name,omitempty"`
	// Messages 完整消息列表，JSONB 原样透传
	Messages json.RawMessage `json:"messages"`
	// LastActivityAt 最后活动时间
	LastActivityAt time.Time `json:"last_activity_at"`
	// CreatedAt 创建时间
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt 更新时间
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateConversationRequest 新建会话请求体。
type CreateConversationRequest struct {
	// Title 会话标题（必填）
	Title string `json:"title"`
	// Model 使用的模型 ID（可选）
	Model *string `json:"model,omitempty"`
	// GroupName 所属分组名称（可选）
	GroupName *string `json:"group_name,omitempty"`
	// Messages 初始消息列表（可选，JSONB 原样存储）
	Messages json.RawMessage `json:"messages,omitempty"`
}

// UpdateConversationRequest 更新会话请求体，所有字段均可选（nil 表示不改）。
// model/group_name 传 null 表示清空，不传（字段缺失）与传 null 语义相同。
type UpdateConversationRequest struct {
	// Title 会话标题（可选，传入则更新）
	Title *string `json:"title,omitempty"`
	// Model 模型标识（可选，传入则更新，传 null 则清空）
	Model *string `json:"model"`
	// GroupName 分组名称（可选，传入则更新，传 null 则清空）
	GroupName *string `json:"group_name"`
	// Messages 完整消息列表（可选，传入则覆盖更新）
	Messages json.RawMessage `json:"messages,omitempty"`
}

// ConversationSummaryFromService 将 service 摘要结构转换为 DTO。
func ConversationSummaryFromService(s *service.PlaygroundConversationSummary) *PlaygroundConversationSummaryDTO {
	if s == nil {
		return nil
	}
	return &PlaygroundConversationSummaryDTO{
		ID:             s.ID,
		Title:          s.Title,
		Model:          s.Model,
		GroupName:      s.GroupName,
		LastActivityAt: s.LastActivityAt,
		CreatedAt:      s.CreatedAt,
	}
}

// ConversationDetailFromService 将 service 完整会话结构转换为 DTO。
func ConversationDetailFromService(c *service.PlaygroundConversation) *PlaygroundConversationDetailDTO {
	if c == nil {
		return nil
	}
	return &PlaygroundConversationDetailDTO{
		ID:             c.ID,
		Title:          c.Title,
		Model:          c.Model,
		GroupName:      c.GroupName,
		Messages:       c.Messages,
		LastActivityAt: c.LastActivityAt,
		CreatedAt:      c.CreatedAt,
		UpdatedAt:      c.UpdatedAt,
	}
}
