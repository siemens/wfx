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
	"github.com/stretchr/testify/require"
)

func TestJobDefinitionGet(t *testing.T) {
	db := newInMemoryDB(t)
	north, south := createNorthAndSouth(t, db)
	job := persistJob(t, db)
	_, err := db.UpdateJob(context.Background(), job, persistence.JobUpdate{
		Definition: &map[string]any{
			"foo": "bar",
		},
	})
	require.NoError(t, err)

	path := fmt.Sprintf("/api/wfx/v1/jobs/%s/definition", job.ID)
	handlers := []http.Handler{north, south}
	for i, name := range allAPIs {
		t.Run(name, func(t *testing.T) {
			apitest.New().
				Handler(handlers[i]).
				Get(path).
				Expect(t).
				Status(http.StatusOK).
				Assert(jsonpath.Equal(`$.foo`, "bar")).
				End()
		})
	}
}

func TestJobDefinitionUpdate(t *testing.T) {
	db := newInMemoryDB(t)
	north, _ := createNorthAndSouth(t, db)
	job := persistJob(t, db)
	path := fmt.Sprintf("/api/wfx/v1/jobs/%s/definition", job.ID)

	oldDefinitionHash := job.Status.DefinitionHash

	// update definition
	apitest.New().
		Handler(north).
		Put(path).
		Body(`{ "url": "http://localhost/update2.tgz" }`).
		ContentType("application/json").
		Expect(t).
		Status(http.StatusOK).
		End()

	job, err := db.GetJob(context.Background(), job.ID, persistence.FetchParams{})
	assert.NoError(t, err)

	// check that Definition and DefinitionHash have been updated
	assert.Equal(t, "http://localhost/update2.tgz", job.Definition["url"])
	assert.Empty(t, job.Definition["sha256"]) // definition is replaced with new value
	assert.NotEqual(t, oldDefinitionHash, job.Status.DefinitionHash)
}
