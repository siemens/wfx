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
	"testing"
	"time"

	"github.com/siemens/wfx/internal/pubsub/subscriber"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTopic(t *testing.T) {
	topic := NewTopic[string](time.Second)
	assert.NotNil(t, topic)
	assert.Equal(t, time.Second, topic.sendTimeout)
}

func TestSubscribe(t *testing.T) {
	topic := NewTopic[string](time.Second)
	ch := topic.Subscribe()
	assert.Equal(t, 1, topic.Len())
	topic.Unsubscribe(ch)
}

func TestUnsubscribe(t *testing.T) {
	topic := NewTopic[string](time.Second)
	ch := topic.Subscribe()
	topic.Unsubscribe(ch)
	assert.NotContains(t, topic.subscribers, ch)
}

func TestUnsubscribe_NotFound(t *testing.T) {
	topic := NewTopic[string](time.Second)
	ch := subscriber.NewSubscriber[string]()
	topic.Unsubscribe(ch)
	assert.NotContains(t, topic.subscribers, ch)
}

func TestPublish_One(t *testing.T) {
	topic := NewTopic[string](time.Second)
	sub := topic.Subscribe()
	topic.Publish("hello world")
	msg := <-sub.Events()
	assert.Equal(t, "hello world", msg)
	sub.Shutdown()
}

func TestPublish_Many(t *testing.T) {
	topic := NewTopic[string](time.Second)
	allSubs := make([]*subscriber.Subscriber[string], 0)
	for i := 0; i < 10; i++ {
		sub := topic.Subscribe()
		allSubs = append(allSubs, sub)
	}
	topic.Publish("hello world")
	for _, sub := range allSubs {
		msg := <-sub.Events()
		assert.Equal(t, "hello world", msg)
		sub.Shutdown()
	}
}

func TestLen(t *testing.T) {
	topic := NewTopic[string](time.Second)
	assert.Equal(t, 0, topic.Len())
	ch := topic.Subscribe()
	assert.Equal(t, 1, topic.Len())
	topic.Unsubscribe(ch)
	assert.Equal(t, 0, topic.Len())
}

func TestShutdown(t *testing.T) {
	topic := NewTopic[string](time.Second)
	_ = topic.Subscribe()
	topic.Shutdown()
	assert.Equal(t, 0, topic.Len())
}

func TestPublish_SequentialOrder(t *testing.T) {
	topic := NewTopic[int](time.Second)
	sub := topic.Subscribe()
	received := make([]int, 0, 100)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for msg := range sub.Events() {
			received = append(received, msg)
		}
	}()

	for i := 0; i < 100; i++ {
		topic.Publish(i)
	}
	topic.Shutdown()

	wg.Wait()
	assert.Len(t, received, 100)
	assert.IsIncreasing(t, received)
}

func TestPublish_Timeout(t *testing.T) {
	timeout := 5 * time.Millisecond
	topic := NewTopic[int](timeout)
	sub := topic.Subscribe()

	// this message is buffered
	topic.Publish(1)
	// the following message will be dropped since the buffer is full and we don't read from the channel
	topic.Publish(2)
	time.Sleep(2 * timeout)

	val := <-sub.Events()
	assert.Equal(t, 1, val)
	assert.True(t, sub.Done())
}

func TestPublish_GCRemoveAll(t *testing.T) {
	topic := NewTopic[int](time.Millisecond)
	sub1 := topic.Subscribe()
	sub2 := topic.Subscribe()

	// send first message, filling the buffer
	topic.Publish(42)
	require.False(t, sub1.Done())
	require.False(t, sub2.Done())

	// send second message, buffer is full, so it will fail
	topic.Publish(43)
	// subscribers have been marked as 'done'
	require.True(t, sub1.Done())
	require.True(t, sub2.Done())
	require.Equal(t, 2, topic.Len())

	// next Publish will trigger GC
	topic.Publish(44)
	assert.Equal(t, 0, topic.Len())
}

func TestPublish_GCRemoveMiddle(t *testing.T) {
	topic := NewTopic[int](time.Millisecond)
	sub1 := topic.Subscribe()
	sub2 := topic.Subscribe()
	sub3 := topic.Subscribe()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for ev := range sub1.Events() {
			t.Log("sub1 received event:", ev)
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		for ev := range sub3.Events() {
			t.Log("sub3 received event:", ev)
		}
	}()

	// send first message, filling the buffer
	topic.Publish(42)
	require.False(t, sub1.Done())
	require.False(t, sub2.Done())
	require.False(t, sub3.Done())

	// send second message, buffer of sub2 is full, so it will be marked as done
	topic.Publish(43)
	assert.False(t, sub1.Done())
	assert.True(t, sub2.Done())
	assert.False(t, sub3.Done())

	assert.Equal(t, 3, topic.Len())

	// next Publish will trigger GC
	topic.Publish(44)
	assert.Equal(t, 2, topic.Len())

	topic.Shutdown()
	wg.Wait()
}
