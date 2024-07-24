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
	"testing"

	"github.com/alexliesenfeld/health"
	"github.com/siemens/wfx/generated/api"
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
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	response, err := wfx.GetJobsEvents(ctx, request)
	require.NoError(t, err)
	assert.NotNil(t, response)
}
