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
	"sort"
	"testing"

	"github.com/siemens/wfx/persistence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJobAddTags(t *testing.T, db persistence.Storage) {
	tmp := newValidJob(defaultClientID)
	_, err := db.CreateWorkflow(context.Background(), tmp.Workflow)
	require.NoError(t, err)

	assert.Empty(t, tmp.ID)
	job, err := db.CreateJob(context.Background(), tmp)
	require.NoError(t, err)

	newTags := []string{"foo", "bar"}
	require.NotContains(t, job.Tags, newTags)

	newTags = append(newTags, job.Tags...)
	sort.Strings(newTags)

	// add some tags
	updatedJob, err := db.UpdateJob(context.Background(), job, persistence.JobUpdate{
		AddTags: &newTags,
	})
	assert.NoError(t, err)
	assert.Equal(t, newTags, updatedJob.Tags)
}

func TestJobAddTagsOverlap(t *testing.T, db persistence.Storage) {
	tmp := newValidJob(defaultClientID)
	tmp.Tags = []string{"foo"}
	_, err := db.CreateWorkflow(context.Background(), tmp.Workflow)
	require.NoError(t, err)

	assert.Empty(t, tmp.ID)
	job, err := db.CreateJob(context.Background(), tmp)
	require.NoError(t, err)

	expectedTags := []string{"bar", "foo"}

	// add some tags which already exist
	updatedJob, err := db.UpdateJob(context.Background(), job, persistence.JobUpdate{
		AddTags: &expectedTags,
	})
	assert.NoError(t, err)
	assert.Equal(t, expectedTags, updatedJob.Tags)
}

func TestJobDeleteTags(t *testing.T, db persistence.Storage) {
	tmp := newValidJob(defaultClientID)
	_, err := db.CreateWorkflow(context.Background(), tmp.Workflow)
	require.NoError(t, err)
	require.NotEmpty(t, tmp.Tags)

	assert.Empty(t, tmp.ID)
	job, err := db.CreateJob(context.Background(), tmp)
	require.NoError(t, err)

	updatedJob, err := db.UpdateJob(context.Background(), job, persistence.JobUpdate{
		DelTags: &job.Tags,
	})
	assert.NoError(t, err)
	assert.Empty(t, updatedJob.Tags)
}

func TestJobDeleteTagsNonExisting(t *testing.T, db persistence.Storage) {
	tmp := newValidJob(defaultClientID)
	_, err := db.CreateWorkflow(context.Background(), tmp.Workflow)
	require.NoError(t, err)

	assert.Empty(t, tmp.ID)
	job, err := db.CreateJob(context.Background(), tmp)
	require.NoError(t, err)
	oldTags := job.Tags
	count := len(job.Tags)

	require.NotContains(t, job.Tags, "foo")
	updatedJob, err := db.UpdateJob(context.Background(), job, persistence.JobUpdate{
		DelTags: &[]string{"foo"},
	})
	assert.NoError(t, err)

	assert.Equal(t, oldTags, updatedJob.Tags)
	assert.Len(t, updatedJob.Tags, count)
}

func TestJobReuseExistingTags(t *testing.T, db persistence.Storage) {
	{
		tmp := newValidJob(defaultClientID)
		_, err := db.CreateWorkflow(context.Background(), tmp.Workflow)
		require.NoError(t, err)
		job, err := db.CreateJob(context.Background(), tmp)
		require.NoError(t, err)
		require.NotEmpty(t, job.Tags)
	}

	tmp := newValidJob(defaultClientID)
	require.NotEmpty(t, tmp.Tags)
	job, err := db.CreateJob(context.Background(), tmp)
	require.NoError(t, err)
	job2, err := db.GetJob(context.Background(), job.ID, persistence.FetchParams{})
	require.NoError(t, err)
	assert.Equal(t, job.Tags, job2.Tags)
}
