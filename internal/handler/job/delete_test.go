package job

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"context"
	"errors"
	"testing"

	"github.com/Southclaws/fault/ftag"
	"github.com/siemens/wfx/generated/api"
	"github.com/siemens/wfx/internal/handler/job/events"
	"github.com/siemens/wfx/persistence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeleteJob(t *testing.T) {
	db := newInMemoryDB(t)
	createDirectWorkflow(t, db)

	tmpJob := newValidJob("abc", "INSTALLING")
	job, err := db.CreateJob(context.Background(), &tmpJob)
	require.NoError(t, err)

	ch, err := events.AddSubscriber(context.Background(), events.FilterParams{}, nil)
	require.NoError(t, err)

	err = DeleteJob(context.Background(), db, job.ID)
	require.NoError(t, err)

	ev := <-ch
	jobEvent := ev.Args[0].(*events.JobEvent)
	assert.Equal(t, events.ActionDelete, jobEvent.Action)
	assert.Equal(t, job.ID, jobEvent.Job.ID)
}

func TestDeleteJob_NotFound(t *testing.T) {
	db := newInMemoryDB(t)
	err := DeleteJob(context.Background(), db, "42")
	assert.Equal(t, ftag.NotFound, ftag.Get(err))
}

func TestDeleteJob_Error(t *testing.T) {
	dbMock := persistence.NewHealthyMockStorage(t)
	ctx := context.Background()
	jobID := "42"
	dbMock.EXPECT().GetJob(ctx, jobID, persistence.FetchParams{History: false}).Return(&api.Job{ID: jobID}, nil)
	dbMock.EXPECT().DeleteJob(ctx, jobID).Return(errors.New("something went wrong"))
	err := DeleteJob(ctx, dbMock, jobID)
	assert.Equal(t, ftag.Internal, ftag.Get(err))
}
