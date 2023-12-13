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

	"github.com/siemens/wfx/persistence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateJobDefinition(t *testing.T, db persistence.Storage) {
	tmp := newValidJob(defaultClientID)
	tmp.Definition = make(map[string]any)
	tmp.Definition["url"] = "http://localhost/file.tgz"
	tmp.Definition["sha256"] = "92a8fa49a16ef4aa7e5dbdbaab6fe102fb79fb19373e3c7f43f0c5ac09eee66c"

	_, err := db.CreateWorkflow(context.Background(), tmp.Workflow)
	require.NoError(t, err)

	job, err := db.CreateJob(context.Background(), tmp)
	require.NoError(t, err)

	mtime := job.Mtime

	{
		job, err := db.GetJob(context.Background(), job.ID, persistence.FetchParams{})
		assert.NoError(t, err)
		assert.Equal(t, "http://localhost/file.tgz", job.Definition["url"])
		assert.Equal(t, "92a8fa49a16ef4aa7e5dbdbaab6fe102fb79fb19373e3c7f43f0c5ac09eee66c", job.Definition["sha256"])
	}

	newDefinition := map[string]any{"url": "http://localhost/new_file.tgz"}
	time.Sleep(10 * time.Millisecond)
	updatedJob, err := db.UpdateJob(context.Background(), job, persistence.JobUpdate{Definition: &newDefinition})
	assert.NoError(t, err)
	assert.Greater(t, *updatedJob.Mtime, *mtime)
	assert.Equal(t, "http://localhost/new_file.tgz", updatedJob.Definition["url"])
	assert.Empty(t, updatedJob.Definition["sha256"])
}
