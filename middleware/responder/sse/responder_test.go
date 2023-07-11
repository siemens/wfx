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
	"time"

	"github.com/siemens/wfx/generated/model"
	"github.com/siemens/wfx/internal/producer"
	"github.com/siemens/wfx/internal/pubsub/subscriber"
	"github.com/stretchr/testify/assert"
)

func TestSSEResponder(t *testing.T) {
	subscriber := subscriber.NewSubscriber[model.JobStatus]()

	rw := httptest.NewRecorder()
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		responder := Responder(context.Background(), subscriber)
		responder.WriteResponse(rw, producer.JSONProducer())
	}()

	event := model.JobStatus{
		ClientID: "klaus",
		Message:  "hello world",
		State:    "INSTALLING",
	}
	subscriber.Send(event, time.Second)
	subscriber.Shutdown()

	wg.Wait()

	assert.Equal(t, `data: {"clientId":"klaus","message":"hello world","state":"INSTALLING"}

`, rw.Body.String())
}
