package job

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

func GetJob(ctx context.Context, storage persistence.Storage, id string, history bool) (*model.Job, error) {
	log := logging.LoggerFromCtx(ctx)
	contextLogger := log.With().
		Str("id", id).
		Bool("history", history).
		Logger()
	contextLogger.Debug().Msg("Fetching job")

	fetchParams := persistence.FetchParams{History: history}
	job, err := storage.GetJob(ctx, id, fetchParams)
	if err != nil {
		return nil, fault.Wrap(err)
	}
	return job, nil
}
