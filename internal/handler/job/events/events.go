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
	"sync"

	"github.com/go-openapi/strfmt"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/siemens/wfx/generated/api"
	"github.com/siemens/wfx/middleware/logging"
)

type Subscriber struct {
	subscriberID string
	ch           chan JobEvent
	jobIDSet     map[string]any
	clientIDSet  map[string]any
	workflowSet  map[string]any
	tags         []string
}

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

var (
	subscribers = make([]Subscriber, 0)
	mu          sync.RWMutex
)

// how many messages are kept in the subscriber's backlog before the subscriber is removed due to being considered dead
const backlogCapacity int = 32

// AddSubscriber adds a new subscriber to receive job events filtered based on the provided filterParams.
func AddSubscriber(ctx context.Context, filter FilterParams, tags []string) <-chan JobEvent {
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

	ch := make(chan JobEvent, backlogCapacity)

	// build hash maps for faster lookup
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

	l := Subscriber{
		subscriberID: subscriberID,
		ch:           ch,
		jobIDSet:     jobIDSet,
		clientIDSet:  clientIDSet,
		workflowSet:  workflowSet,
		tags:         tags,
	}

	mu.Lock()
	subscribers = append(subscribers, l)
	mu.Unlock()

	return ch
}

// ShutdownSubscribers disconnects all subscribers.
func ShutdownSubscribers() {
	mu.Lock()
	newSubscribers := subscribers
	subscribers = make([]Subscriber, 0)
	mu.Unlock()

	for _, l := range newSubscribers {
		close(l.ch)
	}
	log.Info().Msg("Subscriber shutdown complete")
}

// SubscriberCount counts the total number of subscribers across all topics.
func SubscriberCount() int {
	mu.RLock()
	count := len(subscribers)
	mu.RUnlock()
	return count
}

// PublishEvent publishes a new event. This is a synchronous operation.
func PublishEvent(ctx context.Context, event JobEvent) {
	log := logging.LoggerFromCtx(ctx).With().Str("jobID", event.Job.ID).Str("action", string(event.Action)).Logger()
	log.Debug().Msg("Publishing event to subscribers")

	mu.Lock()
	defer mu.Unlock()

	// the subscribers that are still active and we'll keep
	newSubscribers := make([]Subscriber, 0, len(subscribers))

	for _, l := range subscribers {
		ctxLog := log.With().Str("subscriberID", l.subscriberID).Logger()

		// check if we shall notify the subscriber about the event
		// special case: no filters means "catch-all"
		isCatchAll := len(l.jobIDSet) == 0 && len(l.clientIDSet) == 0 && len(l.workflowSet) == 0
		interested := isCatchAll ||
			mapContains(l.jobIDSet, event.Job.ID) ||
			mapContains(l.clientIDSet, event.Job.ClientID) ||
			(event.Job.Workflow != nil && mapContains(l.workflowSet, event.Job.Workflow.Name))
		if !interested {
			// keep subscriber, potentially still alive
			newSubscribers = append(newSubscribers, l)
			ctxLog.Debug().Msg("Subscriber not interested, skipping event notification")
			continue
		}

		// apply tags
		event.Tags = l.tags

		// send event
		ctxLog.Debug().Msg("Sending event to subscriber")
		select {
		case l.ch <- event:
			ctxLog.Debug().Msg("Sent event to subscriber")
			// keep subscriber since it's still alive
			newSubscribers = append(newSubscribers, l)
		default: // No subscriber
			close(l.ch)
			ctxLog.Info().Msg("Subscriber no longer alive, skipping send and dropping subscriber")
		}
	}

	subscribers = newSubscribers
}

func mapContains(set map[string]any, key string) (found bool) {
	_, found = set[key]
	return
}
