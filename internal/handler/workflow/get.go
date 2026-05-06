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
	log.Debug().Msgf("Fetching workflow %q", name)
	workflow, err := storage.GetWorkflow(ctx, name)
	if err != nil {
		if ftag.Get(err) == ftag.NotFound {
			log.Debug().Msgf("Workflow %q not found", name)
		} else {
			log.Error().Err(err).Msgf("Failed to get workflow %q from storage", name)
		}
		return nil, fault.Wrap(err)
	}
	log.Debug().Msgf("Found workflow %q", name)
	return workflow, nil
}
