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
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/olebedev/emitter"
	"github.com/siemens/wfx/generated/model"
	"github.com/siemens/wfx/internal/producer"
	"github.com/stretchr/testify/assert"
)

func TestSSEResponder(t *testing.T) {
	events := make(chan emitter.Event)

	rw := httptest.NewRecorder()
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		responder := Responder(context.Background(), events)
		responder.WriteResponse(rw, producer.JSONProducer())
	}()

	jobStatus := model.JobStatus{
		ClientID: "foo",
		Message:  "hello world",
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

	assert.Equal(t, `data: {"clientId":"foo","message":"hello world","state":"INSTALLING"}
id: 1

`, rw.Body.String())
}
