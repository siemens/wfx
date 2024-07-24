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
	"github.com/siemens/wfx/internal/handler/workflow"
	"github.com/siemens/wfx/persistence"
	"github.com/siemens/wfx/workflow/dau"
	"github.com/steinfletcher/apitest"
	jsonpath "github.com/steinfletcher/apitest-jsonpath"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetJobsHandler_Group(t *testing.T) {
	db := newInMemoryDB(t)

	workflow, err := db.CreateWorkflow(context.Background(), dau.DirectWorkflow())
	require.NoError(t, err)

	_, err = db.CreateJob(context.Background(), &api.Job{
		ClientID: "foo",
		Status: &api.JobStatus{
			State: "INSTALL",
		},
		Workflow: &api.Workflow{Name: workflow.Name},
	})
	assert.NoError(t, err)

	north, south := createNorthAndSouth(t, db)
	handlers := []http.Handler{north, south}
	for i, handler := range handlers {
		t.Run(allAPIs[i], func(t *testing.T) {
			apitest.New().
				Handler(handler).
				Get("/api/wfx/v1/jobs").
				Query("group", "OPEN").
				Expect(t).
				Assert(jsonpath.Len(`$.content`, 1)).
				Status(http.StatusOK).
				End()

			apitest.New().
				Handler(handler).
				Get("/api/wfx/v1/jobs").
				Query("group", "CLOSED").
				Expect(t).
				Assert(jsonpath.Len(`$.content`, 0)).
				Status(http.StatusOK).
				End()
		})
	}
}

func TestCreateJob(t *testing.T) {
	db := newInMemoryDB(t)
	north, _ := createNorthAndSouth(t, db)

	wf, err := workflow.CreateWorkflow(context.Background(), db, dau.DirectWorkflow())
	require.NoError(t, err)

	// create job using that workflow
	apitest.New().
		Handler(north).
		Post("/api/wfx/v1/jobs").
		Body(fmt.Sprintf(`
{
  "clientId": "gotest",
  "workflow": "%s",
  "definition": {
    "url": "http://localhost/update.tgz",
    "sha256": "8fcde8ae3c3641078ed98d5d6f20e706fab34c9768ec6366e2a025fe89e23464"
  }
}
`, wf.Name)).
		ContentType("application/json").
		Expect(t).
		Status(http.StatusCreated).
		Assert(jsonpath.Contains(`$.clientId`, "gotest")).
		Assert(jsonpath.Contains(`$.status.state`, "INSTALL")).
		Assert(jsonpath.Equal(`$.status.definitionHash`, "d1343b40f640f6da42ecc976d31b9e38cd991720436912756a08aafa222d66dc")).
		Assert(jsonpath.Equal(`$.definition.url`, "http://localhost/update.tgz")).
		End()

	jobs, err := db.QueryJobs(context.Background(), persistence.FilterParams{}, persistence.SortParams{}, persistence.PaginationParams{Offset: 0, Limit: 1})
	assert.NoError(t, err)
	assert.Len(t, jobs.Content, 1)
}

func TestCreateJob_SouthNotAllowed(t *testing.T) {
	db := newInMemoryDB(t)
	_, south := createNorthAndSouth(t, db)
	// create job shall fail for south
	apitest.New().
		Handler(south).
		Post("/api/wfx/v1/jobs").
		Body(`{"clientId":"gotest","workflow":"wfx.workflow.kanban","tags":[]}`).
		ContentType("application/json").
		Expect(t).
		Status(http.StatusForbidden).
		End()
}
