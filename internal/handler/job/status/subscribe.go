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
	"errors"
	"sync"
	"time"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/ftag"
	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/siemens/wfx/generated/model"
	"github.com/siemens/wfx/internal/handler/job"
	"github.com/siemens/wfx/internal/pubsub/subscriber"
	"github.com/siemens/wfx/internal/pubsub/topic"
	"github.com/siemens/wfx/internal/workflow"
	"github.com/siemens/wfx/middleware/logging"
	"github.com/siemens/wfx/persistence"
)

// maps a job id to its topic
var topics = cmap.New[*topic.Topic[model.JobStatus]]()

// AddSubscriber sets up a new subscription for the specified jobID, allowing the retrieval of job status updates.
// It returns a channel that will receive events related to the specified jobID.
func AddSubscriber(ctx context.Context, storage persistence.Storage, jobID string) (*subscriber.Subscriber[model.JobStatus], error) {
	contextLogger := logging.LoggerFromCtx(ctx).With().Str("jobID", jobID).Logger()
	contextLogger.Debug().Msg("Adding subscriber for job updates")

	// job must exist and be in a non-terminal state
	j, err := job.GetJob(ctx, storage, jobID, false)
	if err != nil {
		return nil, fault.Wrap(err)
	}
	if workflow.IsTerminal(j.Workflow, j.Status.State) {
		contextLogger.Error().Str("jobID", jobID).Msg("Attempted to subscribe to a job which is in a terminal state")
		return nil, fault.Wrap(
			errors.New("attempted to subscribe to a job which is in a terminal state"),
			ftag.With(ftag.InvalidArgument))
	}
	// job is in correct state, subscription is allowed

	sendTimeout := 5 * time.Second
	t := topic.NewTopic[model.JobStatus](sendTimeout)
	if !topics.SetIfAbsent(jobID, t) {
		// topic was already present in the map, so we have to get it instead
		var found bool
		t, found = topics.Get(jobID)
		if !found {
			return nil, errors.New("internal error: failed to get topic")
		}
	}

	subscriber := t.Subscribe()
	go func() {
		// wait until client disconnects or server shuts down
		<-ctx.Done()

		// this closes the channel
		contextLogger.Info().Msg("Removing subscriber for job updates")
		t.Unsubscribe(subscriber)
	}()

	// send initial status
	subscriber.Send(*j.Status, sendTimeout)

	contextLogger.Info().Msg("Added subscriber for job updates")
	return subscriber, nil
}

// ShutdownSubscribers disconnects all subscribers.
func ShutdownSubscribers() {
	var wg sync.WaitGroup
	topics.IterCb(func(key string, topic *topic.Topic[model.JobStatus]) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			topic.Shutdown()
		}()
	})
	wg.Wait()
	topics.Clear()
}

// CountSubscribers counts the total number of subscribers across all topics.
func CountSubscribers() int {
	total := 0
	topics.IterCb(func(key string, v *topic.Topic[model.JobStatus]) {
		total += v.Len()
	})
	return total
}
