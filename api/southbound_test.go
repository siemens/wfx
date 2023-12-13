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
	"github.com/siemens/wfx/generated/southbound/restapi/operations/southbound"
	"github.com/siemens/wfx/internal/producer"
	"github.com/siemens/wfx/persistence"
	"github.com/stretchr/testify/assert"
)

func TestSouthboundGetJobsIDStatusHandler_NotFound(t *testing.T) {
	params := southbound.NewGetJobsIDStatusParams()
	jobID := "42"
	params.ID = jobID
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	dbMock := persistence.NewMockStorage(t)
	dbMock.EXPECT().GetJob(params.HTTPRequest.Context(), params.ID, persistence.FetchParams{}).Return(nil, fault.Wrap(fmt.Errorf("job with id %s does not exist", jobID), ftag.With(ftag.NotFound)))

	api := NewSouthboundAPI(dbMock)
	resp := api.SouthboundGetJobsIDStatusHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusNotFound, response.StatusCode)
}

func TestSouthboundPutJobsIDStatusHandler_NotFound(t *testing.T) {
	params := southbound.NewPutJobsIDStatusParams()
	jobID := "42"
	params.ID = jobID
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	dbMock := persistence.NewMockStorage(t)
	dbMock.EXPECT().GetJob(params.HTTPRequest.Context(), params.ID, persistence.FetchParams{}).Return(nil, fault.Wrap(fmt.Errorf("job with id %s does not exist", jobID), ftag.With(ftag.NotFound)))

	api := NewSouthboundAPI(dbMock)
	resp := api.SouthboundPutJobsIDStatusHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusNotFound, response.StatusCode)
}

func TestSouthboundGetJobsHandler_InternalError(t *testing.T) {
	params := southbound.NewGetJobsParams()
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	dbMock := persistence.NewMockStorage(t)
	dbMock.EXPECT().
		QueryJobs(params.HTTPRequest.Context(), persistence.FilterParams{}, persistence.SortParams{}, persistence.PaginationParams{Limit: 10}).
		Return(nil, errors.New("something went wrong"))

	api := NewSouthboundAPI(dbMock)
	resp := api.SouthboundGetJobsHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
}

func TestSouthboundGetJobsIDHandler_NotFound(t *testing.T) {
	params := southbound.NewGetJobsIDParams()

	jobID := "42"
	params.ID = jobID
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	dbMock := persistence.NewMockStorage(t)
	dbMock.EXPECT().GetJob(params.HTTPRequest.Context(), params.ID, persistence.FetchParams{}).Return(nil, fault.Wrap(fmt.Errorf("job with id %s does not exist", jobID), ftag.With(ftag.NotFound)))

	api := NewSouthboundAPI(dbMock)
	resp := api.SouthboundGetJobsIDHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusNotFound, response.StatusCode)
}

func TestSouthboundGetJobsIDHandler_InternalError(t *testing.T) {
	params := southbound.NewGetJobsIDParams()
	jobID := "42"
	params.ID = jobID
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	history := true
	params.History = &history
	dbMock := persistence.NewMockStorage(t)
	dbMock.EXPECT().GetJob(params.HTTPRequest.Context(), params.ID, persistence.FetchParams{History: history}).Return(nil, errors.New("something went wrong"))

	api := NewSouthboundAPI(dbMock)
	resp := api.SouthboundGetJobsIDHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
}

func TestSouthboundGetJobsIDStatusHandler_InternalError(t *testing.T) {
	params := southbound.NewGetJobsIDStatusParams()
	jobID := "42"
	params.ID = jobID
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	dbMock := persistence.NewMockStorage(t)
	dbMock.EXPECT().GetJob(params.HTTPRequest.Context(), params.ID, persistence.FetchParams{}).Return(nil, errors.New("something went wrong"))

	api := NewSouthboundAPI(dbMock)
	resp := api.SouthboundGetJobsIDStatusHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
}

func TestSouthboundPutJobsIDStatusHandler_InternalError(t *testing.T) {
	params := southbound.NewPutJobsIDStatusParams()
	jobID := "42"
	params.ID = jobID
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	dbMock := persistence.NewMockStorage(t)
	dbMock.EXPECT().GetJob(params.HTTPRequest.Context(), params.ID, persistence.FetchParams{}).Return(nil, errors.New("something went wrong"))

	api := NewSouthboundAPI(dbMock)
	resp := api.SouthboundPutJobsIDStatusHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
}

func TestSouthboundGetJobsIDDefinitionHandler_NotFound(t *testing.T) {
	params := southbound.NewGetJobsIDDefinitionParams()
	jobID := "42"
	params.ID = jobID
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	dbMock := persistence.NewMockStorage(t)
	dbMock.EXPECT().GetJob(params.HTTPRequest.Context(), params.ID, persistence.FetchParams{}).Return(nil, fault.Wrap(fmt.Errorf("job with id %s does not exist", jobID), ftag.With(ftag.NotFound)))

	api := NewSouthboundAPI(dbMock)
	resp := api.SouthboundGetJobsIDDefinitionHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusNotFound, response.StatusCode)
}

func TestSouthboundGetJobsIDDefinitionHandler_InternalError(t *testing.T) {
	params := southbound.NewGetJobsIDDefinitionParams()
	jobID := "42"
	params.ID = jobID
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	dbMock := persistence.NewMockStorage(t)
	dbMock.EXPECT().GetJob(params.HTTPRequest.Context(), params.ID, persistence.FetchParams{}).Return(nil, errors.New("something went wrong"))

	api := NewSouthboundAPI(dbMock)
	resp := api.SouthboundGetJobsIDDefinitionHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
}

func TestSouthboundPutJobsIDDefinitionHandler_NotFound(t *testing.T) {
	params := southbound.NewPutJobsIDDefinitionParams()
	jobID := "42"
	params.ID = jobID
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	dbMock := persistence.NewMockStorage(t)
	dbMock.EXPECT().GetJob(params.HTTPRequest.Context(), params.ID, persistence.FetchParams{}).Return(nil, fault.Wrap(fmt.Errorf("job with id %s does not exist", jobID), ftag.With(ftag.NotFound)))

	api := NewSouthboundAPI(dbMock)
	resp := api.SouthboundPutJobsIDDefinitionHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusNotFound, response.StatusCode)
}

func TestSouthboundPutJobsIDDefinitionHandler_InternalError(t *testing.T) {
	params := southbound.NewPutJobsIDDefinitionParams()
	jobID := "42"
	params.ID = jobID
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	dbMock := persistence.NewMockStorage(t)
	dbMock.EXPECT().GetJob(params.HTTPRequest.Context(), params.ID, persistence.FetchParams{}).Return(nil, errors.New("something went wrong"))

	api := NewSouthboundAPI(dbMock)
	resp := api.SouthboundPutJobsIDDefinitionHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
}

func TestSouthboundGetWorkflowsNameHandler_InternalError(t *testing.T) {
	params := southbound.NewGetWorkflowsNameParams()
	params.Name = "wfx.test.workflow"
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	dbMock := persistence.NewMockStorage(t)
	dbMock.EXPECT().GetWorkflow(params.HTTPRequest.Context(), params.Name).Return(nil, errors.New("something went wrong"))

	api := NewSouthboundAPI(dbMock)

	resp := api.SouthboundGetWorkflowsNameHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
}

func TestSouthboundGetWorkflowsHandler_InternalError(t *testing.T) {
	params := southbound.NewGetWorkflowsParams()
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	dbMock := persistence.NewMockStorage(t)
	dbMock.EXPECT().QueryWorkflows(params.HTTPRequest.Context(), persistence.PaginationParams{Limit: 10}).Return(nil, errors.New("something went wrong"))

	api := NewSouthboundAPI(dbMock)
	resp := api.SouthboundGetWorkflowsHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
}

func TestSouthboundGetWorkflowsHandler_Empty(t *testing.T) {
	params := southbound.NewGetWorkflowsParams()
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	dbMock := persistence.NewMockStorage(t)
	dbMock.EXPECT().QueryWorkflows(params.HTTPRequest.Context(), persistence.PaginationParams{Limit: 10}).Return(nil, nil)

	api := NewSouthboundAPI(dbMock)

	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	resp := api.SouthboundGetWorkflowsHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusOK, response.StatusCode)
}

func TestSouthboundGetJobsIDTagsHandler_NotFound(t *testing.T) {
	params := southbound.NewGetJobsIDTagsParams()
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))

	dbMock := persistence.NewMockStorage(t)
	dbMock.EXPECT().GetJob(params.HTTPRequest.Context(), params.ID, persistence.FetchParams{}).Return(nil, fault.Wrap(errors.New("not found"), ftag.With(ftag.NotFound)))

	api := NewSouthboundAPI(dbMock)
	resp := api.SouthboundGetJobsIDTagsHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusNotFound, response.StatusCode)
}

func TestSouthboundGetJobsIDTagsHandler_InternalError(t *testing.T) {
	params := southbound.NewGetJobsIDTagsParams()
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))

	dbMock := persistence.NewMockStorage(t)
	dbMock.EXPECT().GetJob(params.HTTPRequest.Context(), params.ID, persistence.FetchParams{}).Return(nil, fault.Wrap(errors.New("something went wrong"), ftag.With(ftag.Internal)))

	api := NewSouthboundAPI(dbMock)
	resp := api.SouthboundGetJobsIDTagsHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
}

func TestParseFilterParamsSouth(t *testing.T) {
	jobIDs := "abc,424-194-123"
	clientIDs := "alpha,beta"
	workflows := "wf1,wf2,wf3"
	params := southbound.GetJobsEventsParams{
		HTTPRequest: &http.Request{},
		JobIds:      &jobIDs,
		ClientIds:   &clientIDs,
		Workflows:   &workflows,
	}
	filter := parseFilterParamsSouth(params)
	assert.Equal(t, []string{"abc", "424-194-123"}, filter.JobIDs)
	assert.Equal(t, []string{"alpha", "beta"}, filter.ClientIDs)
	assert.Equal(t, []string{"wf1", "wf2", "wf3"}, filter.Workflows)
}
