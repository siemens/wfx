package subscriber

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
)

// Subscriber is a subscriber to our topic.
type Subscriber[V any] struct {
	// events is a channel which receives the published messages
	events chan V
	// done indicates whether this subscriber should be removed
	done atomic.Bool
	// pending keeps track of active producers
	pending sync.WaitGroup
	// count delivered messages
	counter atomic.Int64
	// whether the subscriber has been shutdown
	shutdownComplete atomic.Bool
}

// NewSubscriber creates a new Subscriber receiving messages of type V.
func NewSubscriber[V any]() *Subscriber[V] {
	subscriber := Subscriber[V]{events: make(chan V, 1)}
	return &subscriber
}

// Send delivers a message to the subscriber with the specified timeout. The
// timeout helps ensure timely processing and prevents indefinitely blocking
// for message delivery.
// Note that if the subscriber is considered to be done, the message will not be sent.
func (subscriber *Subscriber[V]) Send(message V, timeout time.Duration) {
	if subscriber.Done() {
		log.Warn().Msg("Subscriber is already done, dropping message")
		return
	}

	subscriber.pending.Add(1)
	defer subscriber.pending.Done()
	timer := time.After(timeout)
	select {
	case subscriber.events <- message:
		// message delivered
		subscriber.counter.Add(1)
	case <-timer:
		log.Warn().Dur("timeout", timeout).
			Msg("Timeout occurred while attempting to deliver the message. The message was dropped and the subscriber has been marked as done.")
		subscriber.done.Store(true)
	}
}

// Events returns a channel receiving events (of type V) for the subscriber.
func (subscriber *Subscriber[V]) Events() <-chan V {
	return subscriber.events
}

// Done returns a boolean value indicating whether the subscriber has completed its operation.
// Returns:
//   - true if the subscriber has completed its operation and is no longer accepting new messages.
//   - false if the subscriber is still active and can accept new messages.
func (subscriber *Subscriber[V]) Done() bool {
	return subscriber.done.Load()
}

// Shutdown the given subscriber, i.e. close the channel when it's safe to do so.
func (subscriber *Subscriber[V]) Shutdown() {
	if !subscriber.shutdownComplete.Load() {
		// marks as done to prevent writing new messages to it
		subscriber.done.Store(true)
		subscriber.pending.Wait()
		// no one attempts to send a message anymore, close the channel now
		close(subscriber.events)
		subscriber.shutdownComplete.Store(true)
	}
}
