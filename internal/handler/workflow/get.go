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
	"github.com/siemens/wfx/generated/model"
	"github.com/siemens/wfx/middleware/logging"
	"github.com/siemens/wfx/persistence"
)

func GetWorkflow(ctx context.Context, storage persistence.Storage, name string) (*model.Workflow, error) {
	log := logging.LoggerFromCtx(ctx)
	log.Debug().Str("name", name).Msg("Fetching workflow")
	workflow, err := storage.GetWorkflow(ctx, name)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get workflow from persistence storage")
		return nil, fault.Wrap(err)
	}
	return workflow, nil
}
