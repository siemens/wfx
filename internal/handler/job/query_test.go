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

	"github.com/siemens/wfx/generated/api"
	"github.com/siemens/wfx/persistence"
	"github.com/siemens/wfx/workflow/dau"
	"github.com/stretchr/testify/assert"
)

func TestQueryJobs(t *testing.T) {
	db := newInMemoryDB(t)
	createDirectWorkflow(t, db)

	tmpJob := newValidJob("abc", "INSTALLING")
	job, err := db.CreateJob(context.Background(), &tmpJob)
	assert.NoError(t, err)

	list, err := QueryJobs(context.Background(), db, persistence.FilterParams{}, persistence.PaginationParams{Limit: 10}, nil)
	assert.NoError(t, err)
	assert.Len(t, list.Content, 1)
	assert.Equal(t, job.ID, list.Content[0].ID)
}

func TestQueryJobs_Empty(t *testing.T) {
	db := newInMemoryDB(t)
	list, err := QueryJobs(context.Background(), db, persistence.FilterParams{}, persistence.PaginationParams{Limit: 10}, nil)
	assert.NoError(t, err)
	assert.Empty(t, list.Content)
}

func TestParseSortParamAsc(t *testing.T) {
	sp := parseSortParam("asc")
	assert.Equal(t, false, sp.Desc)
}

func TestParseSortParamDesc(t *testing.T) {
	sp := parseSortParam("desc")
	assert.Equal(t, true, sp.Desc)
}

func newValidJob(clientID, state string) api.Job {
	wf := dau.DirectWorkflow()
	return api.Job{
		ClientID: clientID,
		Status: &api.JobStatus{
			State: state,
		},
		Workflow: wf,
	}
}
