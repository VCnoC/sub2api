// Package service 定义对话广场会话的领域类型、Repository 接口与错误变量。
//
// 会话数据采用会话级 JSONB 整存方案，后端透明存储不解析消息结构。
// 所有访问均需携带 userID 条件，防止越权读取他人会话。
package service

import (
	"context"
	"encoding/json"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

// 单用户会话数量上限与单会话消息体积上限。
const (
	// PlaygroundConversationMaxPerUser 单用户最多创建的会话数量（防滥用）。
	PlaygroundConversationMaxPerUser = 50

	// PlaygroundConversationMaxMessagesBytes 单会话 messages 字段的最大字节数（50MB）。
	// 超过此限制时，Create/Update 操作返回 ErrPlaygroundConversationTooLarge。
	PlaygroundConversationMaxMessagesBytes = 50 << 20 // 50 MB
)

// 错误变量，参照 domain/announcement.go 中的惯例：
// NotFound 类型的错误定义在 domain 层，BadRequest 类型可直接定义在 service 层。
var (
	// ErrPlaygroundConversationNotFound 会话不存在或不属于当前用户（404）。
	ErrPlaygroundConversationNotFound = infraerrors.NotFound(
		"PLAYGROUND_CONVERSATION_NOT_FOUND",
		"playground conversation not found",
	)

	// ErrPlaygroundConversationLimitExceeded 单用户会话数已达上限（400）。
	ErrPlaygroundConversationLimitExceeded = infraerrors.BadRequest(
		"PLAYGROUND_CONVERSATION_LIMIT_EXCEEDED",
		"playground conversation limit exceeded",
	)

	// ErrPlaygroundConversationTooLarge 单会话 messages 超过体积上限（400）。
	ErrPlaygroundConversationTooLarge = infraerrors.BadRequest(
		"PLAYGROUND_CONVERSATION_TOO_LARGE",
		"playground conversation messages too large",
	)
)

// PlaygroundConversation 是对话广场会话的领域结构，包含完整的消息列表。
// 用于 GetByID、Create、Update 等需要读写消息内容的场景。
type PlaygroundConversation struct {
	// ID 主键，由数据库自动分配
	ID int64

	// UserID 会话所属用户，所有操作必须校验此字段（越权防护）
	UserID int64

	// Title 会话标题，默认取首条用户消息前 20 字，支持手动改名
	Title string

	// Model 最后使用的模型标识（可选，如 "claude-3-5-sonnet"）
	Model *string

	// GroupName 所属分组名称（可选，前端分类展示用）
	GroupName *string

	// Messages 完整消息列表，JSONB 整存，后端透明不解析结构
	Messages json.RawMessage

	// LastActivityAt 最后活动时间，每次保存消息时刷新；过期清理依据此字段
	LastActivityAt time.Time

	// CreatedAt 创建时间（由 TimeMixin 自动填充）
	CreatedAt time.Time

	// UpdatedAt 更新时间（由 TimeMixin 自动维护）
	UpdatedAt time.Time
}

// PlaygroundConversationSummary 是会话摘要结构，用于列表查询。
// 不含 Messages 大字段，避免列表查询拉取冗余数据（参见 NFR-001 R-001）。
type PlaygroundConversationSummary struct {
	// ID 主键
	ID int64

	// Title 会话标题
	Title string

	// Model 最后使用的模型标识（可选）
	Model *string

	// GroupName 所属分组名称（可选）
	GroupName *string

	// LastActivityAt 最后活动时间（用于列表排序）
	LastActivityAt time.Time

	// CreatedAt 创建时间
	CreatedAt time.Time

	// UpdatedAt 更新时间
	UpdatedAt time.Time
}

// PlaygroundConversationRepository 定义对话广场会话的持久化接口。
//
// 设计约束：
//   - 所有涉及单条会话的操作均需同时传入 userID，由 Repository 层在 SQL 中附加
//     `WHERE user_id = userID` 条件，作为越权防护的双保险。
//   - ListByUser 返回摘要列表（不含 messages），按 last_activity_at 倒序。
//   - DeleteExpired 为批量清理接口，由定时任务（ConversationCleanupService）调用。
type PlaygroundConversationRepository interface {
	// ListByUser 返回指定用户的所有会话摘要，按 last_activity_at 倒序排列。
	// 不返回 messages 字段（大字段，列表场景不需要）。
	ListByUser(ctx context.Context, userID int64) ([]PlaygroundConversationSummary, error)

	// GetByID 按 ID 获取会话完整数据（含 messages）。
	// 必须同时匹配 userID，防止越权访问；不存在时返回 ErrPlaygroundConversationNotFound。
	GetByID(ctx context.Context, id, userID int64) (*PlaygroundConversation, error)

	// Create 创建新会话。创建成功后，c.ID / c.CreatedAt / c.UpdatedAt 会被回填。
	Create(ctx context.Context, c *PlaygroundConversation) error

	// Update 更新会话的 title/model/group_name/messages/last_activity_at 字段。
	// 必须同时匹配 userID；不存在时返回 ErrPlaygroundConversationNotFound。
	Update(ctx context.Context, c *PlaygroundConversation) error

	// Delete 删除指定会话。必须同时匹配 userID；不存在时返回 ErrPlaygroundConversationNotFound。
	Delete(ctx context.Context, id, userID int64) error

	// CountByUser 返回指定用户当前的会话数量，用于创建前的上限校验。
	CountByUser(ctx context.Context, userID int64) (int, error)

	// DeleteExpired 物理删除 last_activity_at < before 的所有会话（过期清理）。
	// batchSize 为每次删除的最大条数（调用方可分批调用以控制锁粒度）。
	// 返回本次实际删除的记录数。
	DeleteExpired(ctx context.Context, before time.Time, batchSize int) (int, error)
}
