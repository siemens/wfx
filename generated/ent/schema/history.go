// SPDX-FileCopyrightText: 2023 Siemens AG
//
// SPDX-License-Identifier: Apache-2.0
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/siemens/wfx/generated/api"
)

// History holds the schema definition for the Job entity.
type History struct {
	ent.Schema
}

func (History) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "history"},
	}
}

// Fields of the History.
func (History) Fields() []ent.Field {
	return []ent.Field{
		field.Time("mtime").
			Comment("modification time").
			SchemaType(map[string]string{
				dialect.MySQL: "TIMESTAMP(6)", // microsecond precision
			}),
		field.JSON("status", api.JobStatus{}).Optional(),
		field.JSON("definition", map[string]any{}).Optional(),
	}
}

// Edges of the History.
func (History) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("job", Job.Type).
			Ref("history").
			Unique(),
	}
}
