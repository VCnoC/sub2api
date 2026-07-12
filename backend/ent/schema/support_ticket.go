// Package schema 定义 Ent ORM 的数据库 schema。
package schema

import (
	"time"

	"github.com/Wei-Shaw/sub2api/internal/domain"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// SupportTicket 保存用户工单的当前状态和列表查询字段。
type SupportTicket struct{ ent.Schema }

func (SupportTicket) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "support_tickets"}}
}

func (SupportTicket) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("user_id"),
		field.String("subject").MaxLen(200).NotEmpty(),
		field.String("category").MaxLen(32).Default(domain.TicketCategoryOther),
		field.String("status").MaxLen(24).Default(domain.TicketStatusPendingAdmin),
		field.String("priority").MaxLen(16).Default(domain.TicketPriorityNormal),
		field.Int64("assignee_id").Optional().Nillable(),
		field.Int64("closed_by").Optional().Nillable(),
		field.Time("closed_at").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("last_message_at").Default(time.Now).SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("created_at").Default(time.Now).Immutable().SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now).SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (SupportTicket) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("support_tickets").Field("user_id").Unique().Required(),
		edge.From("assignee", User.Type).Ref("assigned_support_tickets").Field("assignee_id").Unique(),
		edge.From("closed_by_user", User.Type).Ref("closed_support_tickets").Field("closed_by").Unique(),
		edge.To("messages", SupportTicketMessage.Type).Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("reads", SupportTicketRead.Type).Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func (SupportTicket) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "created_at"),
		index.Fields("user_id", "status"),
		index.Fields("status", "priority", "last_message_at"),
		index.Fields("assignee_id", "status"),
		index.Fields("category", "status"),
	}
}
