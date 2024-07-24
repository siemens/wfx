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
	"github.com/siemens/wfx/generated/api"
	"github.com/siemens/wfx/middleware/logging"
	"github.com/siemens/wfx/persistence"
)

func GetJob(ctx context.Context, storage persistence.Storage, id string, history bool) (*api.Job, error) {
	fetchParams := persistence.FetchParams{History: history}
	job, err := storage.GetJob(ctx, id, fetchParams)
	if err != nil {
		log := logging.LoggerFromCtx(ctx)
		log.Error().Str("id", id).Bool("history", history).Err(err).Msg("Failed to get job from storage")
		return nil, fault.Wrap(err)
	}
	return job, nil
}
