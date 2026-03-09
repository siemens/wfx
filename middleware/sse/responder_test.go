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
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Southclaws/fault"
	"github.com/siemens/wfx/generated/api"
	"github.com/siemens/wfx/internal/handler/job/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResponder(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)
	sub := events.AddSubscriber(ctx, time.Minute, events.FilterParams{}, []string{})

	events.PublishEvent(ctx, events.JobEvent{
		Action: events.ActionUpdateStatus,
		Job: &api.Job{
			ID: "1",
			Status: &api.JobStatus{
				ClientID: "foo",
				Message:  "hello world",
				State:    "INSTALLING",
			},
		},
		Tags: []string{},
	})

	rw := NewMockResponseRecorder(t)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		responder := NewResponder(ctx, time.Minute, sub)
		if err := responder.VisitGetJobsEventsResponse(rw); err != nil {
			t.Log("Received error from VisitGetJobsEventsResponse:", err)
			t.Fail()
		}
	}()

	expected := `{"ctime":"0001-01-01T00:00:00.000Z","action":"UPDATE_STATUS","job":{"id":"1","status":{"clientId":"foo","message":"hello world","state":"INSTALLING"}},"tags":[]}`

	var resp string
	for range 100 {
		resp = rw.Response()
		time.Sleep(10 * time.Millisecond)
		if strings.Contains(resp, "data:") {
			break
		}
	}

	obj, err := extractAndParseData(resp)
	require.NoError(t, err)
	objJson, _ := json.Marshal(obj)

	assert.JSONEq(t, expected, string(objJson))
	assert.Contains(t, resp, "id: 1")
	assert.Contains(t, resp, "\n\n")

	cancel()
	wg.Wait()
}

func extractAndParseData(response string) (map[string]any, error) {
	lines := strings.SplitSeq(response, "\n")
	for line := range lines {
		line = strings.TrimSpace(line)
		if dataStr, ok := strings.CutPrefix(line, "data: "); ok {
			var result map[string]any
			err := json.Unmarshal([]byte(dataStr), &result)
			return result, fault.Wrap(err)
		}
	}
	return nil, fmt.Errorf("no 'data:' line found")
}

func TestResponder_IdlePing(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)
	sub := events.AddSubscriber(ctx, time.Minute, events.FilterParams{}, []string{})

	rw := NewMockResponseRecorder(t)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		responder := NewResponder(ctx, time.Microsecond, sub)
		_ = responder.VisitGetJobsEventsResponse(rw)
	}()

	expected := ": keepalive"
	var resp string
	for range 100 {
		resp = rw.Response()
		time.Sleep(10 * time.Millisecond)
		if strings.Contains(resp, expected) {
			break
		}
	}
	assert.Contains(t, resp, expected)

	cancel()
	wg.Wait()
}
