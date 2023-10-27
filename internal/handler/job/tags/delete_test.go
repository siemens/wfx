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
	"github.com/siemens/wfx/persistence"
	"github.com/siemens/wfx/workflow/dau"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDelete(t *testing.T) {
	db := newInMemoryDB(t)

	wf, err := db.CreateWorkflow(context.Background(), dau.DirectWorkflow())
	require.NoError(t, err)
	job, err := db.CreateJob(context.Background(), &model.Job{
		ClientID: "klaus",
		Workflow: wf,
		Status:   &model.JobStatus{State: "INSTALL"},
		Tags:     []string{"foo", "bar"},
	})
	require.NoError(t, err)

	tags, err := Delete(context.Background(), db, job.ID, []string{"foo"})
	require.NoError(t, err)
	assert.Equal(t, []string{"bar"}, tags)
}

func TestDelete_FaultyStorageGet(t *testing.T) {
	db := persistence.NewMockStorage(t)
	ctx := context.Background()
	expectedErr := errors.New("mock error")
	db.On("GetJob", ctx, "1", persistence.FetchParams{History: false}).Return(nil, expectedErr)

	tags, err := Delete(ctx, db, "1", []string{"foo", "bar"})
	assert.Nil(t, tags)
	assert.NotNil(t, err)
}

func TestDelete_FaultyStorageUpdate(t *testing.T) {
	db := persistence.NewMockStorage(t)
	ctx := context.Background()

	expectedErr := errors.New("mock error")
	dummyJob := model.Job{ID: "1"}
	tags := []string{"foo", "bar"}

	db.On("GetJob", ctx, "1", persistence.FetchParams{History: false}).Return(&dummyJob, nil)
	db.On("UpdateJob", ctx, &dummyJob, persistence.JobUpdate{DelTags: &tags}).Return(nil, expectedErr)

	tags, err := Delete(ctx, db, "1", tags)
	assert.Nil(t, tags)
	assert.NotNil(t, err)
}
