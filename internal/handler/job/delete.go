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
	"time"

	"github.com/Southclaws/fault"
	"github.com/go-openapi/strfmt"
	"github.com/siemens/wfx/generated/api"
	"github.com/siemens/wfx/internal/handler/job/events"
	"github.com/siemens/wfx/middleware/logging"
	"github.com/siemens/wfx/persistence"
)

func DeleteJob(ctx context.Context, storage persistence.Storage, jobID string) error {
	log := logging.LoggerFromCtx(ctx)

	// we have to fetch the job because we need the `ClientID` and `Workflow` for
	// the job event notification
	job, err := storage.GetJob(ctx, jobID, persistence.FetchParams{History: false})
	if err != nil {
		return fault.Wrap(err)
	}

	if err := storage.DeleteJob(ctx, jobID); err != nil {
		return fault.Wrap(err)
	}

	go func() {
		events.PublishEvent(ctx, events.JobEvent{
			Ctime:  strfmt.DateTime(time.Now()),
			Action: events.ActionDelete,
			Job: &api.Job{
				ID:       jobID,
				ClientID: job.ClientID,
				Workflow: &api.Workflow{Name: job.Workflow.Name},
			},
		})
	}()

	log.Info().Str("id", jobID).Msg("Deleted job")
	return nil
}
