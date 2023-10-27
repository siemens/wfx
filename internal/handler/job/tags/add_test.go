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
	"testing"

	"github.com/siemens/wfx/generated/model"
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
	job, err := db.CreateJob(context.Background(), &model.Job{
		ClientID: "klaus",
		Workflow: wf,
		Status:   &model.JobStatus{State: "CREATED"},
	})
	require.NoError(t, err)

	actual, err := Add(context.Background(), db, job.ID, []string{"foo", "bar"})
	require.NoError(t, err)

	assert.Equal(t, []string{"bar", "foo"}, actual)
}

func TestAdd_FaultyStorageGet(t *testing.T) {
	db := persistence.NewMockStorage(t)
	ctx := context.Background()
	expectedErr := errors.New("mock error")
	db.On("GetJob", ctx, "1", persistence.FetchParams{History: false}).Return(nil, expectedErr)

	tags, err := Add(ctx, db, "1", []string{"foo", "bar"})
	assert.Nil(t, tags)
	assert.NotNil(t, err)
}

func TestAdd_FaultyStorageUpdate(t *testing.T) {
	db := persistence.NewMockStorage(t)
	ctx := context.Background()

	expectedErr := errors.New("mock error")
	dummyJob := model.Job{ID: "1"}
	tags := []string{"foo", "bar"}

	db.On("GetJob", ctx, "1", persistence.FetchParams{History: false}).Return(&dummyJob, nil)
	db.On("UpdateJob", ctx, &dummyJob, persistence.JobUpdate{AddTags: &tags}).Return(nil, expectedErr)

	tags, err := Add(ctx, db, "1", tags)
	assert.Nil(t, tags)
	assert.NotNil(t, err)
}

func newInMemoryDB(t *testing.T) persistence.Storage {
	db := &entgo.SQLite{}
	err := db.Initialize(context.Background(), "file:wfx?mode=memory&cache=shared&_fk=1")
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
			list, _ := db.QueryWorkflows(context.Background(), persistence.PaginationParams{Limit: 100})
			for _, wf := range list.Content {
				_ = db.DeleteWorkflow(context.Background(), wf.Name)
			}
		}
	})
	return db
}
