package events

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"context"

	"github.com/go-openapi/strfmt"
	"github.com/google/uuid"
	"github.com/olebedev/emitter"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/siemens/wfx/generated/api"
	"github.com/siemens/wfx/middleware/logging"
)

type JobEvent struct {
	// Ctime is the time when the event was created
	Ctime  strfmt.DateTime `json:"ctime"`
	Action Action          `json:"action"`
	Job    *api.Job        `json:"job"`
	Tags   []string        `json:"tags"`
}

type FilterParams struct {
	JobIDs    []string
	ClientIDs []string
	Workflows []string
}

type Action string

const (
	ActionCreate           Action = "CREATE"
	ActionDelete           Action = "DELETE"
	ActionAddTags          Action = "ADD_TAGS"
	ActionDeleteTags       Action = "DELETE_TAGS"
	ActionUpdateStatus     Action = "UPDATE_STATUS"
	ActionUpdateDefinition Action = "UPDATE_DEFINITION"
)

// NOTE: The number 32 for capacity is arbitrary; suggestions for a different value are welcome, particularly if supported by evidence.
var e *emitter.Emitter = emitter.New(32)

// AddSubscriber adds a new subscriber to receive job events filtered based on the provided filterParams.
func AddSubscriber(ctx context.Context, filter FilterParams, tags []string) (<-chan emitter.Event, error) {
	log := logging.LoggerFromCtx(ctx)

	// for logging purposes
	subscriberID := uuid.New().String()

	log.Info().
		Str("subscriberID", subscriberID).
		Dict("filterParams", zerolog.Dict().
			Strs("clientIDs", filter.ClientIDs).
			Strs("jobIDs", filter.JobIDs).
			Strs("workflows", filter.Workflows)).
		Strs("tags", tags).
		Msg("Adding new subscriber for job events")

	if filter := filter.createEventFilter(subscriberID, tags); filter != nil {
		return e.On("*", filter), nil
	}
	return e.On("*"), nil
}

// ShutdownSubscribers disconnects all subscribers.
func ShutdownSubscribers() {
	count := len(e.Topics())
	for _, topic := range e.Topics() {
		log.Debug().Str("topic", topic).Msg("Closing subscribers")
		e.Off(topic)
	}
	log.Info().Int("count", count).Msg("Subscriber shutdown complete")
}

// SubscriberCount counts the total number of subscribers across all topics.
func SubscriberCount() int {
	count := 0
	for _, topic := range e.Topics() {
		count += len(e.Listeners(topic))
	}
	return count
}

// PublishEvent publishes a new event. It returns a channel which can be used
// to wait for the delivery of the event to all listeners.
func PublishEvent(ctx context.Context, event *JobEvent) chan struct{} {
	log := logging.LoggerFromCtx(ctx)
	log.Debug().
		Str("jobID", event.Job.ID).
		Str("action", string(event.Action)).Msg("Publishing event")
	return e.Emit(event.Job.ID, event)
}

func (filter FilterParams) createEventFilter(subscriberID string, tags []string) func(*emitter.Event) {
	// special case: no filters means "catch-all"
	if len(filter.JobIDs) == 0 && len(filter.ClientIDs) == 0 && len(filter.Workflows) == 0 {
		return nil
	}

	// build hash sets for faster lookup
	jobIDSet := make(map[string]any, len(filter.JobIDs))
	for _, s := range filter.JobIDs {
		jobIDSet[s] = nil
	}
	clientIDSet := make(map[string]any, len(filter.ClientIDs))
	for _, s := range filter.ClientIDs {
		clientIDSet[s] = nil
	}
	workflowSet := make(map[string]any, len(filter.Workflows))
	for _, s := range filter.Workflows {
		workflowSet[s] = nil
	}
	return func(ev *emitter.Event) {
		event := ev.Args[0].(*JobEvent)
		job := event.Job
		// check if we shall notify the client about the event
		_, interested := jobIDSet[job.ID]
		if !interested {
			_, interested = clientIDSet[job.ClientID]
		}
		if !interested && job.Workflow != nil {
			_, interested = workflowSet[job.Workflow.Name]
		}

		if interested {
			// apply tags
			log.Debug().Strs("tags", tags).Msg("Applying tags to event notification")
			event.Tags = tags
		} else {
			log.Debug().
				Str("subscriberID", subscriberID).
				Str("jobID", job.ID).Msg("Skipping event notification")
			emitter.Void(ev)
		}
	}
}
