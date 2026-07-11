// Package schema 定义视频异步任务的 Ent schema。
package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// VideoTask 保存视频上游终态轮询和失败退款所需的最小快照。
type VideoTask struct {
	ent.Schema
}

func (VideoTask) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "video_tasks"}}
}

func (VideoTask) Fields() []ent.Field {
	return []ent.Field{
		field.String("upstream_task_id").MaxLen(255).NotEmpty(),
		field.String("billing_request_id").MaxLen(128).NotEmpty(),
		field.Int64("user_id"),
		field.Int64("api_key_id"),
		field.Int64("account_id"),
		field.Int64("group_id"),
		field.Float("refund_amount").SchemaType(map[string]string{dialect.Postgres: "decimal(20,10)"}).Default(0),
		field.String("status").MaxLen(20).Default("pending"),
		field.Time("next_poll_at").Default(time.Now).SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("locked_until").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Int("poll_attempts").Default(0),
		field.Time("terminal_at").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("refunded_at").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.String("last_error").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "text"}),
		field.Time("created_at").Default(time.Now).Immutable().SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now).SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (VideoTask) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("upstream_task_id", "account_id").Unique(),
		index.Fields("billing_request_id", "api_key_id").Unique(),
		index.Fields("status", "next_poll_at"),
		index.Fields("locked_until"),
		index.Fields("user_id"),
	}
}
