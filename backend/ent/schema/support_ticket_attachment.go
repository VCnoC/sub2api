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

// SupportTicketAttachment 保存私有附件元数据；文件删除后元数据继续保留。
type SupportTicketAttachment struct{ ent.Schema }

func (SupportTicketAttachment) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "support_ticket_attachments"}}
}

func (SupportTicketAttachment) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("message_id"),
		field.Int64("uploader_id"),
		field.String("original_name").MaxLen(255).NotEmpty(),
		field.String("storage_key").MaxLen(255).NotEmpty().Unique(),
		field.String("media_type").MaxLen(100).NotEmpty(),
		field.Int64("size_bytes"),
		field.Time("delete_after").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("deleted_at").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Int64("deleted_by").Optional().Nillable(),
		field.String("delete_reason").MaxLen(255).Optional().Nillable(),
		field.Time("created_at").Default(time.Now).Immutable().SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (SupportTicketAttachment) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("message", SupportTicketMessage.Type).Ref("attachments").Field("message_id").Unique().Required(),
		edge.From("uploader", User.Type).Ref("support_ticket_attachments").Field("uploader_id").Unique().Required(),
		edge.From("deleted_by_user", User.Type).Ref("deleted_support_ticket_attachments").Field("deleted_by").Unique(),
	}
}

func (SupportTicketAttachment) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("message_id"),
		index.Fields("delete_after", "deleted_at"),
		index.Fields("uploader_id"),
	}
}
