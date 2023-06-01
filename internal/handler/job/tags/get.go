package tags

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
	"github.com/siemens/wfx/middleware/logging"
	"github.com/siemens/wfx/persistence"
)

func Get(ctx context.Context, storage persistence.Storage, jobID string) ([]string, error) {
	log := logging.LoggerFromCtx(ctx)
	contextLogger := log.With().Str("id", jobID).Logger()
	contextLogger.Debug().Msg("Fetching job tags")

	job, err := storage.GetJob(ctx, jobID, persistence.FetchParams{History: false})
	if err != nil {
		return nil, fault.Wrap(err)
	}
	contextLogger.Debug().Msg("Fetched job tags")
	return job.Tags, nil
}
