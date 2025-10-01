package definition

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/Southclaws/fault"
	"github.com/cnf/structhash"
	"github.com/go-openapi/strfmt"
	"github.com/siemens/wfx/generated/api"
	"github.com/siemens/wfx/internal/handler/job/events"
	"github.com/siemens/wfx/middleware/logging"
	"github.com/siemens/wfx/persistence"
)

func Update(ctx context.Context, storage persistence.Storage, jobID string, definition map[string]any) (map[string]any, error) {
	log := logging.LoggerFromCtx(ctx)
	contextLogger := log.With().Str("id", jobID).Logger()

	job, err := storage.GetJob(ctx, jobID, persistence.FetchParams{History: false})
	if err != nil {
		contextLogger.Err(err).Msg("Failed to get job from storage")
		return nil, fault.Wrap(err)
	}

	job.Definition = definition
	job.Status.DefinitionHash = Hash(job)

	result, err := storage.UpdateJob(ctx, job, persistence.JobUpdate{Status: job.Status, Definition: &job.Definition})
	if err != nil {
		contextLogger.Err(err).Msg("Failed to update job")
		return nil, fault.Wrap(err)
	}

	go func() {
		events.PublishEvent(ctx, events.JobEvent{
			Ctime:  strfmt.DateTime(time.Now()),
			Action: events.ActionUpdateDefinition,
			Job: &api.Job{
				ID:         result.ID,
				ClientID:   result.ClientID,
				Workflow:   &api.Workflow{Name: job.Workflow.Name},
				Definition: result.Definition,
				Mtime:      result.Mtime,
				Status: &api.JobStatus{
					DefinitionHash: result.Status.DefinitionHash,
				},
			},
		})
	}()

	contextLogger.Info().Msg("Updated job definition")
	return result.Definition, nil
}

func Hash(job *api.Job) string {
	hasher := sha256.New()
	hasher.Write(structhash.Dump(job.Definition, 1))
	return fmt.Sprintf("%x", hasher.Sum(nil))
}
