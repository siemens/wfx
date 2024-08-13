package persistence

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"context"

	"github.com/siemens/wfx/generated/api"
)

// Storage represents an interface for a persistence layer (such as PostgreSQL, SQLite).
// It provides methods for managing jobs and workflows in the storage.
type Storage interface {
	// Initialize sets up the storage using the provided options string.
	Initialize(options string) error

	// Shutdown gracefully closes the storage connection.
	Shutdown()

	// CheckHealth examines the status of the storage, returning an error if the storage is not available.
	CheckHealth(ctx context.Context) error

	// CreateJob adds a new job to the storage.
	CreateJob(ctx context.Context, job *api.Job) (*api.Job, error)

	// GetJob retrieves an existing job identified by jobID from the storage.
	// If an issue occurs during the fetch operation, the method returns an error.
	GetJob(ctx context.Context, jobID string, fetchParams FetchParams) (*api.Job, error)

	// UpdateJob modifies an existing job in the storage based on the provided JobUpdate request.
	UpdateJob(ctx context.Context, job *api.Job, request JobUpdate) (*api.Job, error)

	// DeleteJob removes an existing job identified by jobID from the storage.
	DeleteJob(ctx context.Context, jobID string) error

	// QueryJobs retrieves jobs that satisfy the filterParams, sortParams, and paginationParams.
	QueryJobs(ctx context.Context, filterParams FilterParams, sortParams SortParams, paginationParams PaginationParams) (*api.PaginatedJobList, error)

	// CreateWorkflow adds a new workflow to the storage.
	CreateWorkflow(ctx context.Context, workflow *api.Workflow) (*api.Workflow, error)

	// GetWorkflow retrieves an existing workflow identified by name from the storage.
	// If an issue occurs during the fetch operation, the method returns an error.
	GetWorkflow(ctx context.Context, name string) (*api.Workflow, error)

	// DeleteWorkflow removes an existing workflow identified by name from the storage.
	DeleteWorkflow(ctx context.Context, name string) error

	// QueryWorkflows retrieves all workflows from the storage respecting the paginationParams.
	QueryWorkflows(ctx context.Context, sortParams SortParams, paginationParams PaginationParams) (*api.PaginatedWorkflowList, error)
}

// JobUpdate encapsulates the properties of a job that can be updated.
// If a property is nil, its corresponding value in the job will not be changed.
type JobUpdate struct {
	// Status is the new job status. If provided, it replaces the existing status of the job.
	Status *api.JobStatus
	// Definition is the new job definition. If provided, it replaces the existing job definition.
	Definition *map[string]any
	// AddTags is a list of tags to be added to the job. If provided, these tags will be added to the existing tags.
	AddTags *[]string
	// DelTags is a list of tags to be removed from the job. If provided, these tags will be
	// removed from the job's existing tags. If a tag specified here does not exist in the
	// job's tags, it will be ignored.
	DelTags *[]string
}

// PaginationParams controls the pagination of response lists.
// It allows the client to specify a subset of results to return, which can be useful for large data sets.
type PaginationParams struct {
	// Offset is the number of items to skip before starting to collect results.
	Offset int64
	// Limit is the maximum number of items to return in the response.
	Limit int32
}

// FetchParams control the level of detail returned by fetch operations.
// By default, all boolean parameters are false in Go.
// If you add a new parameter, ensure that `false` represents the sensible default value.
type FetchParams struct {
	// History, when set to true, includes the transition history of the job in the fetched data.
	History bool
}

// FilterParams define criteria for filtering jobs.
// Each field represents a different filter that can be applied.
// A job entity has to match all criteria in order to be returned.
type FilterParams struct {
	// ClientID allows filtering jobs that belong to a specific client.
	// Only jobs with a matching client ID will be returned.
	ClientID *string
	// Group allows filtering jobs that belong to one of the specified groups.
	// The filter is an OR filter, meaning jobs that belong to any of the provided groups will be returned.
	Group []string
	// State allows filtering jobs based on their current state value.
	// Only jobs with a matching state will be returned.
	State *string
	// Workflow allows filtering jobs that are created from a specific workflow.
	// Only jobs with a matching workflow name will be returned.
	Workflow *string
	// Tags allows filtering jobs that contain one or more of the specified tags.
	// The filter is an OR filter, meaning jobs that contain any of the provided tags will be returned.
	Tags []string
}

// SortParams specify the order of the returned jobs.
type SortParams struct {
	// Desc, when set to true, sorts the jobs in descending order based on their creation time.
	// When set to false, the jobs are sorted in ascending order.
	Desc bool
}
