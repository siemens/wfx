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

	"github.com/siemens/wfx/generated/model"
	"github.com/siemens/wfx/internal/handler/job"
	"github.com/siemens/wfx/internal/handler/workflow"
	"github.com/siemens/wfx/workflow/dau"
	"github.com/steinfletcher/apitest"
	"github.com/stretchr/testify/require"
)

func TestJobTagGet(t *testing.T) {
	db := newInMemoryDB(t)
	north, south := createNorthAndSouth(t, db)

	wf, err := workflow.CreateWorkflow(context.Background(), db, dau.DirectWorkflow())
	require.NoError(t, err)

	jobReq := model.JobRequest{
		ClientID: "foo",
		Workflow: wf.Name,
		Tags:     []string{"foo", "bar"},
	}
	job, err := job.CreateJob(context.Background(), db, &jobReq)
	require.NoError(t, err)

	jobPath := fmt.Sprintf("/api/wfx/v1/jobs/%s/tags", job.ID)
	handlers := []http.Handler{north, south}
	for i, name := range allAPIs {
		t.Run(name, func(t *testing.T) {
			apitest.New().
				Handler(handlers[i]).
				Get(jobPath).
				Expect(t).
				Status(http.StatusOK).
				Body(`["bar", "foo"]`).
				End()
		})
	}
}

func TestJobTagPost(t *testing.T) {
	db := newInMemoryDB(t)
	north, south := createNorthAndSouth(t, db)

	wf, err := workflow.CreateWorkflow(context.Background(), db, dau.DirectWorkflow())
	require.NoError(t, err)

	jobReq := model.JobRequest{
		ClientID: "foo",
		Workflow: wf.Name,
		Tags:     []string{"tag1"},
	}
	job, err := job.CreateJob(context.Background(), db, &jobReq)
	require.NoError(t, err)

	jobPath := fmt.Sprintf("/api/wfx/v1/jobs/%s/tags", job.ID)

	t.Run("north", func(t *testing.T) {
		apitest.New().
			Handler(north).
			Post(jobPath).
			ContentType("application/json").
			Body(`["bar", "foo"]`).
			Expect(t).
			Status(http.StatusOK).
			Body(`["bar", "foo", "tag1"]`).
			End()
	})

	t.Run("south", func(t *testing.T) {
		apitest.New().
			Handler(south).
			Post(jobPath).
			Expect(t).
			Status(http.StatusMethodNotAllowed).
			End()
	})
}

func TestJobTagDelete(t *testing.T) {
	db := newInMemoryDB(t)
	north, south := createNorthAndSouth(t, db)

	wf, err := workflow.CreateWorkflow(context.Background(), db, dau.DirectWorkflow())
	require.NoError(t, err)

	jobReq := model.JobRequest{
		ClientID: "foo",
		Workflow: wf.Name,
		Tags:     []string{"foo", "bar"},
	}
	job, err := job.CreateJob(context.Background(), db, &jobReq)
	require.NoError(t, err)

	jobPath := fmt.Sprintf("/api/wfx/v1/jobs/%s/tags", job.ID)

	t.Run("north", func(t *testing.T) {
		apitest.New().
			Handler(north).
			Delete(jobPath).
			ContentType("application/json").
			Body(`["foo"]`).
			Expect(t).
			Status(http.StatusOK).
			Body(`["bar"]`).
			End()
	})

	t.Run("south", func(t *testing.T) {
		apitest.New().
			Handler(south).
			Delete(jobPath).
			Expect(t).
			Status(http.StatusMethodNotAllowed).
			End()
	})
}
