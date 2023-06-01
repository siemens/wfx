package status

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
	"github.com/siemens/wfx/generated/model"
	"github.com/siemens/wfx/internal/persistence/entgo"
	"github.com/siemens/wfx/persistence"
	"github.com/siemens/wfx/workflow/dau"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetJobStatus(t *testing.T) {
	db := newInMemoryDB(t)

	wf, err := db.CreateWorkflow(context.Background(), dau.PhasedWorkflow())
	require.NoError(t, err)

	tmpJob := model.Job{
		ClientID: "klaus",
		Workflow: wf,
		Status:   &model.JobStatus{State: "CREATED"},
	}

	job, err := db.CreateJob(context.Background(), &tmpJob)
	require.NoError(t, err)

	status, err := Get(context.Background(), db, job.ID)
	assert.NoError(t, err)
	assert.Equal(t, "CREATED", status.State)
}

func TestGetJobStatus_NotFound(t *testing.T) {
	db := newInMemoryDB(t)
	job, err := Get(context.Background(), db, "1")
	assert.Nil(t, job)
	ek := ftag.Get(err)
	assert.Equal(t, ftag.NotFound, ek)
}

func newInMemoryDB(t *testing.T) persistence.Storage {
	db := &entgo.SQLite{}
	err := db.Initialize(context.Background(), "file:wfx?mode=memory&cache=shared&_fk=1")
	require.NoError(t, err)

	require.NoError(t, err)
	t.Cleanup(func() {
		{
			list, err := db.QueryJobs(context.Background(), persistence.FilterParams{}, persistence.SortParams{}, persistence.PaginationParams{Limit: 100})
			assert.NoError(t, err)
			for _, job := range list.Content {
				_ = db.DeleteJob(context.Background(), job.ID)
			}
		}
		{
			list, _ := db.QueryWorkflows(context.Background(), persistence.PaginationParams{Limit: 100})
			for _, wf := range list.Content {
				_ = db.DeleteWorkflow(context.Background(), wf.Name)
			}
		}
	})
	return db
}
