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
	"github.com/siemens/wfx/generated/ent/workflow"
	"github.com/siemens/wfx/middleware/logging"
)

// DeleteWorkflow deletes an existing workflow.
func (db Database) DeleteWorkflow(ctx context.Context, name string) error {
	log := logging.LoggerFromCtx(ctx)
	count, err := db.client.Workflow.
		Delete().
		Where(workflow.Name(name)).
		Exec(ctx)
	log.Debug().Int("count", count).Str("name", name).Msgf("Deleted %d row(s) for workflow %q", count, name)
	if err != nil {
		return fault.Wrap(err, fmsg.With("failed to delete workflow"))
	}
	if count <= 0 {
		return fault.Wrap(fault.Newf("workflow with name %s not found", name), ftag.With(ftag.NotFound))
	}
	return nil
}
