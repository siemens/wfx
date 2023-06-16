// Code generated by go-swagger; DO NOT EDIT.

// SPDX-FileCopyrightText: 2023 Siemens AG
//
// SPDX-License-Identifier: Apache-2.0
//

package jobs

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"net/http"
	"time"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
)

// NewDeleteJobsIDTagsParams creates a new DeleteJobsIDTagsParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewDeleteJobsIDTagsParams() *DeleteJobsIDTagsParams {
	return &DeleteJobsIDTagsParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewDeleteJobsIDTagsParamsWithTimeout creates a new DeleteJobsIDTagsParams object
// with the ability to set a timeout on a request.
func NewDeleteJobsIDTagsParamsWithTimeout(timeout time.Duration) *DeleteJobsIDTagsParams {
	return &DeleteJobsIDTagsParams{
		timeout: timeout,
	}
}

// NewDeleteJobsIDTagsParamsWithContext creates a new DeleteJobsIDTagsParams object
// with the ability to set a context for a request.
func NewDeleteJobsIDTagsParamsWithContext(ctx context.Context) *DeleteJobsIDTagsParams {
	return &DeleteJobsIDTagsParams{
		Context: ctx,
	}
}

// NewDeleteJobsIDTagsParamsWithHTTPClient creates a new DeleteJobsIDTagsParams object
// with the ability to set a custom HTTPClient for a request.
func NewDeleteJobsIDTagsParamsWithHTTPClient(client *http.Client) *DeleteJobsIDTagsParams {
	return &DeleteJobsIDTagsParams{
		HTTPClient: client,
	}
}

/*
DeleteJobsIDTagsParams contains all the parameters to send to the API endpoint

	for the delete jobs ID tags operation.

	Typically these are written to a http.Request.
*/
type DeleteJobsIDTagsParams struct {

	/* Tags.

	   Tags to add
	*/
	Tags []string

	/* ID.

	   Job ID
	*/
	ID string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the delete jobs ID tags params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *DeleteJobsIDTagsParams) WithDefaults() *DeleteJobsIDTagsParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the delete jobs ID tags params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *DeleteJobsIDTagsParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the delete jobs ID tags params
func (o *DeleteJobsIDTagsParams) WithTimeout(timeout time.Duration) *DeleteJobsIDTagsParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the delete jobs ID tags params
func (o *DeleteJobsIDTagsParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the delete jobs ID tags params
func (o *DeleteJobsIDTagsParams) WithContext(ctx context.Context) *DeleteJobsIDTagsParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the delete jobs ID tags params
func (o *DeleteJobsIDTagsParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the delete jobs ID tags params
func (o *DeleteJobsIDTagsParams) WithHTTPClient(client *http.Client) *DeleteJobsIDTagsParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the delete jobs ID tags params
func (o *DeleteJobsIDTagsParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithTags adds the tags to the delete jobs ID tags params
func (o *DeleteJobsIDTagsParams) WithTags(tags []string) *DeleteJobsIDTagsParams {
	o.SetTags(tags)
	return o
}

// SetTags adds the tags to the delete jobs ID tags params
func (o *DeleteJobsIDTagsParams) SetTags(tags []string) {
	o.Tags = tags
}

// WithID adds the id to the delete jobs ID tags params
func (o *DeleteJobsIDTagsParams) WithID(id string) *DeleteJobsIDTagsParams {
	o.SetID(id)
	return o
}

// SetID adds the id to the delete jobs ID tags params
func (o *DeleteJobsIDTagsParams) SetID(id string) {
	o.ID = id
}

// WriteToRequest writes these params to a swagger request
func (o *DeleteJobsIDTagsParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error
	if o.Tags != nil {
		if err := r.SetBodyParam(o.Tags); err != nil {
			return err
		}
	}

	// path param id
	if err := r.SetPathParam("id", o.ID); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}