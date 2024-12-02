package api

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
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/siemens/wfx/generated/api"
	"github.com/siemens/wfx/internal/handler/job"
	"github.com/siemens/wfx/internal/handler/job/events"
	"github.com/siemens/wfx/internal/handler/job/status"
	"github.com/siemens/wfx/internal/handler/workflow"
	"github.com/siemens/wfx/workflow/dau"
	"github.com/steinfletcher/apitest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJobEventsSubscribe(t *testing.T) {
	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.Stamp})

	db := newInMemoryDB(t)
	wf := dau.DirectWorkflow()
	_, err := workflow.CreateWorkflow(context.Background(), db, wf)
	require.NoError(t, err)

	north, south := createNorthAndSouth(t, db)

	handlers := []http.Handler{north, south}
	for i, name := range allAPIs {
		handler := handlers[i]
		t.Run(name, func(t *testing.T) {
			clientID := "TestJobEventsSubscribe"

			var jobID atomic.Pointer[string]

			var wg sync.WaitGroup
			expectedTags := []string{"tag1", "tag2"}
			ch, _ := events.AddSubscriber(context.Background(), events.FilterParams{ClientIDs: []string{clientID}}, expectedTags)
			wg.Add(1)
			go func() {
				defer wg.Done()

				// wait for job created event
				ev := <-ch
				payload := ev.Args[0].(*events.JobEvent)
				assert.Equal(t, events.ActionCreate, payload.Action)
				assert.Equal(t, expectedTags, payload.Tags)
				jobID.Store(&payload.Job.ID)

				// wait for event created by our status.Update below
				<-ch
				// now our GET request should have received the response as well,
				// add some extra time to be safe
				time.Sleep(100 * time.Millisecond)
				events.ShutdownSubscribers()
			}()

			_, err := job.CreateJob(context.Background(), db, &api.JobRequest{ClientID: clientID, Workflow: wf.Name})
			require.NoError(t, err)

			wg.Add(1)
			go func() {
				defer wg.Done()
				// wait for subscriber which is created by our GET request below and our test goroutine above
				for events.SubscriberCount() != 2 {
					time.Sleep(20 * time.Millisecond)
				}
				// update job
				_, err = status.Update(context.Background(), db, *jobID.Load(), &api.JobStatus{State: "INSTALLING"}, api.CLIENT)
				require.NoError(t, err)
			}()

			// wait for job id
			for jobID.Load() == nil {
				time.Sleep(20 * time.Millisecond)
			}

			result := apitest.New().
				Handler(handler).
				Get("/api/wfx/v1/jobs/events").Query("ids", *jobID.Load()).
				Expect(t).
				Status(http.StatusOK).
				Header("Content-Type", "text/event-stream").
				End()

			data, _ := io.ReadAll(result.Response.Body)
			body := string(data)
			require.NotEmpty(t, body)

			lines := strings.Split(body, "\n")

			t.Log("HTTP resonse body:")
			for _, line := range lines {
				t.Logf(">> %s", line)
			}

			assert.Len(t, lines, 4)

			// check body starts with data:
			assert.True(t, strings.HasPrefix(lines[0], "data: "))

			// check content is a job and state is INSTALLING
			var ev events.JobEvent
			err = json.Unmarshal([]byte(strings.TrimPrefix(lines[0], "data: ")), &ev)
			require.NoError(t, err)
			assert.Equal(t, events.ActionUpdateStatus, ev.Action)
			assert.Equal(t, "INSTALLING", ev.Job.Status.State)
			assert.Equal(t, wf.Name, ev.Job.Workflow.Name)
			assert.Equal(t, clientID, ev.Job.ClientID)
			assert.Equal(t, "id: 1", lines[1])

			wg.Wait()
			events.ShutdownSubscribers()
		})
	}
}
