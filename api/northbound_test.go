package api

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/ftag"
	"github.com/siemens/wfx/generated/api"
	"github.com/siemens/wfx/persistence"
	"github.com/siemens/wfx/workflow/dau"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNorthboundDeleteWorkflowsNameHandle_NotFound(t *testing.T) {
	workflow := "wfx.workflow.test"

	dbMock := persistence.NewHealthyMockStorage(t)
	dbMock.EXPECT().DeleteWorkflow(context.Background(), workflow).Return(fault.Wrap(errors.New("workflow not found"), ftag.With(ftag.NotFound)))

	wfx := NewWfxServer(dbMock)
	t.Cleanup(func() { wfx.Stop() })
	server := NewNorthboundServer(wfx)
	resp, err := server.DeleteWorkflowsName(context.Background(), api.DeleteWorkflowsNameRequestObject{Name: workflow})
	assert.NoError(t, err)

	recorder := httptest.NewRecorder()
	_ = resp.VisitDeleteWorkflowsNameResponse(recorder)
	response := recorder.Result()

	assert.Equal(t, http.StatusNotFound, response.StatusCode)
}

func TestNorthboundDeleteWorkflowsNameHandle_InternalError(t *testing.T) {
	workflow := "wfx.workflow.test"

	dbMock := persistence.NewHealthyMockStorage(t)
	dbMock.EXPECT().DeleteWorkflow(context.Background(), workflow).Return(errors.New("something went wrong"))

	wfx := NewWfxServer(dbMock)
	t.Cleanup(func() { wfx.Stop() })
	server := NewNorthboundServer(wfx)
	resp, err := server.DeleteWorkflowsName(context.Background(), api.DeleteWorkflowsNameRequestObject{Name: workflow})
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestNorthboundPostWorkflows_InternalError(t *testing.T) {
	workflow := dau.DirectWorkflow()
	dbMock := persistence.NewHealthyMockStorage(t)
	dbMock.EXPECT().CreateWorkflow(context.Background(), workflow).Return(nil, errors.New("something went wrong"))

	wfx := NewWfxServer(dbMock)
	t.Cleanup(func() { wfx.Stop() })
	server := NewNorthboundServer(wfx)
	resp, err := server.PostWorkflows(context.Background(), api.PostWorkflowsRequestObject{Body: workflow})
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestNorthboundPostWorkflows_AlreadyExists(t *testing.T) {
	workflow := dau.DirectWorkflow()

	dbMock := persistence.NewHealthyMockStorage(t)
	dbMock.EXPECT().CreateWorkflow(context.Background(), workflow).Return(nil, fault.Wrap(errors.New("already exists"), ftag.With(ftag.AlreadyExists)))

	wfx := NewWfxServer(dbMock)
	t.Cleanup(func() { wfx.Stop() })
	server := NewNorthboundServer(wfx)

	resp, err := server.PostWorkflows(context.Background(), api.PostWorkflowsRequestObject{Body: workflow})
	assert.NoError(t, err)

	recorder := httptest.NewRecorder()
	_ = resp.VisitPostWorkflowsResponse(recorder)
	response := recorder.Result()

	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
}

func TestNorthboundPostWorkflows_InvalidWorkflow(t *testing.T) {
	workflow := &api.Workflow{Name: "foo"}

	dbMock := persistence.NewHealthyMockStorage(t)

	wfx := NewWfxServer(dbMock)
	t.Cleanup(func() { wfx.Stop() })
	server := NewNorthboundServer(wfx)

	resp, err := server.PostWorkflows(context.Background(), api.PostWorkflowsRequestObject{Body: workflow})
	assert.NoError(t, err)

	recorder := httptest.NewRecorder()
	_ = resp.VisitPostWorkflowsResponse(recorder)
	response := recorder.Result()

	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
}

func TestNorthboundPostJobs_BadRequest(t *testing.T) {
	wf := dau.DirectWorkflow()
	jobRequest := api.JobRequest{Workflow: wf.Name}

	dbMock := persistence.NewHealthyMockStorage(t)
	dbMock.EXPECT().GetWorkflow(context.Background(), wf.Name).Return(nil, fault.Wrap(errors.New("invalid"), ftag.With(ftag.NotFound)))

	wfx := NewWfxServer(dbMock)
	t.Cleanup(func() { wfx.Stop() })
	server := NewNorthboundServer(wfx)

	resp, err := server.PostJobs(context.Background(), api.PostJobsRequestObject{Body: &jobRequest})
	assert.NoError(t, err)

	recorder := httptest.NewRecorder()
	_ = resp.VisitPostJobsResponse(recorder)
	response := recorder.Result()

	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
}

func TestNorthboundPostJobs_InternalError(t *testing.T) {
	wf := dau.DirectWorkflow()
	jobRequest := api.JobRequest{Workflow: wf.Name}

	dbMock := persistence.NewHealthyMockStorage(t)
	dbMock.EXPECT().GetWorkflow(context.Background(), wf.Name).Return(wf, nil)
	dbMock.EXPECT().CreateJob(context.Background(), mock.Anything).Return(nil, errors.New("something went wrong"))

	wfx := NewWfxServer(dbMock)
	t.Cleanup(func() { wfx.Stop() })
	server := NewNorthboundServer(wfx)

	resp, err := server.PostJobs(context.Background(), api.PostJobsRequestObject{Body: &jobRequest})
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestNorthboundDeleteJobsID_NotFound(t *testing.T) {
	jobID := "42"

	dbMock := persistence.NewHealthyMockStorage(t)
	dbMock.EXPECT().GetJob(context.Background(), jobID, persistence.FetchParams{}).Return(nil, fault.Wrap(errors.New("not found"), ftag.With(ftag.NotFound)))

	wfx := NewWfxServer(dbMock)
	t.Cleanup(func() { wfx.Stop() })
	server := NewNorthboundServer(wfx)

	resp, err := server.DeleteJobsId(context.Background(), api.DeleteJobsIdRequestObject{Id: jobID})
	assert.NoError(t, err)

	recorder := httptest.NewRecorder()
	_ = resp.VisitDeleteJobsIdResponse(recorder)
	response := recorder.Result()

	assert.Equal(t, http.StatusNotFound, response.StatusCode)
}

func TestNorthboundDeleteJobsID_InternalError(t *testing.T) {
	jobID := "42"

	dbMock := persistence.NewHealthyMockStorage(t)
	dbMock.EXPECT().GetJob(context.Background(), jobID, persistence.FetchParams{}).Return(&api.Job{ID: jobID}, nil)
	dbMock.EXPECT().DeleteJob(context.Background(), jobID).Return(fault.Wrap(errors.New("something went wrong"), ftag.With(ftag.Internal)))

	wfx := NewWfxServer(dbMock)
	t.Cleanup(func() { wfx.Stop() })
	server := NewNorthboundServer(wfx)

	resp, err := server.DeleteJobsId(context.Background(), api.DeleteJobsIdRequestObject{Id: jobID})
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestNorthboundPostJobsIDTags_NotFound(t *testing.T) {
	jobID := "42"

	dbMock := persistence.NewHealthyMockStorage(t)
	dbMock.EXPECT().GetJob(context.Background(), jobID, persistence.FetchParams{}).Return(nil, fault.Wrap(errors.New("not found"), ftag.With(ftag.NotFound)))

	wfx := NewWfxServer(dbMock)
	t.Cleanup(func() { wfx.Stop() })
	server := NewNorthboundServer(wfx)

	resp, err := server.PostJobsIdTags(context.Background(), api.PostJobsIdTagsRequestObject{Id: jobID})
	assert.NoError(t, err)

	recorder := httptest.NewRecorder()
	_ = resp.VisitPostJobsIdTagsResponse(recorder)
	response := recorder.Result()

	assert.Equal(t, http.StatusNotFound, response.StatusCode)
}

func TestNorthboundPostJobsIDTags_InternalError(t *testing.T) {
	jobID := "42"

	dbMock := persistence.NewHealthyMockStorage(t)
	dbMock.EXPECT().GetJob(context.Background(), jobID, persistence.FetchParams{}).Return(nil, fault.Wrap(errors.New("something went wrong"), ftag.With(ftag.Internal)))

	wfx := NewWfxServer(dbMock)
	t.Cleanup(func() { wfx.Stop() })
	server := NewNorthboundServer(wfx)

	resp, err := server.PostJobsIdTags(context.Background(), api.PostJobsIdTagsRequestObject{Id: jobID})
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestNorthboundDeleteJobsIDTags_NotFound(t *testing.T) {
	jobID := "42"

	dbMock := persistence.NewHealthyMockStorage(t)
	dbMock.EXPECT().GetJob(context.Background(), jobID, persistence.FetchParams{}).Return(nil, fault.Wrap(errors.New("not found"), ftag.With(ftag.NotFound)))

	wfx := NewWfxServer(dbMock)
	t.Cleanup(func() { wfx.Stop() })
	server := NewNorthboundServer(wfx)

	resp, err := server.DeleteJobsIdTags(context.Background(), api.DeleteJobsIdTagsRequestObject{Id: jobID})
	assert.NoError(t, err)

	recorder := httptest.NewRecorder()
	_ = resp.VisitDeleteJobsIdTagsResponse(recorder)
	response := recorder.Result()

	assert.Equal(t, http.StatusNotFound, response.StatusCode)
}

func TestNorthboundDeleteJobsIDTags_InternalError(t *testing.T) {
	jobID := "42"

	dbMock := persistence.NewHealthyMockStorage(t)
	dbMock.EXPECT().GetJob(context.Background(), jobID, persistence.FetchParams{}).Return(nil, fault.Wrap(errors.New("not found"), ftag.With(ftag.Internal)))

	wfx := NewWfxServer(dbMock)
	t.Cleanup(func() { wfx.Stop() })
	server := NewNorthboundServer(wfx)

	resp, err := server.DeleteJobsIdTags(context.Background(), api.DeleteJobsIdTagsRequestObject{Id: jobID})
	assert.Error(t, err)
	assert.Nil(t, resp)
}
