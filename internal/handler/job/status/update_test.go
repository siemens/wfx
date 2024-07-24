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

	"github.com/siemens/wfx/generated/api"
	"github.com/siemens/wfx/internal/handler/job/events"
	"github.com/siemens/wfx/persistence"
	"github.com/siemens/wfx/workflow/dau"
)

type testCase struct {
	from     string
	to       string
	eligible api.EligibleEnum
	expected string
}

func TestUpdateJob_Ok(t *testing.T) {
	tcs := []testCase{
		{from: "INSTALLING", to: "TERMINATED", eligible: api.CLIENT, expected: "TERMINATED"},
		{from: "ACTIVATING", to: "ACTIVATED", eligible: api.CLIENT, expected: "ACTIVATED"},
	}
	for _, tc := range tcs {
		t.Run(fmt.Sprintf("%v", tc), func(t *testing.T) {
			db := newInMemoryDB(t)
			wf := createDirectWorkflow(t, db)

			var jobID string
			{
				job, err := db.CreateJob(context.Background(), &api.Job{
					ClientID: "abc",
					Workflow: wf,
					Status:   &api.JobStatus{ClientID: "abc", State: tc.from},
				})
				jobID = job.ID
				assert.NoError(t, err)
				assert.Equal(t, tc.from, job.Status.State)
			}

			progress := int32(100)
			status, err := Update(context.Background(), db, jobID, &api.JobStatus{
				ClientID: "foo",
				State:    tc.to,
				Progress: &progress,
			}, tc.eligible)
			assert.NoError(t, err)
			assert.Equal(t, "foo", status.ClientID)
			assert.Equal(t, tc.expected, status.State)
			assert.Equal(t, int32(100), *status.Progress)
		})
	}
}

func TestUpdateJobStatus_Message(t *testing.T) {
	db := newInMemoryDB(t)
	wf := createDirectWorkflow(t, db)

	progress := int32(42)
	job, err := db.CreateJob(context.Background(), &api.Job{
		ClientID: "abc",
		Workflow: wf,
		Status:   &api.JobStatus{ClientID: "abc", State: "INSTALLING", Progress: &progress},
	})
	require.NoError(t, err)

	message := "Updating message!"

	status, err := Update(context.Background(), db, job.ID, &api.JobStatus{ClientID: "foo", Message: message, State: job.Status.State}, api.CLIENT)
	assert.NoError(t, err)
	assert.Equal(t, "INSTALLING", status.State)
	assert.Nil(t, status.Progress)
	assert.Equal(t, message, status.Message)
}

func TestUpdateJobStatus_StateWarp(t *testing.T) {
	from := "INSTALLING"
	to := "INSTALLED"
	source := api.CLIENT

	db := newInMemoryDB(t)
	wf := createDirectWorkflow(t, db)

	job, err := db.CreateJob(context.Background(), &api.Job{
		ClientID: "foo",
		Workflow: wf,
		Status:   &api.JobStatus{ClientID: "foo", State: from, DefinitionHash: "abc"},
	})
	require.NoError(t, err)

	updatedJob, err := db.UpdateJob(context.Background(), job, persistence.JobUpdate{Status: &api.JobStatus{
		ClientID:       "foo",
		State:          "INSTALLING",
		DefinitionHash: job.Status.DefinitionHash,
	}})
	assert.NoError(t, err)
	assert.NotEmpty(t, updatedJob.Status.DefinitionHash)

	progress := int32(100)
	status, err := Update(context.Background(), db, job.ID, &api.JobStatus{
		State:    to,
		Message:  "update installed",
		Progress: &progress,
	}, source)
	require.NoError(t, err)

	assert.Equal(t, "ACTIVATE", status.State)
	assert.Nil(t, status.Progress) // status was reset due to state warp
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
		job, err := db.CreateJob(context.Background(), &api.Job{
			ClientID: "abc",
			Workflow: wf,
			Status:   &api.JobStatus{State: from},
		})
		jobID = job.ID
		require.NoError(t, err)
		require.Equal(t, from, job.Status.State)
	}

	status, err := Update(context.Background(), db, jobID, &api.JobStatus{State: to}, api.WFX)
	assert.Error(t, err)
	assert.Nil(t, status)
}

func TestUpdateJob_NotifySubscribers(t *testing.T) {
	db := newInMemoryDB(t)
	wf := createDirectWorkflow(t, db)
	job, err := db.CreateJob(context.Background(), &api.Job{
		ClientID: "abc",
		Workflow: wf,
		Status:   &api.JobStatus{ClientID: "abc", State: "ACTIVATING"},
	})
	require.NoError(t, err)
	assert.Equal(t, "ACTIVATING", job.Status.State)

	ch, err := events.AddSubscriber(context.Background(), events.FilterParams{JobIDs: []string{job.ID}}, nil)
	require.NoError(t, err)

	progress := int32(100)
	_, err = Update(context.Background(), db, job.ID, &api.JobStatus{
		ClientID: "foo",
		State:    "ACTIVATED",
		Progress: &progress,
	}, api.CLIENT)
	require.NoError(t, err)

	event := <-ch
	receivedEvent := event.Args[0].(*events.JobEvent)
	assert.Equal(t, events.ActionUpdateStatus, receivedEvent.Action)
	assert.Equal(t, "ACTIVATED", receivedEvent.Job.Status.State)
	assert.Equal(t, wf.Name, receivedEvent.Job.Workflow.Name)
}

func createDirectWorkflow(t *testing.T, db persistence.Storage) *api.Workflow {
	wf, err := db.CreateWorkflow(context.Background(), dau.DirectWorkflow())
	require.NoError(t, err)
	return wf
}
