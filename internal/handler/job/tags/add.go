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
	"time"

	"github.com/Southclaws/fault"
	"github.com/go-openapi/strfmt"
	"github.com/siemens/wfx/generated/api"
	"github.com/siemens/wfx/internal/handler/job/events"
	"github.com/siemens/wfx/middleware/logging"
	"github.com/siemens/wfx/persistence"
)

func Add(ctx context.Context, storage persistence.Storage, jobID string, tags []string) ([]string, error) {
	log := logging.LoggerFromCtx(ctx)
	contextLogger := log.With().Str("id", jobID).Strs("tags", tags).Logger()

	job, err := storage.GetJob(ctx, jobID, persistence.FetchParams{History: false})
	if err != nil {
		contextLogger.Err(err).Msg("Failed to get job from storage")
		return nil, fault.Wrap(err)
	}

	updatedJob, err := storage.UpdateJob(ctx, job, persistence.JobUpdate{AddTags: &tags})
	if err != nil {
		contextLogger.Err(err).Msg("Failed to add tags to job")
		return nil, fault.Wrap(err)
	}

	go func() {
		events.PublishEvent(ctx, events.JobEvent{
			Ctime:  strfmt.DateTime(time.Now()),
			Action: events.ActionAddTags,
			Job: &api.Job{
				ID:       updatedJob.ID,
				ClientID: updatedJob.ClientID,
				Workflow: updatedJob.Workflow,
				Tags:     updatedJob.Tags,
				Mtime:    updatedJob.Mtime,
			},
		})
	}()

	contextLogger.Info().Msg("Added job tags")
	return updatedJob.Tags, nil
}
