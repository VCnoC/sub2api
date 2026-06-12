// Package schema 定义 Ent ORM 的数据库 schema。
// playground_conversation.go — 对话广场多会话持久化
//
// 用途：存储用户在对话广场创建的会话记录，采用会话级 JSONB 整存消息的方式，
// 后端透明存储、不解析消息结构。体验数据无审计价值，采用物理删除策略。
package schema

import (
	"encoding/json"
	"time"

	"github.com/Wei-Shaw/sub2api/ent/schema/mixins"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// PlaygroundConversation 定义对话广场会话的 schema。
//
// 删除策略：物理删除（体验数据无需软删除与审计追踪）。
// 清理策略：后台定时任务按 last_activity_at 批量物理删除超过保留期的会话。
type PlaygroundConversation struct {
	ent.Schema
}

// Annotations 指定底层表名。
func (PlaygroundConversation) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "playground_conversations"},
	}
}

// Mixin 仅引入时间戳混入，不引入软删除（物理删除策略）。
func (PlaygroundConversation) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.TimeMixin{},
	}
}

// Fields 定义会话实体的所有字段。
func (PlaygroundConversation) Fields() []ent.Field {
	return []ent.Field{
		// user_id: 会话所属用户 ID，用于鉴权隔离，所有查询必须携带此条件
		field.Int64("user_id").
			Comment("会话所属用户 ID"),

		// title: 会话标题，取首条用户消息前 20 字；支持手动改名
		field.String("title").
			MaxLen(255).
			Default("").
			Optional().
			Comment("会话标题（默认取首条消息前 20 字）"),

		// model: 会话最后使用的模型标识，用于下次打开时恢复模型选择
		field.String("model").
			Optional().
			Nillable().
			Comment("最后使用的模型标识（如 claude-3-5-sonnet）"),

		// group_name: 会话所属分组名称，用于前端分类展示（可选）
		field.String("group_name").
			Optional().
			Nillable().
			Comment("所属分组名称（可选，前端分类展示用）"),

		// messages: 会话完整消息列表，JSONB 整存，后端透明不解析结构
		// 前端负责序列化/反序列化，包含 versions/attachments 等扩展字段
		field.JSON("messages", json.RawMessage{}).
			Optional().
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}).
			Comment("消息列表（JSONB 整存，后端透明存储不解析）"),

		// last_activity_at: 最后活动时间，用于过期清理和列表排序
		// 每次保存消息时刷新；定时任务按此字段批量删除超期会话
		field.Time("last_activity_at").
			Default(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}).
			Comment("最后活动时间（保存消息时刷新，用于过期清理与列表排序）"),
	}
}

// Indexes 定义查询索引。
//
//   - (user_id, last_activity_at): 列表查询主索引，满足 NFR-003 P95 ≤ 200ms
//   - (last_activity_at): 清理任务全表扫描专用索引，避免 Seq Scan
func (PlaygroundConversation) Indexes() []ent.Index {
	return []ent.Index{
		// 用户会话列表查询：按活跃度倒序分页，复合索引覆盖过滤与排序
		index.Fields("user_id", "last_activity_at"),
		// 过期清理任务扫描：按 last_activity_at 范围删除，避免全表顺序扫描
		index.Fields("last_activity_at"),
	}
}
