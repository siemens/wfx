package status

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"context"
	"fmt"
	"time"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/ftag"
	"github.com/go-openapi/strfmt"
	"github.com/siemens/wfx/generated/model"
	"github.com/siemens/wfx/internal/handler/job/events"
	"github.com/siemens/wfx/internal/workflow"
	"github.com/siemens/wfx/middleware/logging"
	"github.com/siemens/wfx/persistence"
)

func Update(ctx context.Context, storage persistence.Storage, jobID string, newStatus *model.JobStatus, actor model.EligibleEnum) (*model.JobStatus, error) {
	contextLogger := logging.LoggerFromCtx(ctx).With().Str("id", jobID).Str("actor", string(actor)).Logger()

	job, err := storage.GetJob(ctx, jobID, persistence.FetchParams{History: false})
	if err != nil {
		return nil, fault.Wrap(err)
	}

	from := job.Status.State

	// update status
	to := newStatus.State
	contextLogger = contextLogger.With().
		Str("name", job.Workflow.Name).
		Str("from", from).
		Str("to", to).
		Logger()
	contextLogger.Debug().Msg("Checking if transition is allowed")
	isAllowed := from == to // always allow trivial updates
	foundTransition := false
	for i := 0; !isAllowed && i < len(job.Workflow.Transitions); i++ {
		transition := job.Workflow.Transitions[i]
		if transition.From == from && transition.To == to {
			foundTransition = true
			if actor == transition.Eligible {
				isAllowed = true
				break
			}
		}
	}
	if !isAllowed {
		if !foundTransition {
			contextLogger.Warn().Msg("Transition does not exist")
			return nil, fault.Wrap(fmt.Errorf("transition from '%s' to '%s' does not exist", from, to), ftag.With(ftag.InvalidArgument))
		}
		contextLogger.Warn().Msg("Transition exists but actor is not allowed to trigger it")
		return nil, fault.Wrap(fmt.Errorf("transition from '%s' to '%s' is not allowed for actor '%s'", from, to, actor), ftag.With(ftag.InvalidArgument))
	}

	// transition is allowed, now apply wfx transitions
	newTo := workflow.FollowImmediateTransitions(job.Workflow, to)
	if newTo != to {
		contextLogger.Debug().Str("to", to).Str("newTo", newTo).Msg("Resetting state since we moved the transition forward")
		newStatus = &model.JobStatus{}
	}
	newStatus.State = newTo
	// override any definitionHash provided by client
	newStatus.DefinitionHash = job.Status.DefinitionHash

	result, err := storage.UpdateJob(ctx, job, persistence.JobUpdate{Status: newStatus})
	if err != nil {
		contextLogger.Err(err).Msg("Failed to persist job update")
		return nil, fault.Wrap(err)
	}

	_ = events.PublishEvent(ctx, &events.JobEvent{
		Ctime:  strfmt.DateTime(time.Now()),
		Action: events.ActionUpdateStatus,
		Job: &model.Job{
			ID:       result.ID,
			ClientID: result.ClientID,
			Workflow: &model.Workflow{Name: job.Workflow.Name},
			Status:   result.Status,
		},
	})

	contextLogger.Info().
		Str("from", from).
		Str("to", newStatus.State).
		Msg("Updated job status")
	return result.Status, nil
}
