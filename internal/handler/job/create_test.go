package job

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
	"time"

	"github.com/siemens/wfx/generated/api"
	"github.com/siemens/wfx/internal/handler/job/events"
	"github.com/siemens/wfx/internal/persistence/entgo"
	"github.com/siemens/wfx/persistence"
	"github.com/siemens/wfx/workflow/dau"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateJob(t *testing.T) {
	db := newInMemoryDB(t)
	wf := createDirectWorkflow(t, db)

	job, err := CreateJob(context.Background(), db, &api.JobRequest{
		ClientID: "foo",
		Workflow: wf.Name,
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, job.Status.DefinitionHash)
	require.NotNil(t, job.Workflow)
	assert.Equal(t, wf.Name, job.Workflow.Name)
}

func TestCreateJob_Notification(t *testing.T) {
	db := newInMemoryDB(t)
	wf := createDirectWorkflow(t, db)

	subscriber := events.AddSubscriber(t.Context(), time.Minute, events.FilterParams{}, nil)

	job, err := CreateJob(context.Background(), db, &api.JobRequest{
		ClientID: "foo",
		Workflow: wf.Name,
	})
	require.NoError(t, err)

	jobEvent := <-subscriber.Events
	assert.Equal(t, events.ActionCreate, jobEvent.Action)
	assert.Equal(t, job.ID, jobEvent.Job.ID)
}

func newInMemoryDB(t *testing.T) persistence.Storage {
	db := &entgo.SQLite{}
	err := db.Initialize("file:wfx?mode=memory&cache=shared&_fk=1")
	require.NoError(t, err)
	t.Cleanup(db.Shutdown)

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
			list, _ := db.QueryWorkflows(context.Background(), persistence.SortParams{}, persistence.PaginationParams{Limit: 100})
			for _, wf := range list.Content {
				_ = db.DeleteWorkflow(context.Background(), wf.Name)
			}
		}
	})
	return db
}

func createDirectWorkflow(t *testing.T, db persistence.Storage) *api.Workflow {
	wf, err := db.CreateWorkflow(context.Background(), dau.DirectWorkflow())
	require.NoError(t, err)
	return wf
}
