package tags

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"context"
	"errors"
	"sort"
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

func TestAdd(t *testing.T) {
	db := newInMemoryDB(t)

	wf, err := db.CreateWorkflow(context.Background(), dau.PhasedWorkflow())
	require.NoError(t, err)
	job, err := db.CreateJob(context.Background(), &api.Job{
		ClientID: "foo",
		Workflow: wf,
		Status:   &api.JobStatus{State: "CREATED"},
	})
	require.NoError(t, err)

	sub := events.AddSubscriber(t.Context(), time.Minute, events.FilterParams{}, nil)

	tags := []string{"foo", "bar"}
	actual, err := Add(context.Background(), db, job.ID, tags)
	require.NoError(t, err)
	sort.Strings(tags)

	assert.Equal(t, tags, actual)

	jobEvent := <-sub.Events
	assert.Equal(t, events.ActionAddTags, jobEvent.Action)
	assert.Equal(t, job.ID, jobEvent.Job.ID)
	assert.Equal(t, tags, jobEvent.Job.Tags)
}

func TestAdd_FaultyStorageGet(t *testing.T) {
	dbMock := persistence.NewHealthyMockStorage(t)
	ctx := context.Background()
	expectedErr := errors.New("mock error")
	dbMock.On("GetJob", ctx, "1", persistence.FetchParams{History: false}).Return(nil, expectedErr)

	tags, err := Add(ctx, dbMock, "1", []string{"foo", "bar"})
	assert.Nil(t, tags)
	assert.NotNil(t, err)
}

func TestAdd_FaultyStorageUpdate(t *testing.T) {
	dbMock := persistence.NewHealthyMockStorage(t)
	ctx := context.Background()

	expectedErr := errors.New("mock error")
	dummyJob := api.Job{ID: "1"}
	tags := []string{"foo", "bar"}

	dbMock.On("GetJob", ctx, "1", persistence.FetchParams{History: false}).Return(&dummyJob, nil)
	dbMock.On("UpdateJob", ctx, &dummyJob, persistence.JobUpdate{AddTags: &tags}).Return(nil, expectedErr)

	tags, err := Add(ctx, dbMock, "1", tags)
	assert.Nil(t, tags)
	assert.NotNil(t, err)
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
			list, _ := db.QueryWorkflows(context.Background(), persistence.SortParams{Desc: false}, persistence.PaginationParams{Limit: 100})
			for _, wf := range list.Content {
				_ = db.DeleteWorkflow(context.Background(), wf.Name)
			}
		}
	})
	return db
}
