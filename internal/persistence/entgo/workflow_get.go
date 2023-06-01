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
	"fmt"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/ftag"
	"github.com/siemens/wfx/generated/ent"
	"github.com/siemens/wfx/generated/ent/workflow"
	"github.com/siemens/wfx/generated/model"
)

func (db Database) GetWorkflow(ctx context.Context, name string) (*model.Workflow, error) {
	wf, err := db.client.Workflow.
		Query().
		Where(workflow.Name(name)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fault.Wrap(fmt.Errorf("workflow %s does not exist", name), ftag.With(ftag.NotFound))
		}
		return nil, fault.Wrap(err)
	}
	return convertWorkflow(wf), nil
}

func convertWorkflow(wf *ent.Workflow) *model.Workflow {
	return &model.Workflow{
		Name:        wf.Name,
		States:      wf.States,
		Transitions: wf.Transitions,
		Groups:      wf.Groups,
	}
}
