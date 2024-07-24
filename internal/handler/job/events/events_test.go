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

	"github.com/olebedev/emitter"
	"github.com/siemens/wfx/generated/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddSubscriberAndShutdown(t *testing.T) {
	for i := 1; i <= 2; i++ {
		_, err := AddSubscriber(context.Background(), FilterParams{JobIDs: []string{"42"}}, nil)
		require.NoError(t, err)
		assert.Equal(t, i, SubscriberCount())
	}
	ShutdownSubscribers()
	assert.Equal(t, 0, SubscriberCount())
}

func TestFiltering(t *testing.T) {
	job1 := api.Job{ID: "1", Workflow: &api.Workflow{Name: "workflow"}, Status: &api.JobStatus{State: "INITIAL"}}
	job2 := api.Job{ID: "2", Workflow: &api.Workflow{Name: "workflow"}, Status: &api.JobStatus{State: "INITIAL"}}

	ctx := context.Background()
	ch1, _ := AddSubscriber(ctx, FilterParams{JobIDs: []string{job1.ID}}, nil)
	ch2, _ := AddSubscriber(ctx, FilterParams{JobIDs: []string{job2.ID}}, nil)
	chCombined, _ := AddSubscriber(ctx, FilterParams{JobIDs: []string{job1.ID, job2.ID}}, nil)
	chAll, _ := AddSubscriber(ctx, FilterParams{}, nil) // no filter should receive all events

	job1.Status.State = "FOO"
	<-PublishEvent(context.Background(), &JobEvent{Action: ActionUpdateStatus, Job: &job1})

	job2.Status.State = "BAR"
	<-PublishEvent(context.Background(), &JobEvent{Action: ActionUpdateStatus, Job: &job2})

	{
		ev := <-ch1
		actual := ev.Args[0].(*JobEvent)
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
		ev := <-ch2
		actual := ev.Args[0].(*JobEvent)
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
		for _, ch := range []<-chan emitter.Event{chCombined, chAll} {
			ev := <-ch
			actual := ev.Args[0].(*JobEvent)
			assert.Equal(t, "FOO", actual.Job.Status.State)

			ev = <-ch
			actual = ev.Args[0].(*JobEvent)
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
