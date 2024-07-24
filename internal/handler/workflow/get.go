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
	"github.com/siemens/wfx/generated/api"
	"github.com/siemens/wfx/middleware/logging"
	"github.com/siemens/wfx/persistence"
)

func GetWorkflow(ctx context.Context, storage persistence.Storage, name string) (*api.Workflow, error) {
	log := logging.LoggerFromCtx(ctx).With().Str("name", name).Logger()
	log.Debug().Str("name", name).Msg("Fetching workflow")
	workflow, err := storage.GetWorkflow(ctx, name)
	if err != nil {
		if ftag.Get(err) == ftag.NotFound {
			log.Debug().Msg("Workflow not found")
		} else {
			log.Error().Err(err).Msg("Failed to get workflow from storage")
		}
		return nil, fault.Wrap(err)
	}
	log.Debug().Msg("Found workflow")
	return workflow, nil
}
