package definition

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"context"
	"testing"

	"github.com/Southclaws/fault/ftag"
	"github.com/siemens/wfx/generated/api"
	"github.com/siemens/wfx/internal/persistence/entgo"
	"github.com/siemens/wfx/persistence"
	"github.com/siemens/wfx/workflow/dau"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetJobDefinition(t *testing.T) {
	db := newInMemoryDB(t)
	wf, err := db.CreateWorkflow(context.Background(), dau.PhasedWorkflow())
	require.NoError(t, err)

	job, err := db.CreateJob(context.Background(), &api.Job{
		ClientID: "abc",
		Workflow: wf,
		Status:   &api.JobStatus{State: "CREATED"},
		Definition: map[string]any{
			"foo": "bar",
		},
	})
	require.NoError(t, err)

	definition, err := Get(context.Background(), db, job.ID)
	assert.NoError(t, err)
	assert.Equal(t, "bar", definition["foo"])
}

func TestGetJobDefinition_NotFound(t *testing.T) {
	db := newInMemoryDB(t)
	job, err := Get(context.Background(), db, "1")
	assert.Nil(t, job)
	assert.Equal(t, ftag.NotFound, ftag.Get(err))
}

func newInMemoryDB(t *testing.T) persistence.Storage {
	db := &entgo.SQLite{}
	err := db.Initialize("file:wfx?mode=memory&cache=shared&_fk=1")
	require.NoError(t, err)
	t.Cleanup(db.Shutdown)
	t.Cleanup(func() {
		{
			list, err := db.QueryJobs(context.Background(), persistence.FilterParams{}, persistence.SortParams{}, persistence.PaginationParams{Limit: 100})
			assert.NoError(t, err)
			for _, job := range list.Content {
				_ = db.DeleteJob(context.Background(), job.ID)
			}
		}
		{
			list, _ := db.QueryWorkflows(context.Background(), persistence.SortParams{}, persistence.PaginationParams{Limit: 100})
			for _, wf := range list.Content {
				_ = db.DeleteWorkflow(context.Background(), wf.Name)
			}
		}
	})
	return db
}
