package subscriber

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewSubscriber(t *testing.T) {
	sub := NewSubscriber[int]()
	assert.NotNil(t, sub.events)
	assert.False(t, sub.Done())
	assert.Equal(t, int64(0), sub.counter.Load())
}

func TestSendAndReceive(t *testing.T) {
	sub := NewSubscriber[int]()
	sub.Send(42, time.Second)
	assert.Equal(t, int64(1), sub.counter.Load())
	val := <-sub.Events()
	assert.Equal(t, 42, val)
}

func TestShutdown(t *testing.T) {
	sub := NewSubscriber[int]()
	sub.Shutdown()
	assert.True(t, sub.done.Load())
	assert.True(t, sub.shutdownComplete.Load())
	_, ok := <-sub.Events()
	assert.False(t, ok, "channel should be closed")
}

func TestShutdown_Twice(t *testing.T) {
	sub := NewSubscriber[int]()
	sub.Shutdown()
	assert.True(t, sub.done.Load())
	sub.Shutdown()
	assert.True(t, sub.done.Load())
	_, ok := <-sub.Events()
	assert.False(t, ok, "channel should be closed")
}

func TestDone(t *testing.T) {
	sub := NewSubscriber[int]()
	sub.Shutdown()
	assert.True(t, sub.Done())
	sub.Send(42, 0)
}

func TestSend_Timeout(t *testing.T) {
	sub := NewSubscriber[int]()
	sub.Send(42, 0)
	sub.Send(42, 0)
	time.Sleep(time.Millisecond)
	assert.True(t, sub.Done())
}
