package status

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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/siemens/wfx/generated/model"
	"github.com/siemens/wfx/persistence"
	"github.com/siemens/wfx/workflow/dau"
)

type testCase struct {
	from     string
	to       string
	eligible model.EligibleEnum
	expected string
}

func TestUpdateJob_Ok(t *testing.T) {
	tcs := []testCase{
		{from: "INSTALLING", to: "TERMINATED", eligible: model.EligibleEnumCLIENT, expected: "TERMINATED"},
		{from: "ACTIVATING", to: "ACTIVATED", eligible: model.EligibleEnumCLIENT, expected: "ACTIVATED"},
	}
	for _, tc := range tcs {
		t.Run(fmt.Sprintf("%v", tc), func(t *testing.T) {
			db := newInMemoryDB(t)
			wf := createDirectWorkflow(t, db)

			var jobID string
			{
				job, err2 := db.CreateJob(context.Background(), &model.Job{
					ClientID: "abc",
					Workflow: wf,
					Status:   &model.JobStatus{ClientID: "abc", State: tc.from},
				})
				jobID = job.ID
				assert.NoError(t, err2)
				assert.Equal(t, tc.from, job.Status.State)
			}

			status, err := Update(context.Background(), db, jobID, &model.JobStatus{
				ClientID: "klaus",
				State:    tc.to,
				Progress: 100,
			}, tc.eligible)
			assert.NoError(t, err)
			assert.Equal(t, "klaus", status.ClientID)
			assert.Equal(t, tc.expected, status.State)
			assert.Equal(t, int32(100), status.Progress)
		})
	}
}

func TestUpdateJobStatus_Message(t *testing.T) {
	db := newInMemoryDB(t)
	wf := createDirectWorkflow(t, db)

	job, err := db.CreateJob(context.Background(), &model.Job{
		ClientID: "abc",
		Workflow: wf,
		Status:   &model.JobStatus{ClientID: "abc", State: "INSTALLING", Progress: 42},
	})
	require.NoError(t, err)

	message := "Updating message!"

	status, err := Update(context.Background(), db, job.ID, &model.JobStatus{ClientID: "klaus", Message: message, State: job.Status.State}, model.EligibleEnumCLIENT)
	assert.NoError(t, err)
	assert.Equal(t, "INSTALLING", status.State)
	assert.Equal(t, int32(0), status.Progress)
	assert.Equal(t, message, status.Message)
}

func TestUpdateJobStatus_StateWarp(t *testing.T) {
	from := "INSTALLING"
	to := "INSTALLED"
	source := model.EligibleEnumCLIENT
	expected := "ACTIVATE"

	db := newInMemoryDB(t)
	wf := createDirectWorkflow(t, db)

	job, err := db.CreateJob(context.Background(), &model.Job{
		ClientID: "klaus",
		Workflow: wf,
		Status:   &model.JobStatus{ClientID: "klaus", State: from, DefinitionHash: "abc"},
	})
	require.NoError(t, err)

	updatedJob, err := db.UpdateJob(context.Background(), job, persistence.JobUpdate{Status: &model.JobStatus{
		ClientID:       "klaus",
		State:          "INSTALLING",
		DefinitionHash: job.Status.DefinitionHash,
	}})
	assert.NoError(t, err)
	assert.NotEmpty(t, updatedJob.Status.DefinitionHash)

	status, err := Update(context.Background(), db, job.ID, &model.JobStatus{
		State:    to,
		Message:  "update installed",
		Progress: 100,
	}, source)
	assert.NoError(t, err)

	assert.Equal(t, expected, status.State)
	assert.Equal(t, int32(0), status.Progress)
	assert.Equal(t, "", status.Message)
	assert.Empty(t, status.Context)
	assert.Equal(t, job.Status.DefinitionHash, status.DefinitionHash)
}

func TestUpdateJobStatusNotAllowed(t *testing.T) {
	from := "ACTIVATING"
	to := "ACTIVATED"
	db := newInMemoryDB(t)
	wf := createDirectWorkflow(t, db)
	var jobID string
	{
		job, err2 := db.CreateJob(context.Background(), &model.Job{
			ClientID: "abc",
			Workflow: wf,
			Status:   &model.JobStatus{State: from},
		})
		jobID = job.ID
		require.NoError(t, err2)
		require.Equal(t, from, job.Status.State)
	}

	status, err := Update(context.Background(), db, jobID, &model.JobStatus{State: to}, model.EligibleEnumWFX)
	assert.Error(t, err)
	assert.Nil(t, status)
}

func createDirectWorkflow(t *testing.T, db persistence.Storage) *model.Workflow {
	wf, err := db.CreateWorkflow(context.Background(), dau.DirectWorkflow())
	require.NoError(t, err)
	return wf
}
