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
	"fmt"
	"sort"
	"sync"
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

	newTags := make([]string, 0)
	newTags = append(newTags, "foo", "bar")
	require.NotContains(t, *job.Tags, newTags)

	if job.Tags != nil {
		newTags = append(newTags, *job.Tags...)
	}
	sort.Strings(newTags)

	// add some tags
	updatedJob, err := db.UpdateJob(context.Background(), job, persistence.JobUpdate{
		AddTags: &newTags,
	})
	assert.NoError(t, err)
	assert.Equal(t, newTags, *updatedJob.Tags)
}

func TestJobAddTagsOverlap(t *testing.T, db persistence.Storage) {
	tmp := newValidJob(defaultClientID)
	tags := []string{"foo"}
	tmp.Tags = &tags
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
	assert.Equal(t, expectedTags, *updatedJob.Tags)
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
		DelTags: job.Tags,
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
	count := len(*job.Tags)

	require.NotContains(t, *job.Tags, "foo")
	updatedJob, err := db.UpdateJob(context.Background(), job, persistence.JobUpdate{
		DelTags: &[]string{"foo"},
	})
	assert.NoError(t, err)

	assert.Equal(t, oldTags, updatedJob.Tags)
	assert.Len(t, *updatedJob.Tags, count)
}

func TestJobAddTagsConcurrent(t *testing.T, db persistence.Storage) {
	// Two jobs concurrently add the SAME brand-new tag name. Without
	// out-of-tx tag resolution this races on the UNIQUE(name) index and
	// one tx fails with a constraint violation. Neither caller should
	// observe such an error.
	const newTag = "shared-new-tag"

	tmp1 := newValidJob(defaultClientID)
	_, err := db.CreateWorkflow(t.Context(), tmp1.Workflow)
	require.NoError(t, err)
	job1, err := db.CreateJob(t.Context(), tmp1)
	require.NoError(t, err)

	tmp2 := newValidJob(defaultClientID + "-2")
	job2, err := db.CreateJob(t.Context(), tmp2)
	require.NoError(t, err)

	const writers = 8
	start := make(chan struct{})
	var wg sync.WaitGroup
	errs := make(chan error, writers)
	for i := range writers {
		j := job1
		if i%2 == 1 {
			j = job2
		}
		// Each writer adds the shared tag plus a unique one so CreateBulk
		// has multiple inserts to coordinate.
		add := []string{newTag, fmt.Sprintf("uniq-%d", i)}
		wg.Go(func() {
			<-start
			_, err := db.UpdateJob(t.Context(), j, persistence.JobUpdate{
				AddTags: &add,
			})
			errs <- err
		})
	}
	close(start)
	wg.Wait()
	close(errs)
	for err := range errs {
		// Concurrent updates on the SAME job will be rejected by the mtime
		// guard with errkind.TOCTOU; that is expected. What we assert here
		// is that no caller fails with a unique-constraint violation from
		// the global tag table.
		if err != nil {
			assert.NotContains(t, err.Error(), "UNIQUE constraint",
				"tag insertion must not race against UNIQUE(name)")
			assert.NotContains(t, err.Error(), "duplicate key",
				"tag insertion must not race against UNIQUE(name)")
		}
	}

	// Verify the shared tag is attached to both jobs.
	final1, err := db.GetJob(t.Context(), job1.ID, persistence.FetchParams{})
	require.NoError(t, err)
	final2, err := db.GetJob(t.Context(), job2.ID, persistence.FetchParams{})
	require.NoError(t, err)
	assert.Contains(t, *final1.Tags, newTag)
	assert.Contains(t, *final2.Tags, newTag)
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
