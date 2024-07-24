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
	"github.com/siemens/wfx/generated/api"
)

// Workflow holds the schema definition for the Workflow entity.
type Workflow struct {
	ent.Schema
}

func (Workflow) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "workflow"},
	}
}

// Fields of the Workflow.
func (Workflow) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Unique().
			MinLen(1).
			MaxLen(64),
		field.String("description").
			Optional().
			MaxLen(1024),
		field.JSON("states", []api.State{}),
		field.JSON("transitions", []api.Transition{}),
		field.JSON("groups", []api.Group{}),
	}
}

// Edges of the Workflow.
func (Workflow) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("jobs", Job.Type).
			Annotations(entsql.Annotation{
				OnDelete: entsql.NoAction,
			}),
	}
}
