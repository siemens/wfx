package sse

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
	"testing"
	"time"

	"github.com/siemens/wfx/generated/api"
	"github.com/stretchr/testify/assert"
)

func TestSSEResponder(t *testing.T) {
	chEvents := make(chan api.JobStatus)

	rw := NewMockResponseRecorder()
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		responder := NewResponder(context.Background(), time.Minute, chEvents)
		if err := responder.VisitGetJobsEventsResponse(rw); err != nil {
			close(chEvents)
			t.Log("Received error from VisitGetJobsEventsResponse:", err)
			t.Fail()
		}
	}()

	clientID := "foo"
	message := "hello world"
	jobStatus := api.JobStatus{
		ClientID: clientID,
		Message:  message,
		State:    "INSTALLING",
	}
	chEvents <- jobStatus
	close(chEvents)

	wg.Wait()

	response := <-rw.ChResponse

	assert.Contains(t, response, `data: {"clientId":"foo","message":"hello world","state":"INSTALLING"}
id: 1

`)
}

func TestSSEResponder_IdlePing(t *testing.T) {
	chEvents := make(chan api.JobStatus)

	rw := NewMockResponseRecorder()
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		responder := NewResponder(context.Background(), time.Microsecond, chEvents)
		if err := responder.VisitGetJobsEventsResponse(rw); err != nil {
			close(chEvents)
			t.Log("Received error from VisitGetJobsEventsResponse:", err)
			t.Fail()
		}
	}()

	time.Sleep(time.Millisecond)
	close(chEvents)
	wg.Wait()

	response := <-rw.ChResponse
	assert.Contains(t, response, ": keepalive\n\n")
}
