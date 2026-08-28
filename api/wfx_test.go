package api

/*
 * SPDX-FileCopyrightText: 2024 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alexliesenfeld/health"
	"github.com/siemens/wfx/cmd/wfx/cmd/config"
	"github.com/siemens/wfx/generated/api"
	"github.com/siemens/wfx/internal/handler/job/events"
	"github.com/siemens/wfx/internal/persistence/entgo"
	"github.com/siemens/wfx/middleware/sse"
	"github.com/siemens/wfx/persistence"
	"github.com/siemens/wfx/workflow/dau"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatusListener(*testing.T) {
	healthStatusListener(context.Background(), health.CheckerState{Status: health.StatusUp})
	healthStatusListener(context.Background(), health.CheckerState{Status: health.StatusDown})
	healthStatusListener(context.Background(), health.CheckerState{Status: health.StatusUnknown})
	healthStatusListener(context.Background(), health.CheckerState{
		Status: health.StatusUp,
		CheckState: map[string]health.CheckState{
			"db": {Status: health.StatusUp},
		},
	})
}

func TestJQOptsProvider(t *testing.T) {
	opts := config.JQOpts{FilterMaxResponseSize: 1}
	wfx := NewWfxServer(persistence.NewHealthyMockStorage(t)).WithJQOpts(func() config.JQOpts { return opts })

	assert.Equal(t, 1, wfx.jqOpts().FilterMaxResponseSize)
	opts.FilterMaxResponseSize = 2
	assert.Equal(t, 2, wfx.jqOpts().FilterMaxResponseSize)
}

func TestSSEOptsProvider(t *testing.T) {
	opts := SSEOpts{PingInterval: time.Second}
	wfx := NewWfxServer(persistence.NewHealthyMockStorage(t)).WithSSEOpts(func() SSEOpts { return opts })

	assert.Equal(t, time.Second, wfx.sseOpts().PingInterval)
	opts.PingInterval = 2 * time.Second
	assert.Equal(t, 2*time.Second, wfx.sseOpts().PingInterval)
}

func TestGetJobsEvents(t *testing.T) {
	t.Cleanup(events.ShutdownSubscribers)
	jobIDs := "1,2,3"
	clientIDs := "4,5,6"
	workflows := "wf1,wf2"
	tags := "tag1,tag2"

	request := api.GetJobsEventsRequestObject{
		Params: api.GetJobsEventsParams{
			JobIds:    &jobIDs,
			ClientIDs: &clientIDs,
			Workflows: &workflows,
			Tags:      &tags,
		},
	}

	wfx := NewWfxServer(persistence.NewHealthyMockStorage(t))
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	response, err := wfx.GetJobsEvents(ctx, request)
	require.NoError(t, err)
	assert.NotNil(t, response)
}

func TestGetJobsClientIDHeader(t *testing.T) {
	db := newSQLiteStorage(t)
	wfx := NewWfxServer(db)
	persistAPIJob(t, db, "foo")
	persistAPIJob(t, db, "bar")

	t.Run("header filters jobs", func(t *testing.T) {
		clientID := "foo"
		response, err := wfx.GetJobs(t.Context(), api.GetJobsRequestObject{
			Params: api.GetJobsParams{XClientID: &clientID},
		})
		require.NoError(t, err)
		jobs := api.PaginatedJobList(response.(api.GetJobs200JSONResponse))
		require.Len(t, jobs.Content, 1)
		assert.Equal(t, clientID, jobs.Content[0].ClientID)
	})

	t.Run("matching query and header", func(t *testing.T) {
		clientID := "foo"
		response, err := wfx.GetJobs(t.Context(), api.GetJobsRequestObject{
			Params: api.GetJobsParams{ParamClientID: &clientID, XClientID: &clientID},
		})
		require.NoError(t, err)
		jobs := api.PaginatedJobList(response.(api.GetJobs200JSONResponse))
		require.Len(t, jobs.Content, 1)
		assert.Equal(t, clientID, jobs.Content[0].ClientID)
	})

	t.Run("mismatched query and header", func(t *testing.T) {
		clientID, queryClientID := "foo", "bar"
		response, err := wfx.GetJobs(t.Context(), api.GetJobsRequestObject{
			Params: api.GetJobsParams{ParamClientID: &queryClientID, XClientID: &clientID},
		})
		require.NoError(t, err)
		apiError := api.ErrorResponse(response.(api.GetJobs400JSONResponse))
		require.NotNil(t, apiError.Errors)
		assert.Equal(t, ClientIDMismatch, (*apiError.Errors)[0])
	})
}

func TestGetJobsEventsClientIDHeader(t *testing.T) {
	clientID := "foo"
	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)
	t.Cleanup(events.ShutdownSubscribers)

	wfx := NewWfxServer(persistence.NewHealthyMockStorage(t))
	response, err := wfx.GetJobsEvents(ctx, api.GetJobsEventsRequestObject{
		Params: api.GetJobsEventsParams{XClientID: &clientID},
	})
	require.NoError(t, err)
	require.IsType(t, sse.Responder{}, response)
	require.Equal(t, 1, events.SubscriberCount())

	events.PublishEvent(ctx, events.JobEvent{Job: &api.Job{ID: "bar", ClientID: "bar"}})
	events.PublishEvent(ctx, events.JobEvent{Job: &api.Job{ID: "foo", ClientID: clientID}})

	recorder := sse.NewMockResponseRecorder(t)
	done := make(chan error, 1)
	go func() { done <- response.VisitGetJobsEventsResponse(recorder) }()
	require.Eventually(t, func() bool { return strings.Contains(recorder.Response(), `"id":"foo"`) }, time.Second, time.Millisecond)
	assert.NotContains(t, recorder.Response(), `"id":"bar"`)
	cancel()
	require.NoError(t, <-done)
}

func TestJobReadsClientIDHeader(t *testing.T) {
	db := newSQLiteStorage(t)
	wfx := NewWfxServer(db)
	job := persistAPIJob(t, db, "foo")
	otherClientID := "bar"

	tests := []struct {
		name string
		get  func() (any, error)
	}{
		{
			name: "job",
			get: func() (any, error) {
				return wfx.GetJobsId(t.Context(), api.GetJobsIdRequestObject{
					Id: job.ID, Params: api.GetJobsIdParams{XClientID: &otherClientID},
				})
			},
		},
		{
			name: "definition",
			get: func() (any, error) {
				return wfx.GetJobsIdDefinition(t.Context(), api.GetJobsIdDefinitionRequestObject{
					Id: job.ID, Params: api.GetJobsIdDefinitionParams{XClientID: &otherClientID},
				})
			},
		},
		{
			name: "status",
			get: func() (any, error) {
				return wfx.GetJobsIdStatus(t.Context(), api.GetJobsIdStatusRequestObject{
					Id: job.ID, Params: api.GetJobsIdStatusParams{XClientID: &otherClientID},
				})
			},
		},
		{
			name: "tags",
			get: func() (any, error) {
				return wfx.GetJobsIdTags(t.Context(), api.GetJobsIdTagsRequestObject{
					Id: job.ID, Params: api.GetJobsIdTagsParams{XClientID: &otherClientID},
				})
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			response, err := tc.get()
			require.NoError(t, err)
			switch response := response.(type) {
			case api.GetJobsId404JSONResponse:
				assertJobNotFound(t, api.ErrorResponse(response))
			case api.GetJobsIdDefinition404JSONResponse:
				assertJobNotFound(t, api.ErrorResponse(response))
			case api.GetJobsIdStatus404JSONResponse:
				assertJobNotFound(t, api.ErrorResponse(response))
			case api.GetJobsIdTags404JSONResponse:
				assertJobNotFound(t, api.ErrorResponse(response))
			default:
				t.Fatalf("unexpected response type %T", response)
			}
		})
	}

	clientID := job.ClientID
	response, err := wfx.GetJobsId(t.Context(), api.GetJobsIdRequestObject{
		Id: job.ID, Params: api.GetJobsIdParams{XClientID: &clientID},
	})
	require.NoError(t, err)
	assert.Equal(t, job.ID, api.Job(response.(api.GetJobsId200JSONResponse)).ID)

	definitionResponse, err := wfx.GetJobsIdDefinition(t.Context(), api.GetJobsIdDefinitionRequestObject{
		Id: job.ID, Params: api.GetJobsIdDefinitionParams{XClientID: &clientID},
	})
	require.NoError(t, err)
	assert.Equal(t, job.Definition, map[string]any(definitionResponse.(api.GetJobsIdDefinition200JSONResponse)))

	statusResponse, err := wfx.GetJobsIdStatus(t.Context(), api.GetJobsIdStatusRequestObject{
		Id: job.ID, Params: api.GetJobsIdStatusParams{XClientID: &clientID},
	})
	require.NoError(t, err)
	assert.Equal(t, *job.Status, api.JobStatus(statusResponse.(api.GetJobsIdStatus200JSONResponse)))

	tagsResponse, err := wfx.GetJobsIdTags(t.Context(), api.GetJobsIdTagsRequestObject{
		Id: job.ID, Params: api.GetJobsIdTagsParams{XClientID: &clientID},
	})
	require.NoError(t, err)
	assert.Equal(t, *job.Tags, []string(tagsResponse.(api.GetJobsIdTags200JSONResponse)))
}

func TestJobReadsClientIDHeaderWithResponseFilter(t *testing.T) {
	db := newSQLiteStorage(t)
	wfx := NewWfxServer(db)
	job := persistAPIJob(t, db, "foo")
	clientID, filter := job.ClientID, "."

	definitionResponse, err := wfx.GetJobsIdDefinition(t.Context(), api.GetJobsIdDefinitionRequestObject{
		Id: job.ID, Params: api.GetJobsIdDefinitionParams{XClientID: &clientID, XResponseFilter: &filter},
	})
	require.NoError(t, err)
	assert.Equal(t, job.Definition, definitionResponse.(JQFilter).body)

	statusResponse, err := wfx.GetJobsIdStatus(t.Context(), api.GetJobsIdStatusRequestObject{
		Id: job.ID, Params: api.GetJobsIdStatusParams{XClientID: &clientID, XResponseFilter: &filter},
	})
	require.NoError(t, err)
	assert.Equal(t, *job.Status, statusResponse.(JQFilter).body)

	tagsResponse, err := wfx.GetJobsIdTags(t.Context(), api.GetJobsIdTagsRequestObject{
		Id: job.ID, Params: api.GetJobsIdTagsParams{XClientID: &clientID, XResponseFilter: &filter},
	})
	require.NoError(t, err)
	assert.Equal(t, job.Tags, tagsResponse.(JQFilter).body)
}

func TestPutJobsIdStatusClientIDHeader(t *testing.T) {
	db := newSQLiteStorage(t)
	wfx := NewWfxServer(db)
	job := persistAPIJob(t, db, "foo")
	ctx := context.WithValue(t.Context(), EligibleKey, api.CLIENT)

	tests := []struct {
		name       string
		headerID   string
		statusID   string
		wantStatus bool
	}{
		{name: "matching IDs", headerID: "foo", statusID: "foo", wantStatus: true},
		{name: "header differs from status", headerID: "foo", statusID: "bar"},
		{name: "header differs from job", headerID: "bar", statusID: "bar"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			response, err := wfx.PutJobsIdStatus(ctx, api.PutJobsIdStatusRequestObject{
				Id:     job.ID,
				Params: api.PutJobsIdStatusParams{XClientID: &tc.headerID},
				Body:   &api.JobStatus{ClientID: tc.statusID, State: "INSTALL"},
			})
			require.NoError(t, err)
			if tc.wantStatus {
				assert.IsType(t, api.PutJobsIdStatus200JSONResponse{}, response)
				return
			}
			apiError := api.ErrorResponse(response.(api.PutJobsIdStatus404JSONResponse))
			assertJobNotFound(t, apiError)
		})
	}
}

func TestPutJobsIdStatusClientIDHeaderMissingJob(t *testing.T) {
	clientID := "foo"
	status := api.JobStatus{ClientID: clientID, State: "INSTALL"}
	ctx := context.WithValue(t.Context(), EligibleKey, api.CLIENT)
	response, err := NewWfxServer(newSQLiteStorage(t)).PutJobsIdStatus(ctx, api.PutJobsIdStatusRequestObject{
		Id: "missing", Params: api.PutJobsIdStatusParams{XClientID: &clientID}, Body: &status,
	})
	require.NoError(t, err)
	assertJobNotFound(t, api.ErrorResponse(response.(api.PutJobsIdStatus404JSONResponse)))
}

func TestPutJobsIdStatusConcurrent(t *testing.T) {
	db := newSQLiteStorage(t)
	wfx := NewWfxServer(db)

	const workers = 10

	// Build a custom workflow: START -> T0..T(workers-1), all triggered by CLIENT.
	// There are no transitions out of any T_i, so only one goroutine can succeed.
	wfDef := &api.Workflow{Name: "concurrent.test"}
	wfDef.States = append(wfDef.States, api.State{Name: "START"})
	for i := range workers {
		target := fmt.Sprintf("T%d", i)
		wfDef.States = append(wfDef.States, api.State{Name: target})
		wfDef.Transitions = append(wfDef.Transitions, api.Transition{
			From:     "START",
			To:       target,
			Eligible: api.CLIENT,
		})
	}
	wf, err := db.CreateWorkflow(t.Context(), wfDef)
	require.NoError(t, err)

	job, err := db.CreateJob(t.Context(), &api.Job{
		ClientID: "foo",
		Workflow: wf,
		Status:   &api.JobStatus{State: "START"},
	})
	require.NoError(t, err)

	var wg sync.WaitGroup
	var successes int32
	var rejected int32
	start := make(chan struct{})

	ctx := context.WithValue(t.Context(), EligibleKey, api.CLIENT)

	for i := range workers {
		body := api.PutJobsIdStatusJSONRequestBody{State: fmt.Sprintf("T%d", i)}
		req := api.PutJobsIdStatusRequestObject{
			Id:   job.ID,
			Body: &body,
		}
		wg.Go(func() {
			<-start
			resp, err := wfx.PutJobsIdStatus(ctx, req)
			require.NoError(t, err)
			switch resp.(type) {
			case api.PutJobsIdStatus200JSONResponse:
				atomic.AddInt32(&successes, 1)
			case api.PutJobsIdStatus400JSONResponse:
				atomic.AddInt32(&rejected, 1)
			}
		})
	}

	close(start)
	wg.Wait()

	assert.Equal(t, int32(1), atomic.LoadInt32(&successes),
		"exactly one concurrent update may succeed in moving the state forward")
	assert.Equal(t, int32(workers-1), atomic.LoadInt32(&rejected),
		"all other updates must be rejected")
}

func assertJobNotFound(t *testing.T, response api.ErrorResponse) {
	t.Helper()
	require.NotNil(t, response.Errors)
	require.Len(t, *response.Errors, 1)
	assert.Equal(t, JobNotFound, (*response.Errors)[0])
}

func persistAPIJob(t *testing.T, db persistence.Storage, clientID string) *api.Job {
	t.Helper()
	wf := dau.DirectWorkflow()
	if _, err := db.GetWorkflow(t.Context(), wf.Name); err != nil {
		_, err = db.CreateWorkflow(t.Context(), wf)
		require.NoError(t, err)
	}
	tags := []string{"tag"}
	job, err := db.CreateJob(t.Context(), &api.Job{
		ClientID: clientID,
		Workflow: wf,
		Status:   &api.JobStatus{State: "INSTALL"},
		Tags:     &tags,
		Definition: map[string]any{
			"client": strings.ToUpper(clientID),
		},
	})
	require.NoError(t, err)
	return job
}

func newSQLiteStorage(t *testing.T) persistence.Storage {
	db := &entgo.SQLite{}
	// File-backed sqlite with WAL journaling and a generous busy timeout so
	// concurrent writers serialize cleanly instead of returning SQLITE_BUSY.
	// A pure in-memory DSN combined with cache=shared still exhibits
	// "database table is locked" under contention because mattn/go-sqlite3
	// opens multiple connections; the WAL journal mode plus busy_timeout
	// avoids that.
	f := filepath.Join(t.TempDir(), "wfx.db")
	dsn := "file:" + f + "?_fk=1&_journal=WAL&_busy_timeout=5000"
	require.NoError(t, db.Initialize(dsn))
	t.Cleanup(db.Shutdown)
	return db
}
