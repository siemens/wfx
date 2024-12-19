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
var _ api.StrictServerInterface = (*NorthboundServer)(nil)

type NorthboundServer struct {
	wfx *WfxServer
}

func NewNorthboundServer(wfx *WfxServer) NorthboundServer {
	return NorthboundServer{wfx: wfx}
}

//revive:disable:var-naming
func (north NorthboundServer) GetJobs(ctx context.Context, request api.GetJobsRequestObject) (api.GetJobsResponseObject, error) {
	return north.wfx.GetJobs(ctx, request)
}

func (north NorthboundServer) PostJobs(ctx context.Context, request api.PostJobsRequestObject) (api.PostJobsResponseObject, error) {
	return north.wfx.PostJobs(ctx, request)
}

func (north NorthboundServer) GetJobsEvents(ctx context.Context, request api.GetJobsEventsRequestObject) (api.GetJobsEventsResponseObject, error) {
	return north.wfx.GetJobsEvents(ctx, request)
}

func (north NorthboundServer) DeleteJobsId(ctx context.Context, request api.DeleteJobsIdRequestObject) (api.DeleteJobsIdResponseObject, error) {
	return north.wfx.DeleteJobsId(ctx, request)
}

func (north NorthboundServer) GetJobsId(ctx context.Context, request api.GetJobsIdRequestObject) (api.GetJobsIdResponseObject, error) {
	return north.wfx.GetJobsId(ctx, request)
}

func (north NorthboundServer) GetJobsIdDefinition(ctx context.Context, request api.GetJobsIdDefinitionRequestObject) (api.GetJobsIdDefinitionResponseObject, error) {
	return north.wfx.GetJobsIdDefinition(ctx, request)
}

func (north NorthboundServer) PutJobsIdDefinition(ctx context.Context, request api.PutJobsIdDefinitionRequestObject) (api.PutJobsIdDefinitionResponseObject, error) {
	return north.wfx.PutJobsIdDefinition(ctx, request)
}

func (north NorthboundServer) GetJobsIdStatus(ctx context.Context, request api.GetJobsIdStatusRequestObject) (api.GetJobsIdStatusResponseObject, error) {
	return north.wfx.GetJobsIdStatus(ctx, request)
}

func (north NorthboundServer) PutJobsIdStatus(ctx context.Context, request api.PutJobsIdStatusRequestObject) (api.PutJobsIdStatusResponseObject, error) {
	return north.wfx.PutJobsIdStatus(ctx, request, api.WFX)
}

func (north NorthboundServer) DeleteJobsIdTags(ctx context.Context, request api.DeleteJobsIdTagsRequestObject) (api.DeleteJobsIdTagsResponseObject, error) {
	return north.wfx.DeleteJobsIdTags(ctx, request)
}

func (north NorthboundServer) GetJobsIdTags(ctx context.Context, request api.GetJobsIdTagsRequestObject) (api.GetJobsIdTagsResponseObject, error) {
	return north.wfx.GetJobsIdTags(ctx, request)
}

func (north NorthboundServer) PostJobsIdTags(ctx context.Context, request api.PostJobsIdTagsRequestObject) (api.PostJobsIdTagsResponseObject, error) {
	return north.wfx.PostJobsIdTags(ctx, request)
}

func (north NorthboundServer) GetWorkflows(ctx context.Context, request api.GetWorkflowsRequestObject) (api.GetWorkflowsResponseObject, error) {
	return north.wfx.GetWorkflows(ctx, request)
}

func (north NorthboundServer) PostWorkflows(ctx context.Context, request api.PostWorkflowsRequestObject) (api.PostWorkflowsResponseObject, error) {
	return north.wfx.PostWorkflows(ctx, request)
}

func (north NorthboundServer) DeleteWorkflowsName(ctx context.Context, request api.DeleteWorkflowsNameRequestObject) (api.DeleteWorkflowsNameResponseObject, error) {
	return north.wfx.DeleteWorkflowsName(ctx, request)
}

func (north NorthboundServer) GetWorkflowsName(ctx context.Context, request api.GetWorkflowsNameRequestObject) (api.GetWorkflowsNameResponseObject, error) {
	return north.wfx.GetWorkflowsName(ctx, request)
}

func (north NorthboundServer) GetHealth(ctx context.Context, request api.GetHealthRequestObject) (api.GetHealthResponseObject, error) {
	return north.wfx.GetHealth(ctx, request)
}

func (north NorthboundServer) GetVersion(ctx context.Context, request api.GetVersionRequestObject) (api.GetVersionResponseObject, error) {
	return north.wfx.GetVersion(ctx, request)
}
