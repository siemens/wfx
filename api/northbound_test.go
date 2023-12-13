package api

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/ftag"
	"github.com/siemens/wfx/generated/model"
	"github.com/siemens/wfx/generated/northbound/restapi/operations/northbound"
	"github.com/siemens/wfx/internal/producer"
	"github.com/siemens/wfx/persistence"
	"github.com/siemens/wfx/workflow/dau"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNorthboundGetJobsIDStatusHandler_NotFound(t *testing.T) {
	params := northbound.NewGetJobsIDStatusParams()
	jobID := "42"
	params.ID = jobID
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	dbMock := persistence.NewMockStorage(t)
	dbMock.EXPECT().GetJob(params.HTTPRequest.Context(), params.ID, persistence.FetchParams{}).Return(nil, fault.Wrap(fmt.Errorf("job with id %s does not exist", jobID), ftag.With(ftag.NotFound)))

	api := NewNorthboundAPI(dbMock)
	resp := api.NorthboundGetJobsIDStatusHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusNotFound, response.StatusCode)
}

func TestNorthboundPutJobsIDStatusHandler_NotFound(t *testing.T) {
	params := northbound.NewPutJobsIDStatusParams()
	jobID := "42"
	params.ID = jobID
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	dbMock := persistence.NewMockStorage(t)
	dbMock.EXPECT().GetJob(params.HTTPRequest.Context(), params.ID, persistence.FetchParams{}).Return(nil, fault.Wrap(fmt.Errorf("job with id %s does not exist", jobID), ftag.With(ftag.NotFound)))

	api := NewNorthboundAPI(dbMock)
	resp := api.NorthboundPutJobsIDStatusHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusNotFound, response.StatusCode)
}

func TestNorthboundGetJobsHandler_InternalError(t *testing.T) {
	params := northbound.NewGetJobsParams()
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	dbMock := persistence.NewMockStorage(t)
	dbMock.EXPECT().
		QueryJobs(params.HTTPRequest.Context(), persistence.FilterParams{}, persistence.SortParams{}, persistence.PaginationParams{Limit: 10}).
		Return(nil, errors.New("something went wrong"))

	api := NewNorthboundAPI(dbMock)
	resp := api.NorthboundGetJobsHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
}

func TestNorthboundGetJobsIDHandler_NotFound(t *testing.T) {
	params := northbound.NewGetJobsIDParams()

	jobID := "42"
	params.ID = jobID
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	dbMock := persistence.NewMockStorage(t)
	dbMock.EXPECT().GetJob(params.HTTPRequest.Context(), params.ID, persistence.FetchParams{}).Return(nil, fault.Wrap(fmt.Errorf("job with id %s does not exist", jobID), ftag.With(ftag.NotFound)))

	api := NewNorthboundAPI(dbMock)
	resp := api.NorthboundGetJobsIDHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusNotFound, response.StatusCode)
}

func TestNorthboundGetJobsIDHandler_InternalError(t *testing.T) {
	params := northbound.NewGetJobsIDParams()
	jobID := "42"
	params.ID = jobID
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	history := true
	params.History = &history
	dbMock := persistence.NewMockStorage(t)
	dbMock.EXPECT().GetJob(params.HTTPRequest.Context(), params.ID, persistence.FetchParams{History: history}).Return(nil, errors.New("something went wrong"))

	api := NewNorthboundAPI(dbMock)
	resp := api.NorthboundGetJobsIDHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
}

func TestNorthboundGetJobsIDStatusHandler_InternalError(t *testing.T) {
	params := northbound.NewGetJobsIDStatusParams()
	jobID := "42"
	params.ID = jobID
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	dbMock := persistence.NewMockStorage(t)
	dbMock.EXPECT().GetJob(params.HTTPRequest.Context(), params.ID, persistence.FetchParams{}).Return(nil, errors.New("something went wrong"))

	api := NewNorthboundAPI(dbMock)
	resp := api.NorthboundGetJobsIDStatusHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
}

func TestNorthboundPutJobsIDStatusHandler_InternalError(t *testing.T) {
	params := northbound.NewPutJobsIDStatusParams()
	jobID := "42"
	params.ID = jobID
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	dbMock := persistence.NewMockStorage(t)
	dbMock.EXPECT().GetJob(params.HTTPRequest.Context(), params.ID, persistence.FetchParams{}).Return(nil, errors.New("something went wrong"))

	api := NewNorthboundAPI(dbMock)
	resp := api.NorthboundPutJobsIDStatusHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
}

func TestNorthboundGetJobsIDDefinitionHandler_NotFound(t *testing.T) {
	params := northbound.NewGetJobsIDDefinitionParams()
	jobID := "42"
	params.ID = jobID
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	dbMock := persistence.NewMockStorage(t)
	dbMock.EXPECT().GetJob(params.HTTPRequest.Context(), params.ID, persistence.FetchParams{}).Return(nil, fault.Wrap(fmt.Errorf("job with id %s does not exist", jobID), ftag.With(ftag.NotFound)))

	api := NewNorthboundAPI(dbMock)
	resp := api.NorthboundGetJobsIDDefinitionHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusNotFound, response.StatusCode)
}

func TestNorthboundGetJobsIDDefinitionHandler_InternalError(t *testing.T) {
	params := northbound.NewGetJobsIDDefinitionParams()
	jobID := "42"
	params.ID = jobID
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	dbMock := persistence.NewMockStorage(t)
	dbMock.EXPECT().GetJob(params.HTTPRequest.Context(), params.ID, persistence.FetchParams{}).Return(nil, errors.New("something went wrong"))

	api := NewNorthboundAPI(dbMock)
	resp := api.NorthboundGetJobsIDDefinitionHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
}

func TestNorthboundPutJobsIDDefinitionHandler_NotFound(t *testing.T) {
	params := northbound.NewPutJobsIDDefinitionParams()
	jobID := "42"
	params.ID = jobID
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	dbMock := persistence.NewMockStorage(t)
	dbMock.EXPECT().GetJob(params.HTTPRequest.Context(), params.ID, persistence.FetchParams{}).Return(nil, fault.Wrap(fmt.Errorf("job with id %s does not exist", jobID), ftag.With(ftag.NotFound)))

	api := NewNorthboundAPI(dbMock)
	resp := api.NorthboundPutJobsIDDefinitionHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusNotFound, response.StatusCode)
}

func TestNorthboundPutJobsIDDefinitionHandler_InternalError(t *testing.T) {
	params := northbound.NewPutJobsIDDefinitionParams()
	jobID := "42"
	params.ID = jobID
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	dbMock := persistence.NewMockStorage(t)
	dbMock.EXPECT().GetJob(params.HTTPRequest.Context(), params.ID, persistence.FetchParams{}).Return(nil, errors.New("something went wrong"))

	api := NewNorthboundAPI(dbMock)
	resp := api.NorthboundPutJobsIDDefinitionHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
}

func TestNorthboundGetWorkflowsNameHandler_InternalError(t *testing.T) {
	params := northbound.NewGetWorkflowsNameParams()
	params.Name = "wfx.test.workflow"
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	dbMock := persistence.NewMockStorage(t)
	dbMock.EXPECT().GetWorkflow(params.HTTPRequest.Context(), params.Name).Return(nil, errors.New("something went wrong"))

	api := NewNorthboundAPI(dbMock)

	resp := api.NorthboundGetWorkflowsNameHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
}

func TestNorthboundGetWorkflowsHandler_InternalError(t *testing.T) {
	params := northbound.NewGetWorkflowsParams()
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	dbMock := persistence.NewMockStorage(t)
	dbMock.EXPECT().QueryWorkflows(params.HTTPRequest.Context(), persistence.PaginationParams{Limit: 10}).Return(nil, errors.New("something went wrong"))

	api := NewNorthboundAPI(dbMock)
	resp := api.NorthboundGetWorkflowsHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
}

func TestNorthboundGetWorkflowsHandler_Empty(t *testing.T) {
	params := northbound.NewGetWorkflowsParams()
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	dbMock := persistence.NewMockStorage(t)
	dbMock.EXPECT().QueryWorkflows(params.HTTPRequest.Context(), persistence.PaginationParams{Limit: 10}).Return(nil, nil)

	api := NewNorthboundAPI(dbMock)

	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	resp := api.NorthboundGetWorkflowsHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusOK, response.StatusCode)
}

func TestNorthboundDeleteWorkflowsNameHandle_NotFound(t *testing.T) {
	params := northbound.NewDeleteWorkflowsNameParams()
	params.Name = "wfx.workflow.test"
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	dbMock := persistence.NewMockStorage(t)
	dbMock.EXPECT().DeleteWorkflow(params.HTTPRequest.Context(), params.Name).Return(fault.Wrap(errors.New("workflow not found"), ftag.With(ftag.NotFound)))

	api := NewNorthboundAPI(dbMock)

	resp := api.NorthboundDeleteWorkflowsNameHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusNotFound, response.StatusCode)
}

func TestNorthboundDeleteWorkflowsNameHandle_InternalError(t *testing.T) {
	params := northbound.NewDeleteWorkflowsNameParams()
	params.Name = "wfx.workflow.test"
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	dbMock := persistence.NewMockStorage(t)
	dbMock.EXPECT().DeleteWorkflow(params.HTTPRequest.Context(), params.Name).Return(errors.New("something went wrong"))

	api := NewNorthboundAPI(dbMock)
	resp := api.NorthboundDeleteWorkflowsNameHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
}

func TestNorthboundPostWorkflowsHandler_InternalError(t *testing.T) {
	params := northbound.NewPostWorkflowsParams()
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	params.Workflow = dau.DirectWorkflow()
	dbMock := persistence.NewMockStorage(t)
	dbMock.EXPECT().CreateWorkflow(params.HTTPRequest.Context(), params.Workflow).Return(nil, errors.New("something went wrong"))

	api := NewNorthboundAPI(dbMock)
	resp := api.NorthboundPostWorkflowsHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
}

func TestNorthboundPostWorkflowsHandler_AlreadyExists(t *testing.T) {
	params := northbound.NewPostWorkflowsParams()
	params.Workflow = dau.DirectWorkflow()
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	dbMock := persistence.NewMockStorage(t)
	dbMock.EXPECT().CreateWorkflow(params.HTTPRequest.Context(), params.Workflow).Return(nil, fault.Wrap(errors.New("already exists"), ftag.With(ftag.AlreadyExists)))

	api := NewNorthboundAPI(dbMock)
	resp := api.NorthboundPostWorkflowsHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
}

func TestNorthboundPostWorkflowsHandler_InvalidWorkflow(t *testing.T) {
	params := northbound.NewPostWorkflowsParams()
	params.Workflow = &model.Workflow{Name: "foo"}
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	dbMock := persistence.NewMockStorage(t)

	api := NewNorthboundAPI(dbMock)
	resp := api.NorthboundPostWorkflowsHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
}

func TestNorthboundPostJobsHandler_BadRequest(t *testing.T) {
	params := northbound.NewPostJobsParams()
	wf := dau.DirectWorkflow()
	params.Job = &model.JobRequest{Workflow: wf.Name}
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	dbMock := persistence.NewMockStorage(t)
	dbMock.EXPECT().GetWorkflow(params.HTTPRequest.Context(), params.Job.Workflow).Return(nil, fault.Wrap(errors.New("invalid"), ftag.With(ftag.NotFound)))

	api := NewNorthboundAPI(dbMock)
	resp := api.NorthboundPostJobsHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
}

func TestNorthboundPostJobsHandler_InternalError(t *testing.T) {
	wf := dau.DirectWorkflow()
	params := northbound.NewPostJobsParams()
	params.Job = &model.JobRequest{Workflow: wf.Name}
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))

	dbMock := persistence.NewMockStorage(t)
	dbMock.EXPECT().GetWorkflow(params.HTTPRequest.Context(), params.Job.Workflow).Return(wf, nil)
	dbMock.EXPECT().CreateJob(params.HTTPRequest.Context(), mock.Anything).Return(nil, errors.New("something went wrong"))

	api := NewNorthboundAPI(dbMock)
	resp := api.NorthboundPostJobsHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
}

func TestNorthboundDeleteJobsIDHandler_NotFound(t *testing.T) {
	params := northbound.NewDeleteJobsIDParams()
	params.ID = "42"
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))

	dbMock := persistence.NewMockStorage(t)
	dbMock.EXPECT().GetJob(params.HTTPRequest.Context(), params.ID, persistence.FetchParams{}).Return(&model.Job{}, nil)
	dbMock.EXPECT().DeleteJob(params.HTTPRequest.Context(), params.ID).Return(fault.Wrap(errors.New("not found"), ftag.With(ftag.NotFound)))

	api := NewNorthboundAPI(dbMock)
	resp := api.NorthboundDeleteJobsIDHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusNotFound, response.StatusCode)
}

func TestNorthboundDeleteJobsIDHandler_InternalError(t *testing.T) {
	params := northbound.NewDeleteJobsIDParams()
	params.ID = "42"
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))

	dbMock := persistence.NewMockStorage(t)
	dbMock.EXPECT().GetJob(params.HTTPRequest.Context(), params.ID, persistence.FetchParams{}).Return(&model.Job{}, nil)
	dbMock.EXPECT().DeleteJob(params.HTTPRequest.Context(), params.ID).Return(fault.Wrap(errors.New("something went wrong"), ftag.With(ftag.Internal)))

	api := NewNorthboundAPI(dbMock)
	resp := api.NorthboundDeleteJobsIDHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
}

func TestNorthboundGetJobsIDTagsHandler_NotFound(t *testing.T) {
	params := northbound.NewGetJobsIDTagsParams()
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))

	dbMock := persistence.NewMockStorage(t)
	dbMock.EXPECT().GetJob(params.HTTPRequest.Context(), params.ID, persistence.FetchParams{}).Return(nil, fault.Wrap(errors.New("not found"), ftag.With(ftag.NotFound)))

	api := NewNorthboundAPI(dbMock)
	resp := api.NorthboundGetJobsIDTagsHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusNotFound, response.StatusCode)
}

func TestNorthboundGetJobsIDTagsHandler_InternalError(t *testing.T) {
	params := northbound.NewGetJobsIDTagsParams()
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))

	dbMock := persistence.NewMockStorage(t)
	dbMock.EXPECT().GetJob(params.HTTPRequest.Context(), params.ID, persistence.FetchParams{}).Return(nil, fault.Wrap(errors.New("something went wrong"), ftag.With(ftag.Internal)))

	api := NewNorthboundAPI(dbMock)
	resp := api.NorthboundGetJobsIDTagsHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
}

func TestNorthboundPostJobsIDTagsHandler_NotFound(t *testing.T) {
	params := northbound.NewPostJobsIDTagsParams()
	params.ID = "42"
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))

	dbMock := persistence.NewMockStorage(t)
	dbMock.EXPECT().GetJob(params.HTTPRequest.Context(), params.ID, persistence.FetchParams{}).Return(nil, fault.Wrap(errors.New("not found"), ftag.With(ftag.NotFound)))

	api := NewNorthboundAPI(dbMock)
	resp := api.NorthboundPostJobsIDTagsHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusNotFound, response.StatusCode)
}

func TestNorthboundPostJobsIDTagsHandler_InternalError(t *testing.T) {
	params := northbound.NewPostJobsIDTagsParams()
	params.ID = "42"
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))

	dbMock := persistence.NewMockStorage(t)
	dbMock.EXPECT().GetJob(params.HTTPRequest.Context(), params.ID, persistence.FetchParams{}).Return(nil, fault.Wrap(errors.New("something went wrong"), ftag.With(ftag.Internal)))

	api := NewNorthboundAPI(dbMock)
	resp := api.NorthboundPostJobsIDTagsHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
}

func TestNorthboundDeleteJobsIDTagsHandler_NotFound(t *testing.T) {
	params := northbound.NewDeleteJobsIDTagsParams()
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))

	dbMock := persistence.NewMockStorage(t)
	dbMock.EXPECT().GetJob(params.HTTPRequest.Context(), params.ID, persistence.FetchParams{}).Return(nil, fault.Wrap(errors.New("not found"), ftag.With(ftag.NotFound)))

	api := NewNorthboundAPI(dbMock)
	resp := api.NorthboundDeleteJobsIDTagsHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusNotFound, response.StatusCode)
}

func TestNorthboundDeleteJobsIDTagsHandler_InternalError(t *testing.T) {
	params := northbound.NewDeleteJobsIDTagsParams()
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))

	dbMock := persistence.NewMockStorage(t)
	dbMock.EXPECT().GetJob(params.HTTPRequest.Context(), params.ID, persistence.FetchParams{}).Return(nil, fault.Wrap(errors.New("not found"), ftag.With(ftag.Internal)))

	api := NewNorthboundAPI(dbMock)
	resp := api.NorthboundDeleteJobsIDTagsHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
}

func TestParseFilterParamsNorth(t *testing.T) {
	jobIDs := "abc,424-194-123"
	clientIDs := "alpha,beta"
	workflows := "wf1,wf2,wf3"
	params := northbound.GetJobsEventsParams{
		HTTPRequest: &http.Request{},
		JobIds:      &jobIDs,
		ClientIds:   &clientIDs,
		Workflows:   &workflows,
	}
	filter := parseFilterParamsNorth(params)
	assert.Equal(t, []string{"abc", "424-194-123"}, filter.JobIDs)
	assert.Equal(t, []string{"alpha", "beta"}, filter.ClientIDs)
	assert.Equal(t, []string{"wf1", "wf2", "wf3"}, filter.Workflows)
}
