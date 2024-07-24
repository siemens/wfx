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

	"github.com/siemens/wfx/persistence"
	"github.com/steinfletcher/apitest"
	jsonpath "github.com/steinfletcher/apitest-jsonpath"
	"github.com/stretchr/testify/assert"
)

func TestJobGetId(t *testing.T) {
	db := newInMemoryDB(t)
	north, south := createNorthAndSouth(t, db)
	job := persistJob(t, db)
	jobPath := fmt.Sprintf("/api/wfx/v1/jobs/%s", job.ID)

	// read job
	handlers := []http.Handler{north, south}
	for i, name := range allAPIs {
		t.Run(name, func(t *testing.T) {
			apitest.New().
				Handler(handlers[i]).
				Get(jobPath).
				Expect(t).
				Status(http.StatusOK).
				End()
		})
	}
}

func TestJobGetIdFilter(t *testing.T) {
	db := newInMemoryDB(t)
	north, south := createNorthAndSouth(t, db)
	job := persistJob(t, db)
	jobPath := fmt.Sprintf("/api/wfx/v1/jobs/%s", job.ID)

	// read job
	handlers := []http.Handler{north, south}
	for i, name := range allAPIs {
		t.Run(name, func(t *testing.T) {
			apitest.New().
				Handler(handlers[i]).
				Get(jobPath).
				Header("X-Response-Filter", ".status").
				Expect(t).
				Status(http.StatusOK).
				Assert(jsonpath.Equal(`$.state`, "INSTALL")).
				End()
		})
	}
}

func TestGetJobsIDHandlerNotFound(t *testing.T) {
	north, south := createNorthAndSouth(t, newInMemoryDB(t))
	handlers := []http.Handler{north, south}
	for i, handler := range handlers {
		t.Run(allAPIs[i], func(t *testing.T) {
			apitest.New().
				Handler(handler).
				Get("/api/wfx/v1/jobs/999999999").
				Expect(t).
				Status(http.StatusNotFound).
				End()
		})
	}
}

func TestDeleteJob(t *testing.T) {
	db := newInMemoryDB(t)
	north, south := createNorthAndSouth(t, db)

	job := persistJob(t, db)
	jobPath := fmt.Sprintf("/api/wfx/v1/jobs/%s", job.ID)

	// delete job shall fail for south
	apitest.New().
		Handler(south).
		Delete(jobPath).
		ContentType("application/json").
		Expect(t).
		Status(http.StatusForbidden).
		End()

	// delete job shall succeed for north
	apitest.New().
		Handler(north).
		Delete(jobPath).
		ContentType("application/json").
		Expect(t).
		Status(http.StatusNoContent).
		End()

	// ensure it's deleted
	list, err := db.QueryJobs(context.Background(), persistence.FilterParams{
		ClientID: &job.ClientID,
	}, persistence.SortParams{}, persistence.PaginationParams{Limit: 1})
	assert.NoError(t, err)
	assert.Empty(t, list.Content)
}
