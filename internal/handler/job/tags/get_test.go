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

	"github.com/siemens/wfx/generated/model"
	"github.com/siemens/wfx/workflow/dau"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGet(t *testing.T) {
	db := newInMemoryDB(t)

	wf, err := db.CreateWorkflow(context.Background(), dau.DirectWorkflow())
	require.NoError(t, err)
	job, err := db.CreateJob(context.Background(), &model.Job{
		ClientID: "foo",
		Workflow: wf,
		Status:   &model.JobStatus{State: "CREATED"},
		Tags:     []string{"foo", "bar"},
	})
	require.NoError(t, err)

	tags, err := Get(context.Background(), db, job.ID)
	require.NoError(t, err)
	assert.Equal(t, []string{"bar", "foo"}, tags)
}

func TestGetEmpty(t *testing.T) {
	db := newInMemoryDB(t)

	wf, err := db.CreateWorkflow(context.Background(), dau.DirectWorkflow())
	require.NoError(t, err)
	job, err := db.CreateJob(context.Background(), &model.Job{
		ClientID: "foo",
		Workflow: wf,
		Status:   &model.JobStatus{State: "CREATED"},
	})
	require.NoError(t, err)

	tags, err := Get(context.Background(), db, job.ID)
	require.NoError(t, err)
	assert.Empty(t, tags)
}
