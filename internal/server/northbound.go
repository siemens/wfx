package server

/*
 * SPDX-FileCopyrightText: 2024 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"context"

	"github.com/Southclaws/fault"
	wfxAPI "github.com/siemens/wfx/api"
	"github.com/siemens/wfx/generated/api"
)

// ensure that we fulfill the interface (compile-time check)
var _ api.StrictServerInterface = (*NorthboundServer)(nil)

type NorthboundServer struct {
	wfx api.StrictServerInterface
}

func NewNorthboundServer(wfx api.StrictServerInterface) NorthboundServer {
	return NorthboundServer{wfx: wfx}
}

//revive:disable:var-naming
func (north NorthboundServer) GetJobs(ctx context.Context, request api.GetJobsRequestObject) (api.GetJobsResponseObject, error) {
	resp, err := north.wfx.GetJobs(ctx, request)
	if err != nil {
		return nil, fault.Wrap(err)
	}
	return resp, nil
}

func (north NorthboundServer) PostJobs(ctx context.Context, request api.PostJobsRequestObject) (api.PostJobsResponseObject, error) {
	resp, err := north.wfx.PostJobs(ctx, request)
	if err != nil {
		return nil, fault.Wrap(err)
	}
	return resp, nil
}

func (north NorthboundServer) GetJobsEvents(ctx context.Context, request api.GetJobsEventsRequestObject) (api.GetJobsEventsResponseObject, error) {
	resp, err := north.wfx.GetJobsEvents(ctx, request)
	if err != nil {
		return nil, fault.Wrap(err)
	}
	return resp, nil
}

func (north NorthboundServer) DeleteJobsId(ctx context.Context, request api.DeleteJobsIdRequestObject) (api.DeleteJobsIdResponseObject, error) {
	resp, err := north.wfx.DeleteJobsId(ctx, request)
	if err != nil {
		return nil, fault.Wrap(err)
	}
	return resp, nil
}

func (north NorthboundServer) GetJobsId(ctx context.Context, request api.GetJobsIdRequestObject) (api.GetJobsIdResponseObject, error) {
	resp, err := north.wfx.GetJobsId(ctx, request)
	if err != nil {
		return nil, fault.Wrap(err)
	}
	return resp, nil
}

func (north NorthboundServer) GetJobsIdDefinition(ctx context.Context, request api.GetJobsIdDefinitionRequestObject) (api.GetJobsIdDefinitionResponseObject, error) {
	resp, err := north.wfx.GetJobsIdDefinition(ctx, request)
	if err != nil {
		return nil, fault.Wrap(err)
	}
	return resp, nil
}

func (north NorthboundServer) PutJobsIdDefinition(ctx context.Context, request api.PutJobsIdDefinitionRequestObject) (api.PutJobsIdDefinitionResponseObject, error) {
	resp, err := north.wfx.PutJobsIdDefinition(ctx, request)
	if err != nil {
		return nil, fault.Wrap(err)
	}
	return resp, nil
}

func (north NorthboundServer) GetJobsIdStatus(ctx context.Context, request api.GetJobsIdStatusRequestObject) (api.GetJobsIdStatusResponseObject, error) {
	resp, err := north.wfx.GetJobsIdStatus(ctx, request)
	if err != nil {
		return nil, fault.Wrap(err)
	}
	return resp, nil
}

func (north NorthboundServer) PutJobsIdStatus(ctx context.Context, request api.PutJobsIdStatusRequestObject) (api.PutJobsIdStatusResponseObject, error) {
	resp, err := north.wfx.PutJobsIdStatus(context.WithValue(ctx, wfxAPI.EligibleKey, api.WFX), request)
	if err != nil {
		return nil, fault.Wrap(err)
	}
	return resp, nil
}

func (north NorthboundServer) DeleteJobsIdTags(ctx context.Context, request api.DeleteJobsIdTagsRequestObject) (api.DeleteJobsIdTagsResponseObject, error) {
	resp, err := north.wfx.DeleteJobsIdTags(ctx, request)
	if err != nil {
		return nil, fault.Wrap(err)
	}
	return resp, nil
}

func (north NorthboundServer) GetJobsIdTags(ctx context.Context, request api.GetJobsIdTagsRequestObject) (api.GetJobsIdTagsResponseObject, error) {
	resp, err := north.wfx.GetJobsIdTags(ctx, request)
	if err != nil {
		return nil, fault.Wrap(err)
	}
	return resp, nil
}

func (north NorthboundServer) PostJobsIdTags(ctx context.Context, request api.PostJobsIdTagsRequestObject) (api.PostJobsIdTagsResponseObject, error) {
	resp, err := north.wfx.PostJobsIdTags(ctx, request)
	if err != nil {
		return nil, fault.Wrap(err)
	}
	return resp, nil
}

func (north NorthboundServer) GetWorkflows(ctx context.Context, request api.GetWorkflowsRequestObject) (api.GetWorkflowsResponseObject, error) {
	resp, err := north.wfx.GetWorkflows(ctx, request)
	if err != nil {
		return nil, fault.Wrap(err)
	}
	return resp, nil
}

func (north NorthboundServer) PostWorkflows(ctx context.Context, request api.PostWorkflowsRequestObject) (api.PostWorkflowsResponseObject, error) {
	resp, err := north.wfx.PostWorkflows(ctx, request)
	if err != nil {
		return nil, fault.Wrap(err)
	}
	return resp, nil
}

func (north NorthboundServer) DeleteWorkflowsName(ctx context.Context, request api.DeleteWorkflowsNameRequestObject) (api.DeleteWorkflowsNameResponseObject, error) {
	resp, err := north.wfx.DeleteWorkflowsName(ctx, request)
	if err != nil {
		return nil, fault.Wrap(err)
	}
	return resp, nil
}

func (north NorthboundServer) GetWorkflowsName(ctx context.Context, request api.GetWorkflowsNameRequestObject) (api.GetWorkflowsNameResponseObject, error) {
	resp, err := north.wfx.GetWorkflowsName(ctx, request)
	if err != nil {
		return nil, fault.Wrap(err)
	}
	return resp, nil
}

func (north NorthboundServer) GetHealth(ctx context.Context, request api.GetHealthRequestObject) (api.GetHealthResponseObject, error) {
	resp, err := north.wfx.GetHealth(ctx, request)
	if err != nil {
		return nil, fault.Wrap(err)
	}
	return resp, nil
}

func (north NorthboundServer) GetVersion(ctx context.Context, request api.GetVersionRequestObject) (api.GetVersionResponseObject, error) {
	resp, err := north.wfx.GetVersion(ctx, request)
	if err != nil {
		return nil, fault.Wrap(err)
	}
	return resp, nil
}
