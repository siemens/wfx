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
	"fmt"
	"net/http"
	"testing"

	"github.com/Southclaws/fault/ftag"
	"github.com/siemens/wfx/generated/api"
	"github.com/siemens/wfx/internal/handler/job"
	"github.com/siemens/wfx/internal/handler/workflow"
	"github.com/siemens/wfx/internal/persistence/entgo"
	"github.com/siemens/wfx/persistence"
	"github.com/siemens/wfx/workflow/dau"
	"github.com/steinfletcher/apitest"
	jsonpath "github.com/steinfletcher/apitest-jsonpath"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var allAPIs = []string{"north", "south"}

func TestGetWorkflow(t *testing.T) {
	db := newInMemoryDB(t)
	north, south := createNorthAndSouth(t, db)

	tmp := dau.DirectWorkflow()
	tmp.Name = "45b68304-4a78-4f78-b4f5-776309c3616f"
	wf, err := workflow.CreateWorkflow(context.Background(), db, tmp)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = db.DeleteWorkflow(context.Background(), wf.Name)
	})

	handlers := []http.Handler{north, south}
	for i, name := range allAPIs {
		t.Run(name, func(t *testing.T) {
			// read one
			apitest.New().
				Handler(handlers[i]).
				Get(fmt.Sprintf("/api/wfx/v1/workflows/%s", wf.Name)).
				Expect(t).
				Status(http.StatusOK).
				Assert(jsonpath.Equal(`$.name`, wf.Name)).
				End()
		})
	}
}

func TestQueryWorkflows(t *testing.T) {
	db := newInMemoryDB(t)
	north, south := createNorthAndSouth(t, db)

	_, err := workflow.CreateWorkflow(context.Background(), db, dau.DirectWorkflow())
	require.NoError(t, err)
	handlers := []http.Handler{north, south}
	for i, name := range allAPIs {
		t.Run(name, func(t *testing.T) {
			// read all
			apitest.New().
				Handler(handlers[i]).
				Get("/api/wfx/v1/workflows").
				Expect(t).
				Status(http.StatusOK).
				Assert(jsonpath.GreaterThan(`$.content`, 0)).
				End()
		})
	}
}

func TestDeleteWorkflow(t *testing.T) {
	db := newInMemoryDB(t)
	north, south := createNorthAndSouth(t, db)

	tmp := dau.DirectWorkflow()
	name := "584802e1-3a90-483a-924f-a638e488c531"
	tmp.Name = name

	wf, err := workflow.CreateWorkflow(context.Background(), db, tmp)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = db.DeleteWorkflow(context.Background(), name)
	})

	// delete shall fail for south
	url := fmt.Sprintf("/api/wfx/v1/workflows/%s", name)
	apitest.New().
		Handler(south).
		Delete(url).
		Expect(t).
		Status(http.StatusForbidden).
		End()

	// delete shall succeed for north
	apitest.New().
		Handler(north).
		Delete(url).
		Expect(t).
		Status(http.StatusNoContent).
		End()

	actual, err := db.GetWorkflow(context.Background(), wf.Name)
	assert.Nil(t, actual)
	assert.Equal(t, ftag.NotFound, ftag.Get(err))
}

func TestCreateWorkflow(t *testing.T) {
	db := newInMemoryDB(t)
	north, south := createNorthAndSouth(t, db)

	wf := dau.DirectWorkflow()
	name := "d3dcf1b9-da32-431b-8efb-2e5e19dd503d"
	wf.Name = name
	wfJSON, _ := json.Marshal(wf)
	t.Cleanup(func() {
		_ = db.DeleteWorkflow(context.Background(), name)
	})

	// south is not allowed
	apitest.New().
		Handler(south).
		Post("/api/wfx/v1/workflows").
		Body(string(wfJSON)).
		Header("content-type", "application/json").
		Expect(t).
		Status(http.StatusForbidden).
		End()

	// north is allowed
	apitest.New().
		Handler(north).
		Post("/api/wfx/v1/workflows").
		Body(string(wfJSON)).
		Header("content-type", "application/json").
		Expect(t).
		Status(http.StatusCreated).
		Assert(jsonpath.Equal(`$.name`, wf.Name)).
		End()
}

func TestGetWorkflow_NotFound(t *testing.T) {
	north, south := createNorthAndSouth(t, newInMemoryDB(t))
	handlers := []http.Handler{north, south}
	for i, handler := range handlers {
		t.Run(allAPIs[i], func(t *testing.T) {
			apitest.New().
				Handler(handler).
				Get("/api/wfx/v1/workflows/foo").
				Expect(t).
				Status(http.StatusNotFound).
				End()
		})
	}
}

func TestDeleteWorkflow_NotFound(t *testing.T) {
	north, _ := createNorthAndSouth(t, newInMemoryDB(t))
	apitest.New().
		Handler(north).
		Delete("/api/wfx/v1/workflows/foo").
		Expect(t).
		Status(http.StatusNotFound).
		End()
}

func newInMemoryDB(t *testing.T) persistence.Storage {
	db := &entgo.SQLite{}
	err := db.Initialize("file:wfx?mode=memory&cache=shared&_fk=1")
	require.NoError(t, err)
	t.Cleanup(db.Shutdown)
	t.Cleanup(func() {
		{
			list, _ := db.QueryJobs(context.Background(), persistence.FilterParams{}, persistence.SortParams{}, persistence.PaginationParams{Limit: 100})
			for _, job := range list.Content {
				_ = db.DeleteJob(context.Background(), job.ID)
			}
		}
		{
			list, _ := db.QueryWorkflows(context.Background(), persistence.SortParams{Desc: false}, persistence.PaginationParams{Limit: 100})
			for _, wf := range list.Content {
				_ = db.DeleteWorkflow(context.Background(), wf.Name)
			}
		}
	})
	return db
}

func createNorthAndSouth(t *testing.T, db persistence.Storage) (http.Handler, http.Handler) {
	wfx := NewWfxServer(db)
	t.Cleanup(func() { wfx.Stop() })
	northAPI := NewNorthboundServer(wfx)
	north := api.HandlerFromMuxWithBaseURL(api.NewStrictHandler(northAPI, nil), http.NewServeMux(), "/api/wfx/v1")
	southAPI := NewSouthboundServer(wfx)
	south := api.HandlerFromMuxWithBaseURL(api.NewStrictHandler(southAPI, nil), http.NewServeMux(), "/api/wfx/v1")
	return north, south
}

func persistJob(t *testing.T, db persistence.Storage) *api.Job {
	wf := dau.DirectWorkflow()
	if found, _ := workflow.GetWorkflow(context.Background(), db, wf.Name); found == nil {
		_, err := workflow.CreateWorkflow(context.Background(), db, wf)
		require.NoError(t, err)
	}

	jobReq := api.JobRequest{
		ClientID: "foo",
		Workflow: wf.Name,
	}
	job, err := job.CreateJob(context.Background(), db, &jobReq)
	require.NoError(t, err)
	return job
}
