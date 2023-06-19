package topic

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"sync"
	"time"

	"github.com/siemens/wfx/internal/pubsub/subscriber"
)

// Topic is a poor man's topic implementation (thread-safe).
type Topic[V any] struct {
	subscribers []*subscriber.Subscriber[V]
	sync.RWMutex
	sendTimeout time.Duration
}

// NewTopic creates a new topic.
// The timeout parameter specifies the maximum amount of time allowed for delivering a message
// before the subscription is canceled.
func NewTopic[V any](timeout time.Duration) *Topic[V] {
	var result Topic[V]
	result.sendTimeout = timeout
	return &result
}

// Subscribe registers a client for receiving updates on the topic's value. It
// returns a channel that receives the updates for values of type V.
func (t *Topic[V]) Subscribe() *subscriber.Subscriber[V] {
	subscriber := subscriber.NewSubscriber[V]()
	t.Lock()
	t.subscribers = append(t.subscribers, subscriber)
	t.Unlock()
	return subscriber
}

// Unsubscribe removes the subscription from the list of subscribers and closes
// the subscription channel.
func (t *Topic[V]) Unsubscribe(subscriber *subscriber.Subscriber[V]) {
	subscriber.Shutdown()

	t.Lock()
	defer t.Unlock()
	for i, needle := range t.subscribers {
		if needle == subscriber {
			// remove the subscriber from the list of subscribers;
			// this is achieved bu moving the last element to the i-th position
			// and shrink the slice by one.
			n := len(t.subscribers)
			t.subscribers[i] = t.subscribers[n-1]
			t.subscribers = t.subscribers[:n-1]
			break
		}
	}
}

// Publish broadcasts a message to all subscribers.
// When this method returns, the message has been delivered (either successfully or not) to all subscribers.
func (t *Topic[V]) Publish(message V) {
	var wg sync.WaitGroup
	{
		t.Lock()
		defer t.Unlock()

		i := 0
		for {
			n := len(t.subscribers)
			if i == n || n == 0 {
				break
			}

			subscriber := t.subscribers[i]
			if subscriber.Done() {
				t.subscribers[i] = t.subscribers[n-1]
				t.subscribers = t.subscribers[:n-1]
				continue
			}
			i++

			wg.Add(1)
			go func() {
				defer wg.Done()
				subscriber.Send(message, t.sendTimeout)
			}()
		}
	}
	wg.Wait()
}

// Len returns the number of subscribers.
func (t *Topic[V]) Len() int {
	t.RLock()
	defer t.RUnlock()
	return len(t.subscribers)
}

// Shutdown closes all channels associated with the topic. This method should
// be called when there are no more messages to be sent and the topic is to be
// gracefully shutdown.
func (t *Topic[V]) Shutdown() {
	var subscribers []*subscriber.Subscriber[V]
	{
		t.Lock()
		defer t.Unlock()
		subscribers = t.subscribers
		t.subscribers = make([]*subscriber.Subscriber[V], 0)
	}
	for _, subscriber := range subscribers {
		subscriber.Shutdown()
	}
}
