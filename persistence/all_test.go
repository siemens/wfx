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
	"testing"

	"github.com/siemens/wfx/generated/model"
	"github.com/stretchr/testify/assert"
)

type testStorage struct{}

func (teststorage *testStorage) Initialize(context.Context, string) error {
	panic("not implemented")
}

func (teststorage *testStorage) Shutdown() {
	panic("not implemented")
}

func (teststorage *testStorage) CheckHealth(context.Context) error {
	panic("not implemented")
}

func (teststorage *testStorage) CreateJob(context.Context, *model.Job) (*model.Job, error) {
	panic("not implemented")
}

func (teststorage *testStorage) GetJob(context.Context, string, FetchParams) (*model.Job, error) {
	panic("not implemented")
}

func (teststorage *testStorage) UpdateJob(context.Context, *model.Job, JobUpdate) (*model.Job, error) {
	panic("not implemented")
}

func (teststorage *testStorage) DeleteJob(context.Context, string) error {
	panic("not implemented")
}

func (teststorage *testStorage) QueryJobs(context.Context, FilterParams, SortParams, PaginationParams) (*model.PaginatedJobList, error) {
	panic("not implemented")
}

func (teststorage *testStorage) CreateWorkflow(context.Context, *model.Workflow) (*model.Workflow, error) {
	panic("not implemented")
}

func (teststorage *testStorage) GetWorkflow(context.Context, string) (*model.Workflow, error) {
	panic("not implemented")
}

func (teststorage *testStorage) DeleteWorkflow(context.Context, string) error {
	panic("not implemented")
}

func (teststorage *testStorage) QueryWorkflows(context.Context, PaginationParams) (*model.PaginatedWorkflowList, error) {
	panic("not implemented")
}

func TestStorageAPI(t *testing.T) {
	storage1 := &testStorage{}
	RegisterStorage("storage1", storage1)
	actual := GetStorage("storage1")
	assert.Same(t, storage1, actual)
	storage2 := &testStorage{}
	RegisterStorage("storage2", storage2)
	all := Storages()
	assert.Len(t, all, 2)
	assert.Nil(t, GetStorage("foo"))
}
