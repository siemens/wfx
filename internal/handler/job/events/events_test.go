package events

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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddSubscriberAndShutdown(t *testing.T) {
	for i := 1; i <= 2; i++ {
		_ = AddSubscriber(context.Background(), FilterParams{JobIDs: []string{"42"}}, nil)
		assert.Equal(t, i, SubscriberCount())
	}
	ShutdownSubscribers()
	assert.Equal(t, 0, SubscriberCount())
}

func TestFiltering(t *testing.T) {
	job1 := api.Job{ID: "1", Workflow: &api.Workflow{Name: "workflow"}, Status: &api.JobStatus{State: "INITIAL"}}
	job2 := api.Job{ID: "2", Workflow: &api.Workflow{Name: "workflow"}, Status: &api.JobStatus{State: "INITIAL"}}

	ctx := context.Background()
	ch1 := AddSubscriber(ctx, FilterParams{JobIDs: []string{job1.ID}}, nil)
	ch2 := AddSubscriber(ctx, FilterParams{JobIDs: []string{job2.ID}}, nil)
	chCombined := AddSubscriber(ctx, FilterParams{JobIDs: []string{job1.ID, job2.ID}}, nil)
	chAll := AddSubscriber(ctx, FilterParams{}, nil) // no filter should receive all events

	job1.Status.State = "FOO"
	PublishEvent(t.Context(), JobEvent{Action: ActionUpdateStatus, Job: &job1})

	job2.Status.State = "BAR"
	PublishEvent(t.Context(), JobEvent{Action: ActionUpdateStatus, Job: &job2})

	{
		actual, ok := <-ch1
		require.True(t, ok)
		assert.Equal(t, ActionUpdateStatus, actual.Action)
		assert.Equal(t, "FOO", actual.Job.Status.State)

		// check there is nothing else
		select {
		case <-ch1:
			assert.Fail(t, "Received unexpected event")
		default:
			// nothing there, good
		}
	}
	{
		actual, ok := <-ch2
		require.True(t, ok)
		assert.Equal(t, "BAR", actual.Job.Status.State)

		// check there is nothing else
		select {
		case <-ch2:
			assert.Fail(t, "Received unexpected event")
		default:
			// nothing there, good
		}
	}
	{
		for _, ch := range []<-chan JobEvent{chCombined, chAll} {
			actual, ok := <-ch
			require.True(t, ok)
			assert.Equal(t, "FOO", actual.Job.Status.State)

			actual, ok = <-ch
			require.True(t, ok)
			assert.Equal(t, "BAR", actual.Job.Status.State)

			// check there is nothing else
			select {
			case <-ch:
				assert.Fail(t, "Received unexpected event")
			default:
				// nothing there, good
			}
		}
	}

	ShutdownSubscribers()
}
