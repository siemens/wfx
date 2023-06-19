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
	"context"
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
)

type faultyStorage struct {
	notFound      bool
	alreadyExists bool
}

func (st *faultyStorage) Initialize(context.Context, string) error {
	return nil
}

func (st *faultyStorage) Shutdown() {}

func (st *faultyStorage) CheckHealth(context.Context) error {
	return nil
}

func (st *faultyStorage) CreateJob(context.Context, *model.Job) (*model.Job, error) {
	return nil, errors.New("CreateJob failed")
}

func (st *faultyStorage) GetJob(context.Context, string, persistence.FetchParams) (*model.Job, error) {
	if st.notFound {
		return nil, fault.Wrap(errors.New("job not found"), ftag.With(ftag.NotFound))
	}
	return nil, errors.New("GetJob failed")
}

func (st *faultyStorage) UpdateJob(context.Context, *model.Job, persistence.JobUpdate) (*model.Job, error) {
	return nil, errors.New("UpdateJob failed")
}

func (st *faultyStorage) DeleteJob(context.Context, string) error {
	if st.notFound {
		return fault.Wrap(errors.New("job not found"), ftag.With(ftag.NotFound))
	}
	return errors.New("DeleteJob failed")
}

func (st *faultyStorage) QueryJobs(context.Context, persistence.FilterParams, persistence.SortParams, persistence.PaginationParams) (*model.PaginatedJobList, error) {
	return nil, errors.New("QueryJobs failed")
}

func (st *faultyStorage) CreateWorkflow(context.Context, *model.Workflow) (*model.Workflow, error) {
	if st.alreadyExists {
		return nil, fault.Wrap(errors.New("already exists"), ftag.With(ftag.AlreadyExists))
	}
	return nil, errors.New("CreateWorkflow failed")
}

func (st *faultyStorage) GetWorkflow(context.Context, string) (*model.Workflow, error) {
	if st.notFound {
		return nil, fault.Wrap(errors.New("workflow not found"), ftag.With(ftag.NotFound))
	}
	return nil, errors.New("GetWorkflow failed")
}

func (st *faultyStorage) DeleteWorkflow(context.Context, string) error {
	if st.notFound {
		return fault.Wrap(fmt.Errorf("workflow not found"), ftag.With(ftag.NotFound))
	}
	return errors.New("DeleteWorkflow failed")
}

func (st *faultyStorage) QueryWorkflows(context.Context, persistence.PaginationParams) (*model.PaginatedWorkflowList, error) {
	if st.notFound {
		return &model.PaginatedWorkflowList{}, nil
	}
	return nil, errors.New("QueryWorkflows failed")
}

func TestNorthboundGetJobsIDStatusHandler_NotFound(t *testing.T) {
	api, _ := NewNorthboundAPI(&faultyStorage{notFound: true})

	params := northbound.NewGetJobsIDStatusParams()
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))

	resp := api.NorthboundGetJobsIDStatusHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusNotFound, response.StatusCode)
}

func TestNorthboundPutJobsIDStatusHandler_NotFound(t *testing.T) {
	api, _ := NewNorthboundAPI(&faultyStorage{notFound: true})

	params := northbound.NewPutJobsIDStatusParams()
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))

	resp := api.NorthboundPutJobsIDStatusHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusNotFound, response.StatusCode)
}

func TestNorthboundGetJobsHandler_InternalError(t *testing.T) {
	api, _ := NewNorthboundAPI(&faultyStorage{})

	params := northbound.NewGetJobsParams()
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	resp := api.NorthboundGetJobsHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
}

func TestNorthboundGetJobsIDHandler_NotFound(t *testing.T) {
	api, _ := NewNorthboundAPI(&faultyStorage{notFound: true})

	params := northbound.NewGetJobsIDParams()
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	resp := api.NorthboundGetJobsIDHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusNotFound, response.StatusCode)
}

func TestNorthboundGetJobsIDHandler_InternalError(t *testing.T) {
	api, _ := NewNorthboundAPI(&faultyStorage{})

	params := northbound.NewGetJobsIDParams()
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	history := true
	params.History = &history
	resp := api.NorthboundGetJobsIDHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
}

func TestNorthboundGetJobsIDStatusHandler_InternalError(t *testing.T) {
	api, _ := NewNorthboundAPI(&faultyStorage{})

	params := northbound.NewGetJobsIDStatusParams()
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	resp := api.NorthboundGetJobsIDStatusHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
}

func TestNorthboundPutJobsIDStatusHandler_InternalError(t *testing.T) {
	api, _ := NewNorthboundAPI(&faultyStorage{})

	params := northbound.NewPutJobsIDStatusParams()
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	resp := api.NorthboundPutJobsIDStatusHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
}

func TestNorthboundGetJobsIDDefinitionHandler_NotFound(t *testing.T) {
	api, _ := NewNorthboundAPI(&faultyStorage{notFound: true})

	params := northbound.NewGetJobsIDDefinitionParams()
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	resp := api.NorthboundGetJobsIDDefinitionHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusNotFound, response.StatusCode)
}

func TestNorthboundGetJobsIDDefinitionHandler_InternalError(t *testing.T) {
	api, _ := NewNorthboundAPI(&faultyStorage{})

	params := northbound.NewGetJobsIDDefinitionParams()
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	resp := api.NorthboundGetJobsIDDefinitionHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
}

func TestNorthboundPutJobsIDDefinitionHandler_NotFound(t *testing.T) {
	api, _ := NewNorthboundAPI(&faultyStorage{notFound: true})

	params := northbound.NewPutJobsIDDefinitionParams()
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	resp := api.NorthboundPutJobsIDDefinitionHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusNotFound, response.StatusCode)
}

func TestNorthboundPutJobsIDDefinitionHandler_InternalError(t *testing.T) {
	api, _ := NewNorthboundAPI(&faultyStorage{})

	params := northbound.NewPutJobsIDDefinitionParams()
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	resp := api.NorthboundPutJobsIDDefinitionHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
}

func TestNorthboundGetWorkflowsNameHandler_InternalError(t *testing.T) {
	api, _ := NewNorthboundAPI(&faultyStorage{})

	params := northbound.NewGetWorkflowsNameParams()
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	resp := api.NorthboundGetWorkflowsNameHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
}

func TestNorthboundGetWorkflowsHandler_InternalError(t *testing.T) {
	api, _ := NewNorthboundAPI(&faultyStorage{})

	params := northbound.NewGetWorkflowsParams()
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	resp := api.NorthboundGetWorkflowsHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
}

func TestNorthboundGetWorkflowsHandler_Empty(t *testing.T) {
	api, _ := NewNorthboundAPI(&faultyStorage{notFound: true})

	params := northbound.NewGetWorkflowsParams()
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	resp := api.NorthboundGetWorkflowsHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusOK, response.StatusCode)
}

func TestNorthboundDeleteWorkflowsNameHandle_NotFound(t *testing.T) {
	api, _ := NewNorthboundAPI(&faultyStorage{notFound: true})

	params := northbound.NewDeleteWorkflowsNameParams()
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	resp := api.NorthboundDeleteWorkflowsNameHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusNotFound, response.StatusCode)
}

func TestNorthboundDeleteWorkflowsNameHandle_InternalError(t *testing.T) {
	api, _ := NewNorthboundAPI(&faultyStorage{})

	params := northbound.NewDeleteWorkflowsNameParams()
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	resp := api.NorthboundDeleteWorkflowsNameHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
}

func TestNorthboundPostWorkflowsHandler_InternalError(t *testing.T) {
	api, _ := NewNorthboundAPI(&faultyStorage{})

	params := northbound.NewPostWorkflowsParams()
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	params.Workflow = dau.DirectWorkflow()
	resp := api.NorthboundPostWorkflowsHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
}

func TestNorthboundPostWorkflowsHandler_AlreadyExists(t *testing.T) {
	api, _ := NewNorthboundAPI(&faultyStorage{alreadyExists: true})

	params := northbound.NewPostWorkflowsParams()
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	params.Workflow = dau.DirectWorkflow()
	resp := api.NorthboundPostWorkflowsHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
}

func TestNorthboundPostWorkflowsHandler_InvalidWorkflow(t *testing.T) {
	api, _ := NewNorthboundAPI(&faultyStorage{})

	params := northbound.NewPostWorkflowsParams()
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	params.Workflow = &model.Workflow{Name: "foo"}
	resp := api.NorthboundPostWorkflowsHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
}

func TestNorthboundPostJobsHandler_NotFound(t *testing.T) {
	api, _ := NewNorthboundAPI(&faultyStorage{notFound: true})

	params := northbound.NewPostJobsParams()
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	params.Job = &model.JobRequest{}
	resp := api.NorthboundPostJobsHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
}

func TestNorthboundPostJobsHandler_InternalError(t *testing.T) {
	api, _ := NewNorthboundAPI(&faultyStorage{})

	params := northbound.NewPostJobsParams()
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	params.Job = &model.JobRequest{}
	resp := api.NorthboundPostJobsHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
}

func TestNorthboundDeleteJobsIDHandler_NotFound(t *testing.T) {
	api, _ := NewNorthboundAPI(&faultyStorage{notFound: true})

	params := northbound.NewDeleteJobsIDParams()
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	resp := api.NorthboundDeleteJobsIDHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusNotFound, response.StatusCode)
}

func TestNorthboundDeleteJobsIDHandler_InternalError(t *testing.T) {
	api, _ := NewNorthboundAPI(&faultyStorage{})

	params := northbound.NewDeleteJobsIDParams()
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	resp := api.NorthboundDeleteJobsIDHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
}

func TestNorthboundGetJobsIDTagsHandler_NotFound(t *testing.T) {
	api, _ := NewNorthboundAPI(&faultyStorage{notFound: true})

	params := northbound.NewGetJobsIDTagsParams()
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	resp := api.NorthboundGetJobsIDTagsHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusNotFound, response.StatusCode)
}

func TestNorthboundGetJobsIDTagsHandler_InternalError(t *testing.T) {
	api, _ := NewNorthboundAPI(&faultyStorage{})

	params := northbound.NewGetJobsIDTagsParams()
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	resp := api.NorthboundGetJobsIDTagsHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
}

func TestNorthboundPostJobsIDTagsHandler_NotFound(t *testing.T) {
	api, _ := NewNorthboundAPI(&faultyStorage{notFound: true})

	params := northbound.NewPostJobsIDTagsParams()
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	resp := api.NorthboundPostJobsIDTagsHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusNotFound, response.StatusCode)
}

func TestNorthboundPostJobsIDTagsHandler_InternalError(t *testing.T) {
	api, _ := NewNorthboundAPI(&faultyStorage{})

	params := northbound.NewPostJobsIDTagsParams()
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	resp := api.NorthboundPostJobsIDTagsHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
}

func TestNorthboundDeleteJobsIDTagsHandler_NotFound(t *testing.T) {
	api, _ := NewNorthboundAPI(&faultyStorage{notFound: true})

	params := northbound.NewDeleteJobsIDTagsParams()
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	resp := api.NorthboundDeleteJobsIDTagsHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusNotFound, response.StatusCode)
}

func TestNorthboundDeleteJobsIDTagsHandler_InternalError(t *testing.T) {
	api, _ := NewNorthboundAPI(&faultyStorage{})

	params := northbound.NewDeleteJobsIDTagsParams()
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	resp := api.NorthboundDeleteJobsIDTagsHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
}
