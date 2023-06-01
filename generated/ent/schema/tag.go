// SPDX-FileCopyrightText: 2023 Siemens AG
//
// SPDX-License-Identifier: Apache-2.0
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Tag represents a job tag.
type Tag struct {
	ent.Schema
}

func (Tag) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "tag"},
	}
}

// Fields of the Job.
func (Tag) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").Comment("Name of the tag"),
	}
}

func (Tag) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name").Unique(),
	}
}

// Edges of the Group.
func (Tag) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("jobs", Job.Type),
	}
}
