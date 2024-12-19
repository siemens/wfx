package api

/*
 * SPDX-FileCopyrightText: 2024 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"context"

	"github.com/siemens/wfx/generated/api"
)

// ensure that we fulfill the interface (compile-time check)
var _ api.StrictServerInterface = (*SouthboundServer)(nil)

type SouthboundServer struct {
	wfx *WfxServer
}

func NewSouthboundServer(wfx *WfxServer) SouthboundServer {
	return SouthboundServer{wfx: wfx}
}

//revive:disable:var-naming
func (south SouthboundServer) GetJobs(ctx context.Context, request api.GetJobsRequestObject) (api.GetJobsResponseObject, error) {
	return south.wfx.GetJobs(ctx, request)
}

func (south SouthboundServer) PostJobs(context.Context, api.PostJobsRequestObject) (api.PostJobsResponseObject, error) {
	return api.PostJobs403Response{}, nil
}

func (south SouthboundServer) GetJobsEvents(ctx context.Context, request api.GetJobsEventsRequestObject) (api.GetJobsEventsResponseObject, error) {
	return south.wfx.GetJobsEvents(ctx, request)
}

func (south SouthboundServer) DeleteJobsId(context.Context, api.DeleteJobsIdRequestObject) (api.DeleteJobsIdResponseObject, error) {
	return api.DeleteJobsId403Response{}, nil
}

func (south SouthboundServer) GetJobsId(ctx context.Context, request api.GetJobsIdRequestObject) (api.GetJobsIdResponseObject, error) {
	return south.wfx.GetJobsId(ctx, request)
}

func (south SouthboundServer) GetJobsIdDefinition(ctx context.Context, request api.GetJobsIdDefinitionRequestObject) (api.GetJobsIdDefinitionResponseObject, error) {
	return south.wfx.GetJobsIdDefinition(ctx, request)
}

func (south SouthboundServer) PutJobsIdDefinition(ctx context.Context, request api.PutJobsIdDefinitionRequestObject) (api.PutJobsIdDefinitionResponseObject, error) {
	return south.wfx.PutJobsIdDefinition(ctx, request)
}

func (south SouthboundServer) GetJobsIdStatus(ctx context.Context, request api.GetJobsIdStatusRequestObject) (api.GetJobsIdStatusResponseObject, error) {
	return south.wfx.GetJobsIdStatus(ctx, request)
}

func (south SouthboundServer) PutJobsIdStatus(ctx context.Context, request api.PutJobsIdStatusRequestObject) (api.PutJobsIdStatusResponseObject, error) {
	return south.wfx.PutJobsIdStatus(ctx, request, api.CLIENT)
}

func (south SouthboundServer) DeleteJobsIdTags(context.Context, api.DeleteJobsIdTagsRequestObject) (api.DeleteJobsIdTagsResponseObject, error) {
	return api.DeleteJobsIdTags403Response{}, nil
}

func (south SouthboundServer) GetJobsIdTags(ctx context.Context, request api.GetJobsIdTagsRequestObject) (api.GetJobsIdTagsResponseObject, error) {
	return south.wfx.GetJobsIdTags(ctx, request)
}

func (south SouthboundServer) PostJobsIdTags(context.Context, api.PostJobsIdTagsRequestObject) (api.PostJobsIdTagsResponseObject, error) {
	return api.PostJobsIdTags403Response{}, nil
}

func (south SouthboundServer) GetWorkflows(ctx context.Context, request api.GetWorkflowsRequestObject) (api.GetWorkflowsResponseObject, error) {
	return south.wfx.GetWorkflows(ctx, request)
}

func (south SouthboundServer) PostWorkflows(context.Context, api.PostWorkflowsRequestObject) (api.PostWorkflowsResponseObject, error) {
	return api.PostWorkflows403Response{}, nil
}

func (south SouthboundServer) DeleteWorkflowsName(context.Context, api.DeleteWorkflowsNameRequestObject) (api.DeleteWorkflowsNameResponseObject, error) {
	return api.DeleteWorkflowsName403Response{}, nil
}

func (south SouthboundServer) GetWorkflowsName(ctx context.Context, request api.GetWorkflowsNameRequestObject) (api.GetWorkflowsNameResponseObject, error) {
	return south.wfx.GetWorkflowsName(ctx, request)
}

func (south SouthboundServer) GetHealth(ctx context.Context, request api.GetHealthRequestObject) (api.GetHealthResponseObject, error) {
	return south.wfx.GetHealth(ctx, request)
}

func (south SouthboundServer) GetVersion(ctx context.Context, request api.GetVersionRequestObject) (api.GetVersionResponseObject, error) {
	return south.wfx.GetVersion(ctx, request)
}
