//go:build testing

package tests

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

	"github.com/siemens/wfx/generated/model"
	"github.com/siemens/wfx/persistence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var defaultClientID = "foo"

func TestUpdateJobStatus(t *testing.T, db persistence.Storage) {
	tmp := newValidJob(defaultClientID)

	_, err := db.CreateWorkflow(context.Background(), tmp.Workflow)
	require.NoError(t, err)

	job, err := db.CreateJob(context.Background(), tmp)
	require.NoError(t, err)
	mtime := job.Mtime

	message := "Some arbitrary message"
	update := model.JobStatus{Message: message, State: "ACTIVATING"}
	updatedJob, err := db.UpdateJob(context.Background(), job, persistence.JobUpdate{Status: &update})
	assert.NoError(t, err)
	assert.Greater(t, *updatedJob.Mtime, *mtime)
	assert.Equal(t, "ACTIVATING", updatedJob.Status.State)
	assert.Equal(t, message, updatedJob.Status.Message)
	assert.Len(t, updatedJob.History, 0)

	{ // now fetch history and check our old state is there
		job, err := db.GetJob(context.Background(), job.ID, persistence.FetchParams{History: true})
		require.NoError(t, err)
		assert.Len(t, job.History, 1)
		assert.Equal(t, *job.Stime, job.History[0].Mtime)
	}
}

func TestUpdateJobStatusNonExisting(t *testing.T, db persistence.Storage) {
	job := newValidJob(defaultClientID)
	updatedJob, err := db.UpdateJob(context.Background(), job,
		persistence.JobUpdate{Status: &model.JobStatus{Message: "message", State: "ACTIVATING"}},
	)
	assert.Error(t, err)
	assert.Nil(t, updatedJob)
}
