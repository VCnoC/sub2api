// Package schema 定义 Ent ORM 的数据库 schema。
package schema

import (
	"encoding/json"
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

// SupportTicketMessage 保存不可变的公开回复、内部备注和系统事件。
type SupportTicketMessage struct{ ent.Schema }

func (SupportTicketMessage) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "support_ticket_messages"}}
}

func (SupportTicketMessage) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("ticket_id"),
		field.Int64("author_id").Optional().Nillable(),
		field.String("kind").MaxLen(16).Default(domain.TicketMessageKindPublic),
		field.String("visibility").MaxLen(16).Default(domain.TicketVisibilityUser),
		field.String("body").SchemaType(map[string]string{dialect.Postgres: "text"}).Default(""),
		field.JSON("metadata", json.RawMessage{}).Optional().SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.Time("created_at").Default(time.Now).Immutable().SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (SupportTicketMessage) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("ticket", SupportTicket.Type).Ref("messages").Field("ticket_id").Unique().Required(),
		edge.From("author", User.Type).Ref("support_ticket_messages").Field("author_id").Unique(),
		edge.To("attachments", SupportTicketAttachment.Type).Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func (SupportTicketMessage) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("ticket_id", "id"),
		index.Fields("ticket_id", "visibility", "id"),
		index.Fields("author_id"),
	}
}
