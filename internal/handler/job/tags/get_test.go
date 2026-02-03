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
	"testing"

	"github.com/siemens/wfx/generated/api"
	"github.com/siemens/wfx/workflow/dau"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGet(t *testing.T) {
	db := newInMemoryDB(t)

	wf, err := db.CreateWorkflow(context.Background(), dau.DirectWorkflow())
	require.NoError(t, err)
	tags := []string{"foo", "bar"}
	job, err := db.CreateJob(context.Background(), &api.Job{
		ClientID: "foo",
		Workflow: wf,
		Status:   &api.JobStatus{State: "CREATED"},
		Tags:     &tags,
	})
	require.NoError(t, err)

	tagList, err := Get(t.Context(), db, job.ID)
	require.NoError(t, err)
	require.NotNil(t, tagList)
	assert.Equal(t, []string{"bar", "foo"}, *tagList)
}

func TestGetEmpty(t *testing.T) {
	db := newInMemoryDB(t)

	wf, err := db.CreateWorkflow(context.Background(), dau.DirectWorkflow())
	require.NoError(t, err)
	job, err := db.CreateJob(context.Background(), &api.Job{
		ClientID: "foo",
		Workflow: wf,
		Status:   &api.JobStatus{State: "CREATED"},
	})
	require.NoError(t, err)

	tags, err := Get(context.Background(), db, job.ID)
	require.NoError(t, err)
	assert.Empty(t, tags)
}
