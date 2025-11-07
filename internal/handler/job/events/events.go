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
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/siemens/wfx/generated/api"
	"github.com/siemens/wfx/middleware/logging"
)

type Subscriber struct {
	Events  <-chan JobEvent // channel which receives the events for the subscriber
	Backlog *Backlog

	id            string        // unique identifier
	ch            chan JobEvent // internal RW of Events
	lastBlock     time.Time     // time of the last failed attempt to send a message
	graceInterval time.Duration // duration after which non-responsive subscribers are dropped

	jobIDSet    map[string]any // job filter
	clientIDSet map[string]any // clientID filter
	workflowSet map[string]any // workflow filter
	actionSet   map[Action]any // actions filter
	tags        []string       // tags to apply
}

func (s *Subscriber) ID() string {
	return s.id
}

type Backlog struct {
	data []JobEvent
	mu   sync.Mutex
}

// Enq adds a new JobEvent to the backlog.
func (b *Backlog) Enq(event JobEvent) {
	b.mu.Lock()
	b.data = append(b.data, event)
	b.mu.Unlock()
}

// Deq removes and returns the oldest (FIFO) JobEvent from the backlog.
func (b *Backlog) Deq() (*JobEvent, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if len(b.data) == 0 {
		return nil, false
	}
	var ev JobEvent
	ev, b.data = b.data[0], b.data[1:]
	return &ev, true
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
	Actions   []Action
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
	subscribers   = make([]*Subscriber, 0)
	muSubscribers sync.RWMutex // mutex used for locking subscribers
)

// AddSubscriber adds a new subscriber to receive job events filtered based on the provided filterParams.
func AddSubscriber(ctx context.Context, graceInterval time.Duration, filter FilterParams, tags []string) *Subscriber {
	log := logging.LoggerFromCtx(ctx)
	// for logging purposes
	subscriberID := uuid.New().String()
	log.Info().
		Str("subscriberID", subscriberID).
		Dict("filterParams", zerolog.Dict().
			Strs("clientIDs", filter.ClientIDs).
			Strs("jobIDs", filter.JobIDs).
			Strs("workflows", filter.Workflows).
			Interface("actions", filter.Actions)).
		Strs("tags", tags).
		Msg("Adding new subscriber for job events")

	ch := make(chan JobEvent, 1)

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

	subscriber := &Subscriber{
		id:            subscriberID,
		Events:        ch,
		ch:            ch,
		Backlog:       &Backlog{data: make([]JobEvent, 0)},
		lastBlock:     time.Time{},
		graceInterval: graceInterval,
		jobIDSet:      jobIDSet,
		clientIDSet:   clientIDSet,
		workflowSet:   workflowSet,
		tags:          tags,
	}

	muSubscribers.Lock()
	subscribers = append(subscribers, subscriber)
	muSubscribers.Unlock()

	return subscriber
}

// ShutdownSubscribers disconnects all subscribers.
func ShutdownSubscribers() {
	log.Info().Msg("Shutting down subscribers")
	muSubscribers.Lock()
	oldSubscribers := subscribers
	subscribers = make([]*Subscriber, 0)
	muSubscribers.Unlock()

	for _, subscriber := range oldSubscribers {
		close(subscriber.ch)
	}
	log.Info().Msg("Subscriber shutdown complete")
}

func RemoveSubscriber(subscriber *Subscriber) {
	log.Info().Str("id", subscriber.id).Msg("Removing subscriber")

	muSubscribers.Lock()
	defer muSubscribers.Unlock()

	oldSubscribers := subscribers
	subscribers = make([]*Subscriber, 0, len(subscribers))
	for _, sub := range oldSubscribers {
		if sub == subscriber {
			continue
		}
		subscribers = append(subscribers, sub)
	}
	close(subscriber.ch)
	log.Info().Str("id", subscriber.id).Msg("Removed subscriber")
}

// SubscriberCount counts the total number of subscribers across all topics.
func SubscriberCount() int {
	muSubscribers.RLock()
	defer muSubscribers.RUnlock()
	return len(subscribers)
}

// PublishEvent publishes a new event. This is a synchronous operation.
func PublishEvent(ctx context.Context, event JobEvent) {
	log := logging.LoggerFromCtx(ctx).With().Str("jobID", event.Job.ID).Str("action", string(event.Action)).Logger()

	muSubscribers.Lock()
	defer muSubscribers.Unlock()

	// the subscribers that are still active and we'll keep
	count := len(subscribers)
	newSubscribers := make([]*Subscriber, 0, count)

	log.Debug().Int("count", count).Msg("Publishing event to subscribers")
	for _, sub := range subscribers {
		ctxLog := log.With().Str("id", sub.id).Logger()

		// check if we shall notify the subscriber about the event
		// special case: no filters means "catch-all"
		isCatchAll := len(sub.jobIDSet) == 0 && len(sub.clientIDSet) == 0 && len(sub.workflowSet) == 0
		interested := isCatchAll ||
			mapContains(sub.jobIDSet, event.Job.ID) ||
			mapContains(sub.actionSet, event.Action) ||
			mapContains(sub.clientIDSet, event.Job.ClientID) ||
			(event.Job.Workflow != nil && mapContains(sub.workflowSet, event.Job.Workflow.Name))
		if !interested {
			// keep subscriber, potentially still alive
			newSubscribers = append(newSubscribers, sub)
			ctxLog.Debug().Msg("Subscriber not interested, skipping event notification")
			continue
		}

		// apply tags
		event.Tags = sub.tags

		// try to send event
		ctxLog.Debug().Msg("Sending event to subscriber")
		keepSubscriber := true
		select {
		case sub.ch <- event:
			ctxLog.Info().Msg("Sent event to subscriber")
			// reset lastBlock
			sub.lastBlock = time.Time{}
		default: // unable to send event (channel full)
			ctxLog.Info().Msg("Unable to send event, channel is full")
			// add to subscriber's backlog
			ctxLog.Debug().Msg("Adding event to subscriber backlog")
			sub.Backlog.Enq(event)
			// check if we should drop the subscriber
			keepSubscriber = sub.lastBlock.IsZero() || time.Since(sub.lastBlock) <= sub.graceInterval
			sub.lastBlock = time.Now()
		}
		if keepSubscriber {
			ctxLog.Debug().Msg("Keeping subscriber")
			newSubscribers = append(newSubscribers, sub)
		} else {
			ctxLog.Info().Msg("Dropping inactive subscriber")
			close(sub.ch)
		}
	}

	subscribers = newSubscribers
}

func mapContains[T comparable](set map[T]any, key T) bool {
	_, found := set[key]
	return found
}
