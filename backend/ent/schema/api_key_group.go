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

// APIKeyGroup 保存 API Key 的有序候选分组。
type APIKeyGroup struct {
	ent.Schema
}

func (APIKeyGroup) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "api_key_groups"},
	}
}

func (APIKeyGroup) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("api_key_id"),
		field.Int64("group_id"),
		field.Int("priority").
			Min(0).
			Max(4),
		field.Time("created_at").
			Immutable().
			Default(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (APIKeyGroup) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("api_key", APIKey.Type).
			Unique().
			Required().
			Field("api_key_id"),
		edge.To("group", Group.Type).
			Unique().
			Required().
			Field("group_id"),
	}
}

func (APIKeyGroup) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("api_key_id", "group_id").Unique(),
		index.Fields("api_key_id", "priority").Unique(),
		index.Fields("group_id"),
	}
}
