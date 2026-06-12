// Package repository 提供对话广场会话的持久化实现。
//
// 实现惯例与 announcement_repo.go 保持一致：
//   - clientFromContext：支持事务上下文
//   - translatePersistenceError：统一错误翻译
//   - entity↔service 转换函数保持双向独立
package repository

import (
	"context"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/playgroundconversation"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

// playgroundConversationRepository 是 PlaygroundConversationRepository 接口的 ent 实现。
type playgroundConversationRepository struct {
	client *dbent.Client
}

// NewPlaygroundConversationRepository 创建对话广场会话仓储实例。
// 返回 service.PlaygroundConversationRepository 接口，屏蔽实现细节。
func NewPlaygroundConversationRepository(client *dbent.Client) service.PlaygroundConversationRepository {
	return &playgroundConversationRepository{client: client}
}

// ListByUser 返回指定用户的所有会话摘要，按 last_activity_at 倒序。
// 使用 Select 只取摘要字段，排除 messages 大字段，减少网络与内存开销。
func (r *playgroundConversationRepository) ListByUser(
	ctx context.Context,
	userID int64,
) ([]service.PlaygroundConversationSummary, error) {
	items, err := r.client.PlaygroundConversation.Query().
		Where(playgroundconversation.UserIDEQ(userID)).
		// 只选摘要字段，明确排除 messages（JSONB 大字段）
		Select(
			playgroundconversation.FieldID,
			playgroundconversation.FieldUserID,
			playgroundconversation.FieldTitle,
			playgroundconversation.FieldModel,
			playgroundconversation.FieldGroupName,
			playgroundconversation.FieldLastActivityAt,
			playgroundconversation.FieldCreatedAt,
			playgroundconversation.FieldUpdatedAt,
		).
		Order(dbent.Desc(playgroundconversation.FieldLastActivityAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return pcEntitiesToSummaries(items), nil
}

// GetByID 按 ID + userID 获取会话完整数据（含 messages）。
// userID 条件在 SQL 层强制附加，作为越权防护双保险。
func (r *playgroundConversationRepository) GetByID(
	ctx context.Context,
	id, userID int64,
) (*service.PlaygroundConversation, error) {
	m, err := r.client.PlaygroundConversation.Query().
		Where(
			playgroundconversation.IDEQ(id),
			playgroundconversation.UserIDEQ(userID),
		).
		Only(ctx)
	if err != nil {
		return nil, translatePersistenceError(err, service.ErrPlaygroundConversationNotFound, nil)
	}
	return pcEntityToService(m), nil
}

// Create 创建新会话记录，创建成功后回填 c.ID / c.CreatedAt / c.UpdatedAt。
func (r *playgroundConversationRepository) Create(
	ctx context.Context,
	c *service.PlaygroundConversation,
) error {
	client := clientFromContext(ctx, r.client)
	builder := client.PlaygroundConversation.Create().
		SetUserID(c.UserID).
		SetTitle(c.Title).
		SetNillableModel(c.Model).
		SetNillableGroupName(c.GroupName).
		SetLastActivityAt(c.LastActivityAt)

	if len(c.Messages) > 0 {
		builder.SetMessages(c.Messages)
	}

	created, err := builder.Save(ctx)
	if err != nil {
		return err
	}

	// 回填数据库生成的字段
	applyPCEntityToService(c, created)
	return nil
}

// Update 更新会话的 title/model/group_name/messages/last_activity_at 字段。
// 必须同时匹配 userID，防止越权修改；不存在时返回 ErrPlaygroundConversationNotFound。
func (r *playgroundConversationRepository) Update(
	ctx context.Context,
	c *service.PlaygroundConversation,
) error {
	client := clientFromContext(ctx, r.client)

	// 构建带 userID 过滤的批量更新（而非 UpdateOneID），确保越权防护
	builder := client.PlaygroundConversation.Update().
		Where(
			playgroundconversation.IDEQ(c.ID),
			playgroundconversation.UserIDEQ(c.UserID),
		).
		SetTitle(c.Title).
		SetLastActivityAt(c.LastActivityAt)

	// model 可选字段：有值则设置，nil 则清空
	if c.Model != nil {
		builder.SetModel(*c.Model)
	} else {
		builder.ClearModel()
	}

	// group_name 可选字段：有值则设置，nil 则清空
	if c.GroupName != nil {
		builder.SetGroupName(*c.GroupName)
	} else {
		builder.ClearGroupName()
	}

	// messages：有内容则更新，nil/空则清空
	if len(c.Messages) > 0 {
		builder.SetMessages(c.Messages)
	} else {
		builder.ClearMessages()
	}

	affected, err := builder.Save(ctx)
	if err != nil {
		return err
	}
	// 受影响行数为 0 说明会话不存在或不属于当前用户
	if affected == 0 {
		return service.ErrPlaygroundConversationNotFound
	}
	return nil
}

// Delete 删除指定会话，必须同时匹配 userID。
// 不存在时返回 ErrPlaygroundConversationNotFound。
func (r *playgroundConversationRepository) Delete(
	ctx context.Context,
	id, userID int64,
) error {
	client := clientFromContext(ctx, r.client)
	affected, err := client.PlaygroundConversation.Delete().
		Where(
			playgroundconversation.IDEQ(id),
			playgroundconversation.UserIDEQ(userID),
		).
		Exec(ctx)
	if err != nil {
		return err
	}
	if affected == 0 {
		return service.ErrPlaygroundConversationNotFound
	}
	return nil
}

// CountByUser 返回指定用户当前的会话总数，用于创建前的上限校验。
func (r *playgroundConversationRepository) CountByUser(
	ctx context.Context,
	userID int64,
) (int, error) {
	return r.client.PlaygroundConversation.Query().
		Where(playgroundconversation.UserIDEQ(userID)).
		Count(ctx)
}

// DeleteExpired 物理删除 last_activity_at < before 的过期会话。
//
// 实现方案选择：ent 的 Delete builder 不直接支持 LIMIT 子句（PostgreSQL 的
// DELETE ... LIMIT 本身也是非标准 SQL），因此本实现先用 SELECT IDs 查询一批 ID，
// 再用 IDIn 执行删除，从而实现 batchSize 控制。
//
// 这种方式的优点：
//   - 精确控制每批删除数量，避免单次长事务锁表
//   - 兼容 PostgreSQL（不依赖非标准 DELETE LIMIT）
//   - 实现简单，无需原生 SQL
//
// 调用方可在定时任务中循环调用直至返回 0，以实现完整清理。
func (r *playgroundConversationRepository) DeleteExpired(
	ctx context.Context,
	before time.Time,
	batchSize int,
) (int, error) {
	if batchSize <= 0 {
		batchSize = 1000 // 默认批次大小兜底
	}

	// 第一步：查询一批待删除的 ID（仅取 ID 列，避免拉取 messages 大字段）
	ids, err := r.client.PlaygroundConversation.Query().
		Where(playgroundconversation.LastActivityAtLT(before)).
		Limit(batchSize).
		IDs(ctx)
	if err != nil {
		return 0, err
	}
	if len(ids) == 0 {
		return 0, nil
	}

	// 第二步：按 ID 批量删除
	affected, err := r.client.PlaygroundConversation.Delete().
		Where(playgroundconversation.IDIn(ids...)).
		Exec(ctx)
	if err != nil {
		return 0, err
	}
	return affected, nil
}

// ---- 实体转换辅助函数 ----

// applyPCEntityToService 将数据库生成字段（ID/CreatedAt/UpdatedAt）回填到 service 结构。
func applyPCEntityToService(dst *service.PlaygroundConversation, src *dbent.PlaygroundConversation) {
	if dst == nil || src == nil {
		return
	}
	dst.ID = src.ID
	dst.CreatedAt = src.CreatedAt
	dst.UpdatedAt = src.UpdatedAt
}

// pcEntityToService 将 ent 实体（含 messages）转换为 service 领域结构。
func pcEntityToService(m *dbent.PlaygroundConversation) *service.PlaygroundConversation {
	if m == nil {
		return nil
	}
	return &service.PlaygroundConversation{
		ID:             m.ID,
		UserID:         m.UserID,
		Title:          m.Title,
		Model:          m.Model,
		GroupName:      m.GroupName,
		Messages:       m.Messages,
		LastActivityAt: m.LastActivityAt,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
}

// pcEntityToSummary 将 ent 实体（不含 messages）转换为摘要结构。
// 注意：当通过 Select 查询时，Messages 字段为 nil，此处不做拷贝。
func pcEntityToSummary(m *dbent.PlaygroundConversation) service.PlaygroundConversationSummary {
	return service.PlaygroundConversationSummary{
		ID:             m.ID,
		Title:          m.Title,
		Model:          m.Model,
		GroupName:      m.GroupName,
		LastActivityAt: m.LastActivityAt,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
}

// pcEntitiesToSummaries 批量转换实体为摘要切片。
func pcEntitiesToSummaries(models []*dbent.PlaygroundConversation) []service.PlaygroundConversationSummary {
	out := make([]service.PlaygroundConversationSummary, 0, len(models))
	for i := range models {
		if models[i] != nil {
			out = append(out, pcEntityToSummary(models[i]))
		}
	}
	return out
}
