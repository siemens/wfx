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

func Delete(ctx context.Context, storage persistence.Storage, jobID string, tags []string) ([]string, error) {
	log := logging.LoggerFromCtx(ctx)
	contextLogger := log.With().Str("id", jobID).Strs("tags", tags).Logger()

	job, err := storage.GetJob(ctx, jobID, persistence.FetchParams{History: false})
	if err != nil {
		contextLogger.Err(err).Msg("Failed to get job from storage")
		return nil, fault.Wrap(err)
	}

	updatedJob, err := storage.UpdateJob(ctx, job, persistence.JobUpdate{DelTags: &tags})
	if err != nil {
		contextLogger.Err(err).Msg("Failed to delete tags to job")
		return nil, fault.Wrap(err)
	}
	contextLogger.Info().Msg("Deleted job tags")
	return updatedJob.Tags, nil
}
