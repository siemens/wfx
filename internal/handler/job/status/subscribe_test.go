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
	"testing"

	"github.com/Southclaws/fault/ftag"
	"github.com/siemens/wfx/generated/model"
	"github.com/siemens/wfx/internal/workflow"
	"github.com/siemens/wfx/workflow/dau"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddSubscriberAndShutdown(t *testing.T) {
	db := newInMemoryDB(t)

	wf, err := db.CreateWorkflow(context.Background(), dau.PhasedWorkflow())
	require.NoError(t, err)

	job, err := db.CreateJob(context.Background(), &model.Job{
		ClientID: "klaus",
		Workflow: wf,
		Status:   &model.JobStatus{State: "CREATED"},
	})
	require.NoError(t, err)
	require.Equal(t, "CREATED", job.Status.State)
	require.False(t, workflow.IsTerminal(wf, job.Status.State))

	_, err = AddSubscriber(context.Background(), db, job.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, topics.Count())

	// test shutdown
	ShutdownSubscribers()
	assert.Equal(t, 0, topics.Count())
}

func TestAddSubscriber_NotFound(t *testing.T) {
	db := newInMemoryDB(t)
	_, err := AddSubscriber(context.Background(), db, "42")
	require.NotNil(t, err)
	assert.Equal(t, ftag.NotFound, ftag.Get(err))
}

func TestAddSubscriber_TerminalState(t *testing.T) {
	db := newInMemoryDB(t)
	wf, err := db.CreateWorkflow(context.Background(), dau.DirectWorkflow())
	require.NoError(t, err)

	job, err := db.CreateJob(context.Background(), &model.Job{
		ClientID: "klaus",
		Workflow: wf,
		Status:   &model.JobStatus{State: "ACTIVATED"},
	})
	require.NoError(t, err)

	_, err = AddSubscriber(context.Background(), db, job.ID)
	assert.ErrorContains(t, err, "attempted to subscribe to a job which is in a terminal state")
	assert.Equal(t, ftag.InvalidArgument, ftag.Get(err))
}

func TestCountSubscribers(t *testing.T) {
	db := newInMemoryDB(t)
	wf, err := db.CreateWorkflow(context.Background(), dau.PhasedWorkflow())
	require.NoError(t, err)

	job, err := db.CreateJob(context.Background(), &model.Job{
		ClientID: "klaus",
		Workflow: wf,
		Status:   &model.JobStatus{State: "CREATED"},
	})
	require.NoError(t, err)

	_, err = AddSubscriber(context.Background(), db, job.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, CountSubscribers())
	_, err = AddSubscriber(context.Background(), db, job.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, CountSubscribers())
}
