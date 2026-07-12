// Package schema 定义 Ent ORM 的数据库 schema。
package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// SupportTicketRead 保存每位查看者的消息已读游标。
type SupportTicketRead struct{ ent.Schema }

func (SupportTicketRead) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "support_ticket_reads"}}
}

func (SupportTicketRead) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("ticket_id"),
		field.Int64("user_id"),
		field.Int64("last_read_message_id").Default(0),
		field.Time("read_at").Default(time.Now).SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (SupportTicketRead) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("ticket", SupportTicket.Type).Ref("reads").Field("ticket_id").Unique().Required(),
		edge.From("user", User.Type).Ref("support_ticket_reads").Field("user_id").Unique().Required(),
	}
}

func (SupportTicketRead) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("ticket_id", "user_id").Unique(),
		index.Fields("user_id", "read_at"),
	}
}
