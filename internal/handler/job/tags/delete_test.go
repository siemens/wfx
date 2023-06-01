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
