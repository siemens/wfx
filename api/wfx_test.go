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
	"sync"
	"sync/atomic"
	"testing"

	"github.com/alexliesenfeld/health"
	"github.com/siemens/wfx/generated/api"
	"github.com/siemens/wfx/internal/persistence/entgo"
	"github.com/siemens/wfx/persistence"
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

func TestGetJobsEvents(t *testing.T) {
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
