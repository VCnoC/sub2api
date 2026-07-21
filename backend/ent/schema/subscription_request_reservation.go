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

// SubscriptionRequestReservation records one request-count entitlement hold.
type SubscriptionRequestReservation struct {
	ent.Schema
}

func (SubscriptionRequestReservation) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "subscription_request_reservations"},
	}
}

func (SubscriptionRequestReservation) Fields() []ent.Field {
	return []ent.Field{
		field.String("request_id").MaxLen(128),
		field.Int64("api_key_id"),
		field.Int64("user_id"),
		field.Int64("subscription_id"),
		field.String("status").MaxLen(20).Default("pending"),
		field.Time("window_5h_start").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("window_1d_start").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("expires_at").SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("created_at").Default(time.Now).Immutable().SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now).SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (SubscriptionRequestReservation) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("api_key", APIKey.Type).Ref("subscription_request_reservations").Field("api_key_id").Unique().Required(),
		edge.From("user", User.Type).Ref("subscription_request_reservations").Field("user_id").Unique().Required(),
		edge.From("subscription", UserSubscription.Type).Ref("request_reservations").Field("subscription_id").Unique().Required(),
	}
}

func (SubscriptionRequestReservation) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("request_id", "subscription_id").Unique(),
		index.Fields("subscription_id", "status", "expires_at"),
		index.Fields("api_key_id"),
		index.Fields("user_id"),
	}
}
