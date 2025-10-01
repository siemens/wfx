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
	"sync"
	"testing"
	"time"

	"github.com/siemens/wfx/generated/api"
	"github.com/stretchr/testify/assert"
)

func TestAddSubscriberAndShutdown(t *testing.T) {
	for i := 1; i <= 2; i++ {
		_ = AddSubscriber(t.Context(), time.Minute, FilterParams{JobIDs: []string{"42"}}, nil)
		assert.Equal(t, i, SubscriberCount())
	}
	ShutdownSubscribers()
	assert.Equal(t, 0, SubscriberCount())
}

func TestFiltering(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)

	job1 := api.Job{ID: "1", Workflow: &api.Workflow{Name: "workflow"}, Status: &api.JobStatus{State: "FOO"}}
	job2 := api.Job{ID: "2", Workflow: &api.Workflow{Name: "workflow"}, Status: &api.JobStatus{State: "BAR"}}

	sub1 := AddSubscriber(ctx, time.Minute, FilterParams{JobIDs: []string{job1.ID}}, nil)
	sub2 := AddSubscriber(ctx, time.Minute, FilterParams{JobIDs: []string{job2.ID}}, nil)
	sub3 := AddSubscriber(ctx, time.Minute, FilterParams{JobIDs: []string{job1.ID}, Actions: []Action{ActionUpdateStatus}}, nil)
	sub1and2 := AddSubscriber(ctx, time.Minute, FilterParams{JobIDs: []string{job1.ID, job2.ID}}, nil)
	subAll := AddSubscriber(ctx, time.Minute, FilterParams{}, nil) // no filter should receive all events

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		actual := receiveEventBlocking(sub1)
		assert.Equal(t, ActionUpdateStatus, actual.Action)
		assert.Equal(t, "FOO", actual.Job.Status.State)

		// check there is nothing else
		select {
		case <-sub1.Events:
			assert.Fail(t, "Received unexpected event")
		default:
			// nothing there, good
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		actual := receiveEventBlocking(sub2)
		assert.Equal(t, "BAR", actual.Job.Status.State)

		// check there is nothing else
		select {
		case <-sub2.Events:
			assert.Fail(t, "Received unexpected event")
		default:
			// nothing there, good
		}
	}()

	arr := []*Subscriber{sub1and2, subAll}
	for i := range len(arr) {
		sub := arr[i]

		wg.Add(1)
		go func() {
			defer wg.Done()

			actual := receiveEventBlocking(sub)
			assert.Equal(t, "FOO", actual.Job.Status.State)

			actual = receiveEventBlocking(sub)
			assert.Equal(t, "BAR", actual.Job.Status.State)

			// check there is nothing else
			select {
			case <-sub.Events:
				assert.Fail(t, "Received unexpected event")
			default:
				// nothing there, good
			}
			assert.Empty(t, sub.Backlog.data)
		}()

	}

	wg.Add(1)
	go func() {
		defer wg.Done()

		actual := receiveEventBlocking(sub3)
		assert.Equal(t, ActionUpdateStatus, actual.Action)
		assert.Equal(t, "FOO", actual.Job.Status.State)

		// check there is nothing else
		select {
		case <-sub3.Events:
			assert.Fail(t, "Received unexpected event")
		default:
			// nothing there, good
		}
	}()

	PublishEvent(ctx, JobEvent{Action: ActionUpdateStatus, Job: &job1})
	PublishEvent(ctx, JobEvent{Action: ActionUpdateStatus, Job: &job2})

	wg.Wait()

	cancel()
	ShutdownSubscribers()
}

func TestBacklog(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)
	t.Cleanup(ShutdownSubscribers)

	job1 := api.Job{ID: "1", Workflow: &api.Workflow{Name: "workflow"}, Status: &api.JobStatus{State: "ALPHA"}}
	job2 := api.Job{ID: "2", Workflow: &api.Workflow{Name: "workflow"}, Status: &api.JobStatus{State: "BETA"}}
	sub := AddSubscriber(ctx, time.Minute, FilterParams{}, nil)

	PublishEvent(ctx, JobEvent{Action: ActionUpdateStatus, Job: &job1})
	PublishEvent(ctx, JobEvent{Action: ActionUpdateStatus, Job: &job2})

	ev := <-sub.Events
	assert.Equal(t, "ALPHA", ev.Job.Status.State)

	assert.Len(t, sub.Backlog.data, 1)
	assert.Equal(t, "BETA", sub.Backlog.data[0].Job.Status.State)
}

func TestGracePeriod(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)
	t.Cleanup(ShutdownSubscribers)

	job := api.Job{ID: "1", Workflow: &api.Workflow{Name: "workflow"}, Status: &api.JobStatus{State: "ALPHA"}}
	sub := AddSubscriber(ctx, time.Microsecond, FilterParams{}, nil)

	PublishEvent(ctx, JobEvent{Action: ActionUpdateStatus, Job: &job})
	PublishEvent(ctx, JobEvent{Action: ActionUpdateStatus, Job: &job})
	time.Sleep(time.Microsecond)
	PublishEvent(ctx, JobEvent{Action: ActionUpdateStatus, Job: &job})

	ev := <-sub.Events
	assert.Equal(t, job.ID, ev.Job.ID)
	assert.Len(t, sub.Backlog.data, 2)

	_, ok := <-sub.Events
	assert.False(t, ok, "channel should be closed")
}

func receiveEventBlocking(sub *Subscriber) *JobEvent {
	for {
		select {
		case ev, ok := <-sub.Events:
			if !ok {
				return nil
			}
			return &ev
		default:
			// check backlog
			ev, ok := sub.Backlog.Deq()
			if ok {
				return ev
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
}
