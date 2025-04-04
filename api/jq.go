package api

/*
 * SPDX-FileCopyrightText: 2024 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"encoding/json"
	"net/http"

	"github.com/Southclaws/fault"
	"github.com/itchyny/gojq"
	"github.com/rs/zerolog/log"
)

// JQFilter applies a JQ filter to the response body.
type JQFilter struct {
	filter string
	body   any
}

func NewJQFilter(filter string, body any) JQFilter {
	return JQFilter{filter: filter, body: body}
}

func applyFilter(w http.ResponseWriter, body any, filter string) error {
	contextLogger := log.With().Str("filter", filter).Logger()
	contextLogger.Debug().Msg("Applying JQ filter")

	query, err := gojq.Parse(filter)
	if err != nil {
		contextLogger.Err(err).Msg("Failed to parse JQ filter")
		return fault.Wrap(err)
	}

	jsonData, err := json.Marshal(body)
	if err != nil {
		return fault.Wrap(err)
	}

	var input any
	// need to unmarshal again, but to type 'any'; this cannot fail
	// because we own the local variable jsonData and know it's
	// valid JSON
	_ = json.Unmarshal(jsonData, &input)

	w.Header().Set("X-Response-Filter", filter)
	iter := query.Run(input)
	ok := true
	encoder := json.NewEncoder(w)
	for ok {
		var v any
		v, ok = iter.Next()
		if ok {
			// this cannot fail because we know the input
			_ = encoder.Encode(v)
		}
	}
	return nil
}

// the below methods ensure we fulfill all the interfaces generated by oapi-codegen

//revive:disable:var-naming
func (jq JQFilter) VisitGetHealthResponse(w http.ResponseWriter) error {
	return applyFilter(w, jq.body, jq.filter)
}

func (jq JQFilter) VisitGetJobsResponse(w http.ResponseWriter) error {
	return applyFilter(w, jq.body, jq.filter)
}

func (jq JQFilter) VisitPostJobsResponse(w http.ResponseWriter) error {
	return applyFilter(w, jq.body, jq.filter)
}

func (jq JQFilter) VisitGetJobsEventsResponse(w http.ResponseWriter) error {
	return applyFilter(w, jq.body, jq.filter)
}

func (jq JQFilter) VisitDeleteJobsIdResponse(w http.ResponseWriter) error {
	return applyFilter(w, jq.body, jq.filter)
}

func (jq JQFilter) VisitGetJobsIdResponse(w http.ResponseWriter) error {
	return applyFilter(w, jq.body, jq.filter)
}

func (jq JQFilter) VisitGetJobsIdDefinitionResponse(w http.ResponseWriter) error {
	return applyFilter(w, jq.body, jq.filter)
}

func (jq JQFilter) VisitPutJobsIdDefinitionResponse(w http.ResponseWriter) error {
	return applyFilter(w, jq.body, jq.filter)
}

func (jq JQFilter) VisitGetJobsIdStatusResponse(w http.ResponseWriter) error {
	return applyFilter(w, jq.body, jq.filter)
}

func (jq JQFilter) VisitPutJobsIdStatusResponse(w http.ResponseWriter) error {
	return applyFilter(w, jq.body, jq.filter)
}

func (jq JQFilter) VisitDeleteJobsIdTagsResponse(w http.ResponseWriter) error {
	return applyFilter(w, jq.body, jq.filter)
}

func (jq JQFilter) VisitGetJobsIdTagsResponse(w http.ResponseWriter) error {
	return applyFilter(w, jq.body, jq.filter)
}

func (jq JQFilter) VisitPostJobsIdTagsResponse(w http.ResponseWriter) error {
	return applyFilter(w, jq.body, jq.filter)
}

func (jq JQFilter) VisitGetVersionResponse(w http.ResponseWriter) error {
	return applyFilter(w, jq.body, jq.filter)
}

func (jq JQFilter) VisitGetWorkflowsResponse(w http.ResponseWriter) error {
	return applyFilter(w, jq.body, jq.filter)
}

func (jq JQFilter) VisitPostWorkflowsResponse(w http.ResponseWriter) error {
	return applyFilter(w, jq.body, jq.filter)
}

func (jq JQFilter) VisitDeleteWorkflowsNameResponse(w http.ResponseWriter) error {
	return applyFilter(w, jq.body, jq.filter)
}

func (jq JQFilter) VisitGetWorkflowsNameResponse(w http.ResponseWriter) error {
	return applyFilter(w, jq.body, jq.filter)
}
