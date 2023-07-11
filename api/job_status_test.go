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
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/siemens/wfx/generated/model"
	"github.com/siemens/wfx/internal/handler/job"
	"github.com/siemens/wfx/internal/handler/job/status"
	"github.com/siemens/wfx/internal/handler/workflow"
	"github.com/siemens/wfx/persistence"
	"github.com/siemens/wfx/workflow/dau"
	"github.com/steinfletcher/apitest"
	jsonpath "github.com/steinfletcher/apitest-jsonpath"
	"github.com/stretchr/testify/require"
)

func TestJobStatusGet(t *testing.T) {
	db := newInMemoryDB(t)
	north, south := createNorthAndSouth(t, db)
	job := persistJob(t, db)
	jobPath := fmt.Sprintf("/api/wfx/v1/jobs/%s/status", job.ID)
	handlers := []http.Handler{north, south}
	for i, name := range allAPIs {
		t.Run(name, func(t *testing.T) {
			apitest.New().
				Handler(handlers[i]).
				Get(jobPath).
				Expect(t).
				Status(http.StatusOK).
				Assert(jsonpath.Equal(`$.state`, "INSTALL")).
				End()
		})
	}
}

func TestPutJobsIDStatusHandlerNotFound(t *testing.T) {
	north, south := createNorthAndSouth(t, newInMemoryDB(t))
	handlers := []http.Handler{north, south}
	for i, handler := range handlers {
		t.Run(allAPIs[i], func(t *testing.T) {
			apitest.New().
				Handler(handler).
				Put("/api/wfx/v1/jobs/999999999/status").
				Body(`{"clientId": "foo", "state": "INSTALL", "message":"hello world"}`).
				ContentType("application/json").
				Expect(t).
				Status(http.StatusNotFound).
				End()
		})
	}
}

func TestJobStatusUpdate(t *testing.T) {
	db := newInMemoryDB(t)
	north, south := createNorthAndSouth(t, db)

	wf, err := workflow.CreateWorkflow(context.Background(), db, dau.PhasedWorkflow())
	require.NoError(t, err)

	jobReq := model.JobRequest{
		ClientID: "foo",
		Workflow: wf.Name,
	}
	job, err := job.CreateJob(context.Background(), db, &jobReq)
	require.NoError(t, err)

	jobID := job.ID
	statusPath := fmt.Sprintf("/api/wfx/v1/jobs/%s/status", jobID)

	// CREATED -> DOWNLOAD shall fail for south
	apitest.New().
		Handler(south).
		Put(statusPath).
		Body(`{"clientId": "klaus", "state":"DOWNLOAD"}`).
		ContentType("application/json").
		Expect(t).
		Status(http.StatusBadRequest).
		End()

	// CREATED -> DOWNLOAD shall succeed for north
	apitest.New().
		Handler(north).
		Put(statusPath).
		Body(`{"clientId": "klaus", "state":"DOWNLOAD"}`).
		ContentType("application/json").
		Expect(t).
		Status(http.StatusOK).
		Assert(jsonpath.Contains(`$.state`, "DOWNLOAD")).
		End()

	// DOWNLOAD -> DOWNLOADING shall fail for north
	apitest.New().
		Handler(north).
		Put(statusPath).
		Body(`{"clientId": "klaus", "state":"DOWNLOADING"}`).
		ContentType("application/json").
		Expect(t).
		Status(http.StatusBadRequest).
		End()

	// DOWNLOAD -> DOWNLOADING shall succeed for south
	apitest.New().
		Handler(south).
		Put(statusPath).
		Body(`{"clientId":"klaus","state":"DOWNLOADING"}`).
		ContentType("application/json").
		Expect(t).
		Status(http.StatusOK).
		Assert(jsonpath.Contains(`$.state`, "DOWNLOADING")).
		End()
}

func TestJobStatusSubscribe(t *testing.T) {
	db := newInMemoryDB(t)
	wf := dau.DirectWorkflow()
	_, err := workflow.CreateWorkflow(context.Background(), db, wf)
	require.NoError(t, err)

	north, south := createNorthAndSouth(t, db)

	handlers := []http.Handler{north, south}
	for i, name := range allAPIs {
		handler := handlers[i]
		t.Run(name, func(t *testing.T) {
			jobReq := model.JobRequest{
				ClientID: "foo",
				Workflow: wf.Name,
			}
			job, err := job.CreateJob(context.Background(), db, &jobReq)
			require.NoError(t, err)
			jobPath := fmt.Sprintf("/api/wfx/v1/jobs/%s/status/subscribe", job.ID)

			job, err = db.UpdateJob(context.Background(), job, persistence.JobUpdate{
				Status: &model.JobStatus{State: "ACTIVATING"},
			})
			require.NoError(t, err)

			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer wg.Done()
				for {
					if status.CountSubscribers() > 0 {
						break
					}
					time.Sleep(time.Millisecond)
				}
				// give it some extra time, just to be safe
				time.Sleep(500 * time.Millisecond)
				// update job to terminal state
				_, err = status.Update(context.Background(), db, job.ID, &model.JobStatus{State: "ACTIVATED"}, model.EligibleEnumCLIENT)
				require.NoError(t, err)
			}()

			apitest.New().
				Handler(handler).
				Get(jobPath).
				Expect(t).
				Status(http.StatusOK).
				Header("Content-Type", "text/event-stream").
				Body(`data: {"state":"ACTIVATING"}

data: {"state":"ACTIVATED"}

`).
				End()

			wg.Wait()
			status.ShutdownSubscribers()
		})
	}
}

func TestJobStatusSubscribe_NotFound(t *testing.T) {
	db := newInMemoryDB(t)
	north, south := createNorthAndSouth(t, db)

	handlers := []http.Handler{north, south}
	for i, name := range allAPIs {
		handler := handlers[i]
		t.Run(name, func(t *testing.T) {
			apitest.New().
				Handler(handler).
				Get("/api/wfx/v1/jobs/42/status/subscribe").
				Expect(t).
				Status(http.StatusNotFound).
				Header("Content-Type", "application/json").
				Assert(jsonpath.Equal(`$.errors[0].code`, "wfx.jobNotFound")).
				End()
		})
	}
}

func TestJobStatusSubscribe_TerminalState(t *testing.T) {
	db := newInMemoryDB(t)

	wf := dau.DirectWorkflow()
	_, _ = workflow.CreateWorkflow(context.Background(), db, wf)

	tmpJob := model.Job{
		ClientID: "foo",
		Workflow: wf,
		Status:   &model.JobStatus{State: "ACTIVATED"},
	}

	job, err := db.CreateJob(context.Background(), &tmpJob)
	require.NoError(t, err)

	jobPath := fmt.Sprintf("/api/wfx/v1/jobs/%s/status/subscribe", job.ID)

	north, south := createNorthAndSouth(t, db)
	handlers := []http.Handler{north, south}
	for i, name := range allAPIs {
		handler := handlers[i]
		t.Run(name, func(t *testing.T) {
			apitest.New().
				Handler(handler).
				Get(jobPath).
				Expect(t).
				Status(http.StatusBadRequest).
				Header("Content-Type", "application/json").
				Assert(jsonpath.Equal(`$.errors[0].code`, "wfx.jobTerminalState")).
				End()
		})
	}
}
