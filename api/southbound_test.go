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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/siemens/wfx/generated/southbound/restapi/operations/southbound"
	"github.com/siemens/wfx/internal/producer"
	"github.com/stretchr/testify/assert"
)

func TestSouthboundGetJobsIDStatusHandler_NotFound(t *testing.T) {
	api := NewSouthboundAPI(&faultyStorage{notFound: true})

	params := southbound.NewGetJobsIDStatusParams()
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))

	resp := api.SouthboundGetJobsIDStatusHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusNotFound, response.StatusCode)
}

func TestSouthboundPutJobsIDStatusHandler_NotFound(t *testing.T) {
	api := NewSouthboundAPI(&faultyStorage{notFound: true})

	params := southbound.NewPutJobsIDStatusParams()
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))

	resp := api.SouthboundPutJobsIDStatusHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusNotFound, response.StatusCode)
}

func TestSouthboundGetJobsHandler_InternalError(t *testing.T) {
	api := NewSouthboundAPI(&faultyStorage{})

	params := southbound.NewGetJobsParams()
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	resp := api.SouthboundGetJobsHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
}

func TestSouthboundGetJobsIDHandler_NotFound(t *testing.T) {
	api := NewSouthboundAPI(&faultyStorage{notFound: true})

	params := southbound.NewGetJobsIDParams()
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	resp := api.SouthboundGetJobsIDHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusNotFound, response.StatusCode)
}

func TestSouthboundGetJobsIDHandler_InternalError(t *testing.T) {
	api := NewSouthboundAPI(&faultyStorage{})

	params := southbound.NewGetJobsIDParams()
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	history := true
	params.History = &history
	resp := api.SouthboundGetJobsIDHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
}

func TestSouthboundGetJobsIDStatusHandler_InternalError(t *testing.T) {
	api := NewSouthboundAPI(&faultyStorage{})

	params := southbound.NewGetJobsIDStatusParams()
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	resp := api.SouthboundGetJobsIDStatusHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
}

func TestSouthboundPutJobsIDStatusHandler_InternalError(t *testing.T) {
	api := NewSouthboundAPI(&faultyStorage{})

	params := southbound.NewPutJobsIDStatusParams()
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	resp := api.SouthboundPutJobsIDStatusHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
}

func TestSouthboundGetJobsIDDefinitionHandler_NotFound(t *testing.T) {
	api := NewSouthboundAPI(&faultyStorage{notFound: true})

	params := southbound.NewGetJobsIDDefinitionParams()
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	resp := api.SouthboundGetJobsIDDefinitionHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusNotFound, response.StatusCode)
}

func TestSouthboundGetJobsIDDefinitionHandler_InternalError(t *testing.T) {
	api := NewSouthboundAPI(&faultyStorage{})

	params := southbound.NewGetJobsIDDefinitionParams()
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	resp := api.SouthboundGetJobsIDDefinitionHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
}

func TestSouthboundPutJobsIDDefinitionHandler_NotFound(t *testing.T) {
	api := NewSouthboundAPI(&faultyStorage{notFound: true})

	params := southbound.NewPutJobsIDDefinitionParams()
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	resp := api.SouthboundPutJobsIDDefinitionHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusNotFound, response.StatusCode)
}

func TestSouthboundPutJobsIDDefinitionHandler_InternalError(t *testing.T) {
	api := NewSouthboundAPI(&faultyStorage{})

	params := southbound.NewPutJobsIDDefinitionParams()
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	resp := api.SouthboundPutJobsIDDefinitionHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
}

func TestSouthboundGetWorkflowsNameHandler_InternalError(t *testing.T) {
	api := NewSouthboundAPI(&faultyStorage{})

	params := southbound.NewGetWorkflowsNameParams()
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	resp := api.SouthboundGetWorkflowsNameHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
}

func TestSouthboundGetWorkflowsHandler_InternalError(t *testing.T) {
	api := NewSouthboundAPI(&faultyStorage{})

	params := southbound.NewGetWorkflowsParams()
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	resp := api.SouthboundGetWorkflowsHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
}

func TestSouthboundGetWorkflowsHandler_Empty(t *testing.T) {
	api := NewSouthboundAPI(&faultyStorage{notFound: true})

	params := southbound.NewGetWorkflowsParams()
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	resp := api.SouthboundGetWorkflowsHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusOK, response.StatusCode)
}

func TestSouthboundGetJobsIDTagsHandler_NotFound(t *testing.T) {
	api := NewSouthboundAPI(&faultyStorage{notFound: true})

	params := southbound.NewGetJobsIDTagsParams()
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	resp := api.SouthboundGetJobsIDTagsHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusNotFound, response.StatusCode)
}

func TestSouthboundGetJobsIDTagsHandler_InternalError(t *testing.T) {
	api := NewSouthboundAPI(&faultyStorage{})

	params := southbound.NewGetJobsIDTagsParams()
	params.HTTPRequest = httptest.NewRequest(http.MethodGet, "http://localhost", new(bytes.Buffer))
	resp := api.SouthboundGetJobsIDTagsHandler.Handle(params)

	recorder := httptest.NewRecorder()
	resp.WriteResponse(recorder, producer.JSONProducer())
	response := recorder.Result()

	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
}
