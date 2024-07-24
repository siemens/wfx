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
	"time"

	"github.com/siemens/wfx/generated/api"
	"github.com/siemens/wfx/persistence"
	"github.com/siemens/wfx/workflow/dau"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetJob(t *testing.T, db persistence.Storage) {
	tmpJob := newValidJob(defaultClientID)
	require.NotEmpty(t, tmpJob.Tags)
	_, err := db.CreateWorkflow(context.Background(), tmpJob.Workflow)
	require.NoError(t, err)

	require.NotEmpty(t, tmpJob.Tags)
	job, err := db.CreateJob(context.Background(), tmpJob)
	require.NoError(t, err)
	// ensure that an ID was generated
	require.NotEmpty(t, job.ID)
	require.Equal(t, tmpJob.Tags, job.Tags)

	actual, err := db.GetJob(context.Background(), job.ID, persistence.FetchParams{})
	require.NoError(t, err)
	assert.Equal(t, job.ID, actual.ID)

	// tags
	assert.Equal(t, job.Tags, actual.Tags)
}

func TestGetJobWithHistory(t *testing.T, db persistence.Storage) {
	clientID := "foo"

	tmpJob := newValidJob(clientID)
	_, err := db.CreateWorkflow(context.Background(), tmpJob.Workflow)
	require.NoError(t, err)
	job, err := db.CreateJob(context.Background(), tmpJob)
	require.NoError(t, err)

	var progress int32 = 42
	message := "First Update"
	_, err = db.UpdateJob(context.Background(), job,
		persistence.JobUpdate{Status: &api.JobStatus{Progress: &progress, Message: message, State: "DOWNLOADING"}})
	require.NoError(t, err)

	{
		job, err := db.GetJob(context.Background(), job.ID, persistence.FetchParams{History: true})
		require.NoError(t, err)
		assert.Len(t, *job.History, 1)

		job, err = db.GetJob(context.Background(), job.ID, persistence.FetchParams{History: false})
		require.NoError(t, err)
		assert.Nil(t, job.History)
	}

	result, err := db.QueryJobs(context.Background(), persistence.FilterParams{ClientID: &clientID}, sortAsc, defaultPaginationParams)
	require.NoError(t, err)
	actualJobs := result.Content
	assert.Len(t, actualJobs, 1)
	assert.Equal(t, job.ID, actualJobs[0].ID)
	// query jobs does not fetch history
	assert.Nil(t, actualJobs[0].History)
}

// Create a new, *unpersisted* job entity.
func newValidJob(clientID string) *api.Job {
	now := time.Now()
	return &api.Job{
		Mtime:    &now,
		Stime:    &now,
		ClientID: clientID,
		Status: &api.JobStatus{
			ClientID: clientID,
			State:    "CREATED",
		},
		Tags: []string{
			"tag1",
			"tag2",
		},
		Workflow: dau.DirectWorkflow(),
		History:  &[]api.History{},
	}
}
