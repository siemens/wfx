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
	log.Debug().Int("count", count).Str("name", name).Msg("Deleted rows")
	if err != nil {
		log.Error().Err(err).Msg("Failed to delete workflow")
		return fault.Wrap(err)
	}
	if count <= 0 {
		return fault.Wrap(fmt.Errorf("workflow with name %s not found", name), ftag.With(ftag.NotFound))
	}
	return nil
}
