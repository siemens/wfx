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
	"testing"

	"github.com/Southclaws/fault/ftag"
	"github.com/siemens/wfx/generated/api"
	"github.com/siemens/wfx/internal/errkind"
	"github.com/siemens/wfx/persistence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var defaultClientID = "foo"

func TestUpdateJobStatus(t *testing.T, db persistence.Storage) {
	tmp := newValidJob(defaultClientID)

	_, err := db.CreateWorkflow(t.Context(), tmp.Workflow)
	require.NoError(t, err)

	job, err := db.CreateJob(t.Context(), tmp)
	require.NoError(t, err)
	mtime := job.Mtime

	message := "Some arbitrary message"
	update := api.JobStatus{Message: message, State: "ACTIVATING"}
	updatedJob, err := db.UpdateJob(t.Context(), job, persistence.JobUpdate{Status: &update})
	assert.NoError(t, err)
	assert.Greater(t, *updatedJob.Mtime, *mtime)
	assert.Equal(t, "ACTIVATING", updatedJob.Status.State)
	assert.Equal(t, message, updatedJob.Status.Message)
	assert.Nil(t, updatedJob.History)

	{ // now fetch history and check our old state is there
		job, err := db.GetJob(t.Context(), job.ID, persistence.FetchParams{History: true})
		history := *job.History
		require.NoError(t, err)
		assert.Len(t, history, 1)
		assert.Equal(t, *job.Stime, *history[0].Mtime)
	}
}

func TestUpdateJobStatusNonExisting(t *testing.T, db persistence.Storage) {
	job := newValidJob(defaultClientID)
	message := "message"
	updatedJob, err := db.UpdateJob(t.Context(), job,
		persistence.JobUpdate{Status: &api.JobStatus{Message: message, State: "ACTIVATING"}},
	)
	assert.Error(t, err)
	assert.Nil(t, updatedJob)
}

func TestUpdateJobStatusStaleView(t *testing.T, db persistence.Storage) {
	tmp := newValidJob(defaultClientID)
	_, err := db.CreateWorkflow(t.Context(), tmp.Workflow)
	require.NoError(t, err)

	job, err := db.CreateJob(t.Context(), tmp)
	require.NoError(t, err)

	// staleJob captures the state right after creation; we will use it for
	// the second update, which should fail.
	staleJob, err := db.GetJob(t.Context(), job.ID, persistence.FetchParams{})
	require.NoError(t, err)

	// First update succeeds and bumps mtime.
	freshJob, err := db.GetJob(t.Context(), job.ID, persistence.FetchParams{})
	require.NoError(t, err)
	winner, err := db.UpdateJob(t.Context(), freshJob,
		persistence.JobUpdate{Status: &api.JobStatus{State: "INSTALLING"}},
	)
	require.NoError(t, err)
	require.Equal(t, "INSTALLING", winner.Status.State)
	require.Greater(t, *winner.Mtime, *staleJob.Mtime)

	// Second update uses the stale view; it must be rejected.
	loser, err := db.UpdateJob(t.Context(), staleJob,
		persistence.JobUpdate{Status: &api.JobStatus{State: "ACTIVATING"}},
	)
	assert.Error(t, err, "stale update must not silently succeed")
	assert.Nil(t, loser)
	assert.Equal(t, errkind.TOCTOU, ftag.Get(err),
		"stale update should be flagged as a concurrent-modification conflict")

	// The job in the database must reflect only the winner's update.
	finalJob, err := db.GetJob(t.Context(), job.ID, persistence.FetchParams{})
	require.NoError(t, err)
	assert.Equal(t, "INSTALLING", finalJob.Status.State,
		"loser update must not have overwritten the winner")
}
