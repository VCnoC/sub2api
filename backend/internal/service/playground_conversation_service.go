// Package service 提供对话广场会话的业务逻辑实现。
//
// 本文件实现 PlaygroundConversationService，封装会话 CRUD 的业务规则：
//   - 单用户会话数上限（PlaygroundConversationMaxPerUser）
//   - 单会话 messages 体积上限（PlaygroundConversationMaxMessagesBytes）
//   - Update 时刷新 last_activity_at
//   - 越权防护由 Repository 层 SQL WHERE user_id=? 双保险
package service

import (
	"context"
	"encoding/json"
	"strings"
	"time"
)

// PlaygroundConversationService 提供对话广场会话的业务逻辑。
// 所有操作均需传入 userID，由 service 层做业务校验，同时 repository 层在 SQL 中附加 user_id 条件作为双保险。
type PlaygroundConversationService struct {
	repo PlaygroundConversationRepository
}

// NewPlaygroundConversationService 创建 PlaygroundConversationService 实例。
func NewPlaygroundConversationService(repo PlaygroundConversationRepository) *PlaygroundConversationService {
	return &PlaygroundConversationService{repo: repo}
}

// List 返回指定用户的所有会话摘要（不含 messages 大字段），按最后活动时间倒序。
func (s *PlaygroundConversationService) List(
	ctx context.Context,
	userID int64,
) ([]PlaygroundConversationSummary, error) {
	return s.repo.ListByUser(ctx, userID)
}

// Get 获取指定会话的完整数据（含 messages）。
// 若会话不存在或不属于当前用户，返回 ErrPlaygroundConversationNotFound。
func (s *PlaygroundConversationService) Get(
	ctx context.Context,
	id, userID int64,
) (*PlaygroundConversation, error) {
	return s.repo.GetByID(ctx, id, userID)
}

// Create 创建新会话。
//
// 业务规则：
//   - 单用户会话数不超过 PlaygroundConversationMaxPerUser（50），超限返回 ErrPlaygroundConversationLimitExceeded
//   - messages 字节数不超过 PlaygroundConversationMaxMessagesBytes（50MB），超限返回 ErrPlaygroundConversationTooLarge
//   - title 超过 255 字符时截断
//   - LastActivityAt 设置为当前时间
//
// 创建成功后，返回回填了 ID/CreatedAt/UpdatedAt 的 *PlaygroundConversation。
func (s *PlaygroundConversationService) Create(
	ctx context.Context,
	userID int64,
	title string,
	model *string,
	groupName *string,
	messages json.RawMessage,
) (*PlaygroundConversation, error) {
	// 检查会话数量上限
	count, err := s.repo.CountByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if count >= PlaygroundConversationMaxPerUser {
		return nil, ErrPlaygroundConversationLimitExceeded
	}

	// 检查 messages 体积上限
	if len(messages) > PlaygroundConversationMaxMessagesBytes {
		return nil, ErrPlaygroundConversationTooLarge
	}

	// title 截断至 255 字符（rune 安全截断）
	title = pcTruncateString(title, 255)

	now := time.Now()
	c := &PlaygroundConversation{
		UserID:         userID,
		Title:          title,
		Model:          model,
		GroupName:      groupName,
		Messages:       messages,
		LastActivityAt: now,
	}

	if err := s.repo.Create(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

// Update 部分更新会话字段。
//
// 业务规则：
//   - 传入 nil 的指针字段不修改（但 model/groupName 为 nil 时清空对应字段）
//   - messages 非 nil 时校验体积上限
//   - 任何有效更新都会刷新 LastActivityAt
//   - 若会话不存在或不属于当前用户，返回 ErrPlaygroundConversationNotFound
func (s *PlaygroundConversationService) Update(
	ctx context.Context,
	id, userID int64,
	title *string,
	model *string,
	groupName *string,
	messages json.RawMessage,
) error {
	// 先获取当前会话（校验归属 + 获取当前字段值）
	current, err := s.repo.GetByID(ctx, id, userID)
	if err != nil {
		return err
	}

	// 仅在有传入值时更新 title
	if title != nil {
		truncated := pcTruncateString(*title, 255)
		current.Title = truncated
	}

	// model/groupName：nil 表示清空（置为 nil 指针），非 nil 表示更新
	current.Model = model
	current.GroupName = groupName

	// messages 非 nil 时更新并校验体积
	if messages != nil {
		if len(messages) > PlaygroundConversationMaxMessagesBytes {
			return ErrPlaygroundConversationTooLarge
		}
		current.Messages = messages
	}

	// 刷新最后活动时间
	current.LastActivityAt = time.Now()

	return s.repo.Update(ctx, current)
}

// Delete 删除指定会话。
// 若会话不存在或不属于当前用户，返回 ErrPlaygroundConversationNotFound。
func (s *PlaygroundConversationService) Delete(
	ctx context.Context,
	id, userID int64,
) error {
	return s.repo.Delete(ctx, id, userID)
}

// pcTruncateString 按 rune（Unicode 字符）安全截断字符串至 maxRunes 个字符。
// 命名加 pc 前缀以区别于 ops_metrics_collector.go 中同名的工具函数。
func pcTruncateString(s string, maxRunes int) string {
	runes := []rune(s)
	if len(runes) <= maxRunes {
		return s
	}
	return strings.TrimSpace(string(runes[:maxRunes]))
}
