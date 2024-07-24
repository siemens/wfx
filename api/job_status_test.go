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
	"testing"

	"github.com/siemens/wfx/generated/api"
	"github.com/siemens/wfx/internal/handler/job"
	"github.com/siemens/wfx/internal/handler/workflow"
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

	jobReq := api.JobRequest{
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
		Body(`{"clientId": "foo", "state":"DOWNLOAD"}`).
		ContentType("application/json").
		Expect(t).
		Status(http.StatusBadRequest).
		End()

	// CREATED -> DOWNLOAD shall succeed for north
	apitest.New().
		Handler(north).
		Put(statusPath).
		Body(`{"clientId": "foo", "state":"DOWNLOAD"}`).
		ContentType("application/json").
		Expect(t).
		Status(http.StatusOK).
		Assert(jsonpath.Contains(`$.state`, "DOWNLOAD")).
		End()

	// DOWNLOAD -> DOWNLOADING shall fail for north
	apitest.New().
		Handler(north).
		Put(statusPath).
		Body(`{"clientId": "foo", "state":"DOWNLOADING"}`).
		ContentType("application/json").
		Expect(t).
		Status(http.StatusBadRequest).
		End()

	// DOWNLOAD -> DOWNLOADING shall succeed for south
	apitest.New().
		Handler(south).
		Put(statusPath).
		Body(`{"clientId":"foo","state":"DOWNLOADING"}`).
		ContentType("application/json").
		Expect(t).
		Status(http.StatusOK).
		Assert(jsonpath.Contains(`$.state`, "DOWNLOADING")).
		End()
}
