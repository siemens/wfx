//go:build testing

package tests

/*
 * SPDX-FileCopyrightText: 2026 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import (
	"context"
	"testing"

	"github.com/Southclaws/fault/ftag"
	"github.com/siemens/wfx/persistence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientIDContext(t *testing.T, db persistence.Storage) {
	fooJob := newValidJob("foo")
	_, err := db.CreateWorkflow(t.Context(), fooJob.Workflow)
	require.NoError(t, err)
	fooJob, err = db.CreateJob(t.Context(), fooJob)
	require.NoError(t, err)

	barJob := newValidJob("bar")
	barJob, err = db.CreateJob(t.Context(), barJob)
	require.NoError(t, err)

	ctx := persistence.WithClientID(t.Context(), "foo")
	got, err := db.GetJob(ctx, fooJob.ID, persistence.FetchParams{})
	require.NoError(t, err)
	assert.Equal(t, fooJob.ID, got.ID)

	_, err = db.GetJob(ctx, barJob.ID, persistence.FetchParams{})
	assert.Equal(t, ftag.NotFound, ftag.Get(err))

	jobs, err := db.QueryJobs(ctx, persistence.FilterParams{}, persistence.SortParams{}, defaultPaginationParams)
	require.NoError(t, err)
	require.Len(t, jobs.Content, 1)
	assert.Equal(t, fooJob.ID, jobs.Content[0].ID)

	bar := "bar"
	jobs, err = db.QueryJobs(ctx, persistence.FilterParams{ClientID: &bar}, persistence.SortParams{}, defaultPaginationParams)
	require.NoError(t, err)
	assert.Empty(t, jobs.Content)

	err = db.DeleteJob(ctx, barJob.ID)
	assert.Equal(t, ftag.NotFound, ftag.Get(err))

	unscoped, err := db.QueryJobs(context.Background(), persistence.FilterParams{}, persistence.SortParams{}, defaultPaginationParams)
	require.NoError(t, err)
	assert.Len(t, unscoped.Content, 2)
}
