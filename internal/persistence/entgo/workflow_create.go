package entgo

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"context"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fmsg"
	"github.com/Southclaws/fault/ftag"
	"github.com/siemens/wfx/generated/api"
	"github.com/siemens/wfx/generated/ent"
)

// CreateWorkflow creates a new workflow.
func (db Database) CreateWorkflow(ctx context.Context, workflow *api.Workflow) (*api.Workflow, error) {
	builder := db.client.Workflow.
		Create().
		SetName(workflow.Name).
		SetStates(workflow.States).
		SetTransitions(workflow.Transitions).
		SetGroups(workflow.Groups).
		SetDescription(workflow.Description)
	entity, err := builder.Save(ctx)
	if err != nil {
		if ent.IsConstraintError(err) {
			return nil, fault.Wrap(err, fmsg.With("failed to persist workflow due to constraints"), ftag.With(ftag.AlreadyExists))
		}
		return nil, fault.Wrap(err, fmsg.With("failed to persist workflow due to internal problem"), ftag.With(ftag.Internal))
	}
	wf := convertWorkflow(entity)
	return &wf, nil
}
