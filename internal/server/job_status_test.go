package server

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

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/siemens/wfx/generated/api"
	"github.com/siemens/wfx/internal/handler/job"
	"github.com/siemens/wfx/internal/handler/workflow"
	"github.com/siemens/wfx/workflow/dau"
	"github.com/steinfletcher/apitest"
	jsonpath "github.com/steinfletcher/apitest-jsonpath"
	"github.com/stretchr/testify/assert"
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

func TestGetJobsIDStatusClientIDHeader(t *testing.T) {
	var logs syncBuffer
	originalLogger := log.Logger
	t.Cleanup(func() { log.Logger = originalLogger })
	log.Logger = zerolog.New(&logs)

	db := newInMemoryDB(t)
	north, south := createNorthAndSouth(t, db)
	job := persistJob(t, db)
	path := fmt.Sprintf("/api/wfx/v1/jobs/%s/status", job.ID)

	apitest.New().
		Handler(north).
		Get(path).
		Header("X-Client-Id", "other").
		Expect(t).
		Status(http.StatusOK).
		End()

	apitest.New().
		Handler(south).
		Get(path).
		Header("X-Client-Id", job.ClientID).
		Expect(t).
		Status(http.StatusOK).
		End()

	apitest.New().
		Handler(south).
		Get(path).
		Header("X-Client-Id", "other").
		Expect(t).
		Status(http.StatusNotFound).
		Assert(jsonpath.Equal(`$.errors[0].code`, "wfx.jobNotFound")).
		End()
	assert.Contains(t, logs.String(), "Client ID mismatch for job status access")
}

func TestPutJobsIDStatusClientIDHeader(t *testing.T) {
	var logs syncBuffer
	originalLogger := log.Logger
	t.Cleanup(func() { log.Logger = originalLogger })
	log.Logger = zerolog.New(&logs)

	db := newInMemoryDB(t)
	north, south := createNorthAndSouth(t, db)
	job := persistJob(t, db)
	path := fmt.Sprintf("/api/wfx/v1/jobs/%s/status", job.ID)

	apitest.New().
		Handler(north).
		Put(path).
		Header("X-Client-Id", "other").
		Body(fmt.Sprintf(`{"clientId":%q,"state":"INSTALL"}`, job.ClientID)).
		ContentType("application/json").
		Expect(t).
		Status(http.StatusOK).
		End()

	tests := []struct {
		name     string
		headerID string
		bodyID   string
		status   int
	}{
		{name: "match", headerID: job.ClientID, bodyID: job.ClientID, status: http.StatusOK},
		{name: "body mismatch", headerID: "other", bodyID: job.ClientID, status: http.StatusNotFound},
		{name: "job mismatch", headerID: "other", bodyID: "other", status: http.StatusNotFound},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := apitest.New().
				Handler(south).
				Put(path).
				Header("X-Client-Id", tc.headerID).
				Body(fmt.Sprintf(`{"clientId":%q,"state":"INSTALL"}`, tc.bodyID)).
				ContentType("application/json").
				Expect(t).
				Status(tc.status)
			if tc.status == http.StatusNotFound {
				result.Assert(jsonpath.Equal(`$.errors[0].code`, "wfx.jobNotFound"))
			}
			result.End()
		})
	}
	assert.Contains(t, logs.String(), "Client ID mismatch for job status update")
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
