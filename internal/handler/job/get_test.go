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
	"testing"
	"time"

	"github.com/Southclaws/fault/ftag"
	"github.com/siemens/wfx/generated/model"
	"github.com/stretchr/testify/assert"
)

func TestGetJob(t *testing.T) {
	db := newInMemoryDB(t)
	wf := createDirectWorkflow(t, db)

	now := time.Now()

	var jobID string
	{
		job, err := CreateJob(context.Background(), db, &model.JobRequest{
			ClientID:   "klaus",
			Workflow:   wf.Name,
			Definition: map[string]interface{}{"foo": "bar"},
		})
		assert.NoError(t, err)
		jobID = job.ID
	}

	job, err := GetJob(context.Background(), db, jobID, false)
	assert.NoError(t, err)
	assert.Equal(t, jobID, job.ID)
	assert.Equal(t, job.Mtime, job.Stime)
	assert.GreaterOrEqual(t, time.Time(job.Mtime).UnixMicro(), now.UnixMicro())
	assert.Equal(t, "adc1cfc1577119ba2a0852133340088390c1103bdf82d8102970d3e6c53ec10b", job.Status.DefinitionHash)
}

func TestGetJob_NotFound(t *testing.T) {
	db := newInMemoryDB(t)
	job, err := GetJob(context.Background(), db, "1", false)
	assert.Nil(t, job)
	assert.Equal(t, ftag.NotFound, ftag.Get(err))
}
