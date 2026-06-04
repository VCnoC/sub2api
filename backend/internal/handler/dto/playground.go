package dto

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
