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
	"errors"
	"time"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/ftag"
	"github.com/go-openapi/strfmt"
	"github.com/siemens/wfx/generated/api"
	"github.com/siemens/wfx/internal/handler/job/definition"
	"github.com/siemens/wfx/internal/handler/job/events"
	"github.com/siemens/wfx/internal/workflow"
	"github.com/siemens/wfx/middleware/logging"
	"github.com/siemens/wfx/persistence"
)

func CreateJob(ctx context.Context, storage persistence.Storage, request *api.JobRequest) (*api.Job, error) {
	log := logging.LoggerFromCtx(ctx)
	contextLogger := log.With().Str("clientId", request.ClientID).Str("name", request.Workflow).Logger()

	wf, err := storage.GetWorkflow(ctx, request.Workflow)
	if err != nil {
		contextLogger.Error().Msg("Failed to get workflow from storage")
		return nil, fault.Wrap(err)
	}

	initial := workflow.FindInitialState(wf)
	if initial == nil {
		// should be caught by workflow validation
		return nil, errors.New("workflow has no initial state")
	}
	initialState := workflow.FollowImmediateTransitions(wf, *initial)

	now := time.Now()
	job := api.Job{
		ClientID: request.ClientID,
		Workflow: wf,
		Mtime:    &now,
		Stime:    &now,
		Status: &api.JobStatus{
			ClientID: request.ClientID,
			State:    initialState,
		},
		Definition: request.Definition,
		Tags:       request.Tags,
		History:    &[]api.History{},
	}
	job.Status.DefinitionHash = definition.Hash(&job)

	createdJob, err := storage.CreateJob(ctx, &job)
	if err != nil {
		contextLogger.Error().Err(err).Msg("Failed to persist job")
		return nil, fault.Wrap(err, ftag.With(ftag.Internal))
	}

	_ = events.PublishEvent(ctx, &events.JobEvent{
		Ctime:  strfmt.DateTime(time.Now()),
		Action: events.ActionCreate,
		Job:    createdJob,
	})

	contextLogger.Info().Str("id", createdJob.ID).Msg("Created new job")
	return createdJob, nil
}
