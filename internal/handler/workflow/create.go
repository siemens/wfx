package workflow

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
	"github.com/Southclaws/fault/ftag"
	"github.com/rs/zerolog/log"
	"github.com/siemens/wfx/generated/api"
	"github.com/siemens/wfx/persistence"
	"github.com/siemens/wfx/workflow"
)

func CreateWorkflow(ctx context.Context, storage persistence.Storage, wf *api.Workflow) (*api.Workflow, error) {
	if err := workflow.ValidateWorkflow(wf); err != nil {
		return nil, fault.Wrap(err, ftag.With(ftag.InvalidArgument))
	}
	wf, err := storage.CreateWorkflow(ctx, wf)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create workflow")
		return nil, fault.Wrap(err)
	}
	log.Info().Str("name", wf.Name).Msg("Created new workflow")
	return wf, nil
}
