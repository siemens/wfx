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
	"time"

	"github.com/siemens/wfx/generated/api"
	"github.com/siemens/wfx/internal/handler/job/events"
	"github.com/siemens/wfx/persistence"
	"github.com/siemens/wfx/workflow/dau"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDelete(t *testing.T) {
	db := newInMemoryDB(t)

	wf, err := db.CreateWorkflow(context.Background(), dau.DirectWorkflow())
	require.NoError(t, err)
	tags := []string{"foo", "bar"}
	job, err := db.CreateJob(context.Background(), &api.Job{
		ClientID: "foo",
		Workflow: wf,
		Status:   &api.JobStatus{State: "INSTALL"},
		Tags:     &tags,
	})
	require.NoError(t, err)

	sub := events.AddSubscriber(t.Context(), time.Minute, events.FilterParams{}, nil)

	tagList, err := Delete(t.Context(), db, job.ID, []string{"foo"})
	require.NoError(t, err)
	require.NotNil(t, tagList)
	expectedTags := []string{"bar"}
	assert.Equal(t, expectedTags, *tagList)

	jobEvent := <-sub.Events
	assert.Equal(t, events.ActionDeleteTags, jobEvent.Action)
	assert.Equal(t, job.ID, jobEvent.Job.ID)
	assert.Equal(t, expectedTags, *jobEvent.Job.Tags)
}

func TestDelete_FaultyStorageGet(t *testing.T) {
	db := persistence.NewHealthyMockStorage(t)
	ctx := context.Background()
	expectedErr := errors.New("mock error")
	db.On("GetJob", ctx, "1", persistence.FetchParams{History: false}).Return(nil, expectedErr)

	tags, err := Delete(ctx, db, "1", []string{"foo", "bar"})
	assert.Nil(t, tags)
	assert.NotNil(t, err)
}

func TestDelete_FaultyStorageUpdate(t *testing.T) {
	db := persistence.NewHealthyMockStorage(t)
	ctx := context.Background()

	expectedErr := errors.New("mock error")
	dummyJob := api.Job{ID: "1"}
	tags := []string{"foo", "bar"}

	db.On("GetJob", ctx, "1", persistence.FetchParams{History: false}).Return(&dummyJob, nil)
	db.On("UpdateJob", ctx, &dummyJob, persistence.JobUpdate{DelTags: &tags}).Return(nil, expectedErr)

	tagList, err := Delete(ctx, db, "1", tags)
	assert.Nil(t, tagList)
	assert.NotNil(t, err)
}
