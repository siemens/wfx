package definition

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

	"github.com/Southclaws/fault/ftag"
	"github.com/siemens/wfx/generated/model"
	"github.com/siemens/wfx/persistence"
	"github.com/siemens/wfx/workflow/dau"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateJobDefinition(t *testing.T) {
	db := newInMemoryDB(t)

	wf, err := db.CreateWorkflow(context.Background(), dau.PhasedWorkflow())
	require.NoError(t, err)

	tmpJob := model.Job{
		ClientID: "abc",
		Workflow: wf,
		Status:   &model.JobStatus{ClientID: "abc", State: "CREATED"},
		Definition: map[string]any{
			"foo": "bar",
		},
	}
	tmpJob.Status.DefinitionHash = Hash(&tmpJob)

	job, err := db.CreateJob(context.Background(), &tmpJob)
	require.NoError(t, err)

	oldDefinitionHash := job.Status.DefinitionHash
	assert.NotEmpty(t, oldDefinitionHash)

	newDefinition := map[string]any{
		"foo": "baz",
	}
	definition, err := Update(context.Background(), db, job.ID, newDefinition)
	require.NoError(t, err)
	assert.Equal(t, "baz", definition["foo"])
	assert.Len(t, definition, 1)

	{
		job, err := db.GetJob(context.Background(), job.ID, persistence.FetchParams{})
		assert.NoError(t, err)
		assert.NotEqual(t, oldDefinitionHash, job.Status.DefinitionHash)
	}
}

func TestUpdateJobDefinition_NotFound(t *testing.T) {
	db := newInMemoryDB(t)
	definition, err := Update(context.Background(), db, "999999", map[string]any{})
	assert.Nil(t, definition)
	assert.Equal(t, ftag.NotFound, ftag.Get(err))
}
