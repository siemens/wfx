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

	"github.com/olebedev/emitter"
	"github.com/siemens/wfx/generated/api"
	"github.com/stretchr/testify/assert"
)

func TestSSEResponder(t *testing.T) {
	events := make(chan emitter.Event)

	rw := NewMockResponseRecorder()
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		responder := NewResponder(context.Background(), events)
		if err := responder.VisitGetJobsEventsResponse(rw); err != nil {
			close(events)
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
	events <- emitter.Event{
		Topic:         "",
		OriginalTopic: "",
		Flags:         0,
		Args:          []any{jobStatus},
	}
	close(events)

	wg.Wait()

	response := <-rw.ChResponse

	assert.Contains(t, response, `data: {"clientId":"foo","message":"hello world","state":"INSTALLING"}
id: 1

`)
}
