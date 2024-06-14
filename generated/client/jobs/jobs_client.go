// Code generated by go-swagger; DO NOT EDIT.

// SPDX-FileCopyrightText: 2023 Siemens AG
//
// SPDX-License-Identifier: Apache-2.0
//

package jobs

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
)

// New creates a new jobs API client.
func New(transport runtime.ClientTransport, formats strfmt.Registry) ClientService {
	return &Client{transport: transport, formats: formats}
}

// New creates a new jobs API client with basic auth credentials.
// It takes the following parameters:
// - host: http host (github.com).
// - basePath: any base path for the API client ("/v1", "/v3").
// - scheme: http scheme ("http", "https").
// - user: user for basic authentication header.
// - password: password for basic authentication header.
func NewClientWithBasicAuth(host, basePath, scheme, user, password string) ClientService {
	transport := httptransport.New(host, basePath, []string{scheme})
	transport.DefaultAuthentication = httptransport.BasicAuth(user, password)
	return &Client{transport: transport, formats: strfmt.Default}
}

// New creates a new jobs API client with a bearer token for authentication.
// It takes the following parameters:
// - host: http host (github.com).
// - basePath: any base path for the API client ("/v1", "/v3").
// - scheme: http scheme ("http", "https").
// - bearerToken: bearer token for Bearer authentication header.
func NewClientWithBearerToken(host, basePath, scheme, bearerToken string) ClientService {
	transport := httptransport.New(host, basePath, []string{scheme})
	transport.DefaultAuthentication = httptransport.BearerToken(bearerToken)
	return &Client{transport: transport, formats: strfmt.Default}
}

/*
Client for jobs API
*/
type Client struct {
	transport runtime.ClientTransport
	formats   strfmt.Registry
}

// ClientOption may be used to customize the behavior of Client methods.
type ClientOption func(*runtime.ClientOperation)

// This client is generated with a few options you might find useful for your swagger spec.
//
// Feel free to add you own set of options.

// WithAccept allows the client to force the Accept header
// to negotiate a specific Producer from the server.
//
// You may use this option to set arbitrary extensions to your MIME media type.
func WithAccept(mime string) ClientOption {
	return func(r *runtime.ClientOperation) {
		r.ProducesMediaTypes = []string{mime}
	}
}

// WithAcceptApplicationJSON sets the Accept header to "application/json".
func WithAcceptApplicationJSON(r *runtime.ClientOperation) {
	r.ProducesMediaTypes = []string{"application/json"}
}

// WithAcceptTextEventStream sets the Accept header to "text/event-stream".
func WithAcceptTextEventStream(r *runtime.ClientOperation) {
	r.ProducesMediaTypes = []string{"text/event-stream"}
}

// ClientService is the interface for Client methods
type ClientService interface {
	DeleteJobsID(params *DeleteJobsIDParams, opts ...ClientOption) (*DeleteJobsIDNoContent, error)

	DeleteJobsIDTags(params *DeleteJobsIDTagsParams, opts ...ClientOption) (*DeleteJobsIDTagsOK, error)

	GetJobs(params *GetJobsParams, opts ...ClientOption) (*GetJobsOK, error)

	GetJobsEvents(params *GetJobsEventsParams, opts ...ClientOption) (*GetJobsEventsOK, error)

	GetJobsID(params *GetJobsIDParams, opts ...ClientOption) (*GetJobsIDOK, error)

	GetJobsIDDefinition(params *GetJobsIDDefinitionParams, opts ...ClientOption) (*GetJobsIDDefinitionOK, error)

	GetJobsIDStatus(params *GetJobsIDStatusParams, opts ...ClientOption) (*GetJobsIDStatusOK, error)

	GetJobsIDTags(params *GetJobsIDTagsParams, opts ...ClientOption) (*GetJobsIDTagsOK, error)

	PostJobs(params *PostJobsParams, opts ...ClientOption) (*PostJobsCreated, error)

	PostJobsIDTags(params *PostJobsIDTagsParams, opts ...ClientOption) (*PostJobsIDTagsOK, error)

	PutJobsIDDefinition(params *PutJobsIDDefinitionParams, opts ...ClientOption) (*PutJobsIDDefinitionOK, error)

	PutJobsIDStatus(params *PutJobsIDStatusParams, opts ...ClientOption) (*PutJobsIDStatusOK, error)

	SetTransport(transport runtime.ClientTransport)
}

/*
DeleteJobsID deletes an existing job

Delete an existing job
*/
func (a *Client) DeleteJobsID(params *DeleteJobsIDParams, opts ...ClientOption) (*DeleteJobsIDNoContent, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewDeleteJobsIDParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "DeleteJobsID",
		Method:             "DELETE",
		PathPattern:        "/jobs/{id}",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &DeleteJobsIDReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*DeleteJobsIDNoContent)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*DeleteJobsIDDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
DeleteJobsIDTags deletes a tag

Delete a tag from an existing job
*/
func (a *Client) DeleteJobsIDTags(params *DeleteJobsIDTagsParams, opts ...ClientOption) (*DeleteJobsIDTagsOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewDeleteJobsIDTagsParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "DeleteJobsIDTags",
		Method:             "DELETE",
		PathPattern:        "/jobs/{id}/tags",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &DeleteJobsIDTagsReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*DeleteJobsIDTagsOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*DeleteJobsIDTagsDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
	GetJobs lists of job descriptions

	List of job descriptions

By default, this endpoint returns the list of jobs in a specific order and predetermined paging properties.
These defaults are:
  - Ascending sort on stime
  - 10 entries per page
*/
func (a *Client) GetJobs(params *GetJobsParams, opts ...ClientOption) (*GetJobsOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetJobsParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "GetJobs",
		Method:             "GET",
		PathPattern:        "/jobs",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &GetJobsReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetJobsOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*GetJobsDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
GetJobsEvents subscribes to job related events such as status updates

Obtain instant notifications when there are job changes matching the criteria. This endpoint utilizes server-sent events (SSE), where responses are "chunked" with double newline breaks. For example, a single event might look like this:
data: {"clientId":"example_client","state":"INSTALLING"}\n\n
*/
func (a *Client) GetJobsEvents(params *GetJobsEventsParams, opts ...ClientOption) (*GetJobsEventsOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetJobsEventsParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "GetJobsEvents",
		Method:             "GET",
		PathPattern:        "/jobs/events",
		ProducesMediaTypes: []string{"application/json", "text/event-stream"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &GetJobsEventsReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetJobsEventsOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*GetJobsEventsDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
GetJobsID jobs description for a given ID

Job description for a given ID
*/
func (a *Client) GetJobsID(params *GetJobsIDParams, opts ...ClientOption) (*GetJobsIDOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetJobsIDParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "GetJobsID",
		Method:             "GET",
		PathPattern:        "/jobs/{id}",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &GetJobsIDReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetJobsIDOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*GetJobsIDDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
GetJobsIDDefinition gets job definition

Retrieve the job definition
*/
func (a *Client) GetJobsIDDefinition(params *GetJobsIDDefinitionParams, opts ...ClientOption) (*GetJobsIDDefinitionOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetJobsIDDefinitionParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "GetJobsIDDefinition",
		Method:             "GET",
		PathPattern:        "/jobs/{id}/definition",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &GetJobsIDDefinitionReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetJobsIDDefinitionOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*GetJobsIDDefinitionDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
GetJobsIDStatus gets job status

Retrieve the job status
*/
func (a *Client) GetJobsIDStatus(params *GetJobsIDStatusParams, opts ...ClientOption) (*GetJobsIDStatusOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetJobsIDStatusParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "GetJobsIDStatus",
		Method:             "GET",
		PathPattern:        "/jobs/{id}/status",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &GetJobsIDStatusReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetJobsIDStatusOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*GetJobsIDStatusDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
GetJobsIDTags gets tags

Get the tags of a job
*/
func (a *Client) GetJobsIDTags(params *GetJobsIDTagsParams, opts ...ClientOption) (*GetJobsIDTagsOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetJobsIDTagsParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "GetJobsIDTags",
		Method:             "GET",
		PathPattern:        "/jobs/{id}/tags",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &GetJobsIDTagsReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetJobsIDTagsOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*GetJobsIDTagsDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
PostJobs adds a new job

Add a new job
*/
func (a *Client) PostJobs(params *PostJobsParams, opts ...ClientOption) (*PostJobsCreated, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewPostJobsParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "PostJobs",
		Method:             "POST",
		PathPattern:        "/jobs",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &PostJobsReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*PostJobsCreated)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*PostJobsDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
PostJobsIDTags adds a tag

Add a tag to an existing job
*/
func (a *Client) PostJobsIDTags(params *PostJobsIDTagsParams, opts ...ClientOption) (*PostJobsIDTagsOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewPostJobsIDTagsParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "PostJobsIDTags",
		Method:             "POST",
		PathPattern:        "/jobs/{id}/tags",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &PostJobsIDTagsReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*PostJobsIDTagsOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*PostJobsIDTagsDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
PutJobsIDDefinition modifies job definition

Modify the job definition of an existing job
*/
func (a *Client) PutJobsIDDefinition(params *PutJobsIDDefinitionParams, opts ...ClientOption) (*PutJobsIDDefinitionOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewPutJobsIDDefinitionParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "PutJobsIDDefinition",
		Method:             "PUT",
		PathPattern:        "/jobs/{id}/definition",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &PutJobsIDDefinitionReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*PutJobsIDDefinitionOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*PutJobsIDDefinitionDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
PutJobsIDStatus modifies status of an existing job

Modify status of an existing job
*/
func (a *Client) PutJobsIDStatus(params *PutJobsIDStatusParams, opts ...ClientOption) (*PutJobsIDStatusOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewPutJobsIDStatusParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "PutJobsIDStatus",
		Method:             "PUT",
		PathPattern:        "/jobs/{id}/status",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &PutJobsIDStatusReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*PutJobsIDStatusOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*PutJobsIDStatusDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

// SetTransport changes the transport on the client
func (a *Client) SetTransport(transport runtime.ClientTransport) {
	a.transport = transport
}
