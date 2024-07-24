//go:build testing

package tests

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
	"strconv"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/siemens/wfx/generated/api"
	"github.com/siemens/wfx/persistence"
	"github.com/siemens/wfx/workflow/dau"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	defaultPaginationParams = persistence.PaginationParams{Offset: 0, Limit: 10}
	sortAsc                 = persistence.SortParams{Desc: false}
)

func TestQueryJobsFilter(t *testing.T, db persistence.Storage) {
	installedState := "INSTALLED"
	installState := "INSTALL"
	activatedState := "ACTIVATED"

	clientID := "42"
	wf := dau.DirectWorkflow()
	_, err := db.CreateWorkflow(context.Background(), wf)
	require.NoError(t, err)

	now := time.Now()

	tags := []string{"bar", "foo"}
	firstJob, err := db.CreateJob(context.Background(), &api.Job{
		ClientID: clientID,
		Workflow: wf,
		Status: &api.JobStatus{
			State: installedState,
		},
		Tags: tags,
	})
	require.NoError(t, err)
	assert.NotNil(t, firstJob.Stime)
	assert.NotNil(t, firstJob.Mtime)
	assert.True(t, time.Time(*firstJob.Stime).After(now) || time.Time(*firstJob.Stime).Equal(now))
	assert.True(t, time.Time(*firstJob.Mtime).After(now) || time.Time(*firstJob.Mtime).Equal(now))

	secondStime := now.Add(time.Second)
	secondJob, err := db.CreateJob(context.Background(), &api.Job{
		ClientID: clientID,
		Workflow: wf,
		Status: &api.JobStatus{
			State: installState,
		},
		Stime: &secondStime,
	})
	require.NoError(t, err)

	thirdStime := now.Add(2 * time.Second)
	thirdJob, err := db.CreateJob(context.Background(), &api.Job{
		ClientID: clientID,
		Workflow: wf,
		Status: &api.JobStatus{
			State: activatedState,
		},
		Tags:  []string{"meh"},
		Stime: &thirdStime,
	})
	require.NoError(t, err)

	{ // filter by group
		result, err := db.QueryJobs(context.Background(), persistence.FilterParams{ClientID: &clientID, Group: []string{"OPEN"}},
			sortAsc, defaultPaginationParams)
		actual := result.Content
		require.NoError(t, err)
		assert.Len(t, actual, 2)
		assert.Equal(t, firstJob.ID, actual[0].ID)
		assert.Equal(t, firstJob.Tags, actual[0].Tags)
		assert.Equal(t, secondJob.ID, actual[1].ID)
	}

	{ // filter by group
		result, err := db.QueryJobs(context.Background(), persistence.FilterParams{ClientID: &clientID, Group: []string{"CLOSED"}},
			sortAsc, defaultPaginationParams)
		actual := result.Content
		require.NoError(t, err)
		assert.Len(t, actual, 1)
		assert.Equal(t, thirdJob.ID, actual[0].ID)
	}

	{ // filter by group
		result, err := db.QueryJobs(context.Background(), persistence.FilterParams{ClientID: &clientID, Group: []string{"OPEN", "CLOSED"}},
			sortAsc, defaultPaginationParams)
		actual := result.Content
		require.NoError(t, err)
		assert.Len(t, actual, 3)
		assert.Equal(t, firstJob.ID, actual[0].ID)
		assert.Equal(t, secondJob.ID, actual[1].ID)
		assert.Equal(t, thirdJob.ID, actual[2].ID)
	}

	{
		result, err := db.QueryJobs(context.Background(), persistence.FilterParams{ClientID: &clientID, State: &installedState},
			sortAsc, defaultPaginationParams)
		actual := result.Content
		require.NoError(t, err)
		assert.Len(t, actual, 1)
		assert.Equal(t, firstJob.ID, actual[0].ID)
	}

	{
		// filter by name
		result, err := db.QueryJobs(context.Background(), persistence.FilterParams{ClientID: &clientID, Workflow: &wf.Name},
			sortAsc, defaultPaginationParams)
		actual := result.Content
		require.NoError(t, err)
		assert.Len(t, actual, 3)
		assert.Equal(t, []string{firstJob.ID, secondJob.ID, thirdJob.ID}, []string{actual[0].ID, actual[1].ID, actual[2].ID})

		doesNotExist := "doesNotExist"
		result, err = db.QueryJobs(context.Background(), persistence.FilterParams{ClientID: &clientID, Workflow: &doesNotExist},
			sortAsc, defaultPaginationParams)
		actual = result.Content
		require.NoError(t, err)
		assert.Len(t, actual, 0)
	}

	// filter by tags
	{
		result, err := db.QueryJobs(context.Background(), persistence.FilterParams{Tags: []string{"foo"}},
			sortAsc, defaultPaginationParams)
		actual := result.Content
		require.NoError(t, err)
		assert.Len(t, actual, 1)
		assert.Equal(t, firstJob.ID, actual[0].ID)
	}
	{
		result, err := db.QueryJobs(context.Background(), persistence.FilterParams{Tags: []string{"bar"}},
			sortAsc, defaultPaginationParams)
		actual := result.Content
		require.NoError(t, err)
		assert.Len(t, actual, 1)
		assert.Equal(t, firstJob.ID, actual[0].ID)
	}
	{
		result, err := db.QueryJobs(context.Background(), persistence.FilterParams{Tags: []string{"foo", "bar"}},
			sortAsc, defaultPaginationParams)
		actual := result.Content
		require.NoError(t, err)
		assert.Len(t, actual, 1)
		assert.Equal(t, firstJob.ID, actual[0].ID)
	}
}

func TestGetJobsSorted(t *testing.T, db persistence.Storage) {
	clientID := "my_client"

	var first, second, third *api.Job
	var err error

	{
		tmp := newValidJob(clientID)
		_, err := db.CreateWorkflow(context.Background(), tmp.Workflow)
		require.NoError(t, err)

		stime := time.Now().Add(-2 * time.Minute)
		tmp.Stime = &stime
		first, err = db.CreateJob(context.Background(), tmp)
		require.NoError(t, err)
	}

	{
		tmp := newValidJob(clientID)
		mtime := time.Now().Add(-time.Minute)
		tmp.Mtime = &mtime
		second, err = db.CreateJob(context.Background(), tmp)
		require.NoError(t, err)
	}

	{
		tmp := newValidJob(clientID)
		third, err = db.CreateJob(context.Background(), tmp)
		require.NoError(t, err)
	}

	{
		result, err := db.QueryJobs(
			context.Background(),
			persistence.FilterParams{ClientID: &clientID},
			persistence.SortParams{Desc: false},
			defaultPaginationParams,
		)
		actual := result.Content
		require.NoError(t, err)
		assert.Len(t, actual, 3)
		// oldest first
		assert.Equal(t, []string{first.ID, second.ID, third.ID}, []string{actual[0].ID, actual[1].ID, actual[2].ID})
	}

	{
		result, err := db.QueryJobs(
			context.Background(),
			persistence.FilterParams{ClientID: &clientID},
			persistence.SortParams{Desc: true},
			defaultPaginationParams,
		)
		actual := result.Content
		require.NoError(t, err)
		assert.Len(t, actual, 3)
		// newest first
		assert.Equal(t, []string{third.ID, second.ID, first.ID}, []string{actual[0].ID, actual[1].ID, actual[2].ID})
	}
}

func TestGetJobMaxHistorySize(t *testing.T, db persistence.Storage) {
	// don't spam the logs
	oldLevel := zerolog.GlobalLevel()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	defer zerolog.SetGlobalLevel(oldLevel)

	maxEntries := 8192

	var jobID string

	{
		// job which we are going to update often
		tmp := newValidJob("foo")
		message := "0"
		tmp.Status.Message = message
		_, err := db.CreateWorkflow(context.Background(), tmp.Workflow)
		require.NoError(t, err)
		job, err := db.CreateJob(context.Background(), tmp)
		require.NoError(t, err)
		jobID = job.ID
	}

	{
		job, err := db.GetJob(context.Background(), jobID, persistence.FetchParams{History: true})
		require.NoError(t, err)
		require.Empty(t, job.History)
		require.Equal(t, job.Status.Message, "0")

		// update the job n+1 times
		for i := 1; i <= maxEntries+1; i++ {
			msg := fmt.Sprintf("%d", i)
			job, err = db.UpdateJob(context.Background(), job, persistence.JobUpdate{
				Status: &api.JobStatus{
					State:   job.Status.State,
					Message: msg,
				},
			})
			require.NoError(t, err)
			require.Equal(t, msg, job.Status.Message, "Message must be updated")
		}
	}

	job, err := db.GetJob(context.Background(), jobID, persistence.FetchParams{History: true})
	require.NoError(t, err)
	require.NotNil(t, job.History)
	require.Len(t, *job.History, maxEntries)

	history := *job.History
	ids := make([]int, 0, len(history))
	for _, hist := range history {
		id, _ := strconv.Atoi(hist.Status.Message)
		ids = append(ids, id)
	}
	t.Log(ids)
	assert.IsDecreasing(t, ids)

	assert.Equal(t, fmt.Sprintf("%d", maxEntries), history[0].Status.Message)
	assert.Equal(t, "1", history[len(history)-1].Status.Message)
}

func TestJobsPagination(t *testing.T, db persistence.Storage) {
	filterParams := persistence.FilterParams{ClientID: &defaultClientID}
	total := 5
	var ids []string

	_, err := db.CreateWorkflow(context.Background(), dau.DirectWorkflow())
	assert.NoError(t, err)
	for i := 0; i < total; i++ {
		tmp := newValidJob(*filterParams.ClientID)
		job, err := db.CreateJob(context.Background(), tmp)
		require.NoError(t, err)
		ids = append(ids, job.ID)
	}

	{
		result, err := db.QueryJobs(context.Background(), filterParams, sortAsc, persistence.PaginationParams{Offset: 0, Limit: 2})
		assert.NoError(t, err)
		assert.Equal(t, int64(0), result.Pagination.Offset)
		assert.Equal(t, int32(2), result.Pagination.Limit)
		assert.Equal(t, int64(total), result.Pagination.Total)
		assert.Len(t, result.Content, 2)
		assert.Equal(t, []string{ids[0], ids[1]}, []string{result.Content[0].ID, result.Content[1].ID})
	}
	{
		result, err := db.QueryJobs(context.Background(), filterParams, sortAsc, persistence.PaginationParams{Offset: 1, Limit: 2})
		assert.NoError(t, err)
		assert.Equal(t, int64(1), result.Pagination.Offset)
		assert.Equal(t, int32(2), result.Pagination.Limit)
		assert.Equal(t, int64(total), result.Pagination.Total)
		assert.Len(t, result.Content, 2)
		assert.Equal(t, ids[1:3], []string{result.Content[0].ID, result.Content[1].ID})
	}
	{
		result, err := db.QueryJobs(context.Background(), filterParams, sortAsc, persistence.PaginationParams{Offset: 2, Limit: 2})
		assert.NoError(t, err)
		assert.Equal(t, int64(2), result.Pagination.Offset)
		assert.Equal(t, int32(2), result.Pagination.Limit)
		assert.Equal(t, int64(total), result.Pagination.Total)
		assert.Len(t, result.Content, 2)
		assert.Equal(t, ids[2:4], []string{result.Content[0].ID, result.Content[1].ID})
	}
	{ // one past the last page
		result, err := db.QueryJobs(context.Background(), filterParams, sortAsc, persistence.PaginationParams{Offset: 6, Limit: 2})
		assert.NoError(t, err)
		assert.Equal(t, int64(6), result.Pagination.Offset)
		assert.Equal(t, int32(2), result.Pagination.Limit)
		assert.Equal(t, int64(total), result.Pagination.Total)
		assert.Len(t, result.Content, 0)
	}
}
