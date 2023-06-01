package api

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"net/http"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/ftag"
	"github.com/go-openapi/loads"
	"github.com/go-openapi/runtime/middleware"
	"github.com/siemens/wfx/generated/model"
	"github.com/siemens/wfx/generated/southbound/restapi"
	"github.com/siemens/wfx/generated/southbound/restapi/operations"
	"github.com/siemens/wfx/generated/southbound/restapi/operations/southbound"
	"github.com/siemens/wfx/internal/handler/job"
	"github.com/siemens/wfx/internal/handler/job/definition"
	"github.com/siemens/wfx/internal/handler/job/status"
	"github.com/siemens/wfx/internal/handler/job/tags"
	"github.com/siemens/wfx/internal/handler/workflow"
	"github.com/siemens/wfx/middleware/logging"
	"github.com/siemens/wfx/persistence"
)

func NewSouthboundAPI(storage persistence.Storage) (*operations.WorkflowExecutorAPI, error) {
	swaggerSpec, err := loads.Embedded(restapi.SwaggerJSON, restapi.FlatSwaggerJSON)
	if err != nil {
		return nil, fault.Wrap(err)
	}
	serverAPI := operations.NewWorkflowExecutorAPI(swaggerSpec)

	serverAPI.SouthboundGetJobsHandler = southbound.GetJobsHandlerFunc(
		func(params southbound.GetJobsParams) middleware.Responder {
			filterParams := persistence.FilterParams{
				ClientID: params.ClientID,
				Group:    params.Group,
				State:    params.State,
				Workflow: params.Workflow,
				Tags:     params.Tag,
			}
			paginationParams := persistence.PaginationParams{
				Offset: *params.Offset,
				Limit:  *params.Limit,
			}
			jobs, err := job.QueryJobs(params.HTTPRequest.Context(), storage,
				filterParams, paginationParams, params.Sort)
			if err != nil {
				return southbound.NewGetJobsDefault(http.StatusInternalServerError)
			}
			return southbound.NewGetJobsOK().WithPayload(jobs)
		})
	serverAPI.SouthboundGetJobsIDHandler = southbound.GetJobsIDHandlerFunc(
		func(params southbound.GetJobsIDParams) middleware.Responder {
			history := false
			if params.History != nil {
				history = *params.History
			}
			job, err := job.GetJob(params.HTTPRequest.Context(), storage, params.ID, history)
			if ftag.Get(err) == ftag.NotFound {
				return southbound.NewGetJobsIDNotFound().WithPayload(&model.ErrorResponse{
					Errors: []*model.Error{&JobNotFound},
				})
			}
			if err != nil {
				return southbound.NewGetJobsIDDefault(http.StatusInternalServerError)
			}
			return southbound.NewGetJobsIDOK().WithPayload(job)
		})

	serverAPI.SouthboundGetJobsIDStatusHandler = southbound.GetJobsIDStatusHandlerFunc(
		func(params southbound.GetJobsIDStatusParams) middleware.Responder {
			status, err := status.Get(params.HTTPRequest.Context(), storage, params.ID)
			if ftag.Get(err) == ftag.NotFound {
				return southbound.NewGetJobsIDStatusNotFound().WithPayload(&model.ErrorResponse{
					Errors: []*model.Error{&JobNotFound},
				})
			}
			if err != nil {
				return southbound.NewGetJobsIDStatusDefault(http.StatusInternalServerError)
			}
			return southbound.NewGetJobsIDStatusOK().WithPayload(status)
		})

	serverAPI.SouthboundPutJobsIDStatusHandler = southbound.PutJobsIDStatusHandlerFunc(
		func(params southbound.PutJobsIDStatusParams) middleware.Responder {
			status, err := status.Update(params.HTTPRequest.Context(), storage, params.ID, params.NewJobStatus, model.EligibleEnumCLIENT)
			if err != nil {
				switch ftag.Get(err) {
				case ftag.NotFound:
					return southbound.NewPutJobsIDStatusNotFound().WithPayload(&model.ErrorResponse{
						Errors: []*model.Error{&JobNotFound},
					})
				case ftag.InvalidArgument:
					err2 := InvalidRequest
					err2.Message = err.Error()
					return southbound.NewPutJobsIDStatusBadRequest().WithPayload(&model.ErrorResponse{Errors: []*model.Error{&err2}})
				default:
					return southbound.NewPutJobsIDStatusDefault(http.StatusInternalServerError)
				}
			}
			return southbound.NewPutJobsIDStatusOK().WithPayload(status)
		})

	serverAPI.SouthboundGetJobsIDDefinitionHandler = southbound.GetJobsIDDefinitionHandlerFunc(
		func(params southbound.GetJobsIDDefinitionParams) middleware.Responder {
			definition, err := definition.Get(params.HTTPRequest.Context(), storage, params.ID)
			if ftag.Get(err) == ftag.NotFound {
				return southbound.NewGetJobsIDDefinitionNotFound().WithPayload(&model.ErrorResponse{
					Errors: []*model.Error{&JobNotFound},
				})
			}
			if err != nil {
				return southbound.NewGetJobsIDDefinitionDefault(http.StatusInternalServerError)
			}
			return southbound.NewGetJobsIDDefinitionOK().WithPayload(definition)
		})
	serverAPI.SouthboundPutJobsIDDefinitionHandler = southbound.PutJobsIDDefinitionHandlerFunc(
		func(params southbound.PutJobsIDDefinitionParams) middleware.Responder {
			definition, err := definition.Update(params.HTTPRequest.Context(), storage, params.ID, params.JobDefinition)
			if ftag.Get(err) == ftag.NotFound {
				return southbound.NewPutJobsIDDefinitionNotFound().WithPayload(&model.ErrorResponse{
					Errors: []*model.Error{&JobNotFound},
				})
			}
			if err != nil {
				return southbound.NewPutJobsIDDefinitionDefault(http.StatusInternalServerError)
			}
			return southbound.NewPutJobsIDDefinitionOK().WithPayload(definition)
		})

	serverAPI.SouthboundGetWorkflowsNameHandler = southbound.GetWorkflowsNameHandlerFunc(
		func(params southbound.GetWorkflowsNameParams) middleware.Responder {
			workflow, err := workflow.GetWorkflow(params.HTTPRequest.Context(), storage, params.Name)
			if ftag.Get(err) == ftag.NotFound {
				return southbound.NewGetWorkflowsNameNotFound().WithPayload(
					&model.ErrorResponse{
						Errors: []*model.Error{&WorkflowNotFound},
					})
			}
			if err != nil {
				return southbound.NewGetWorkflowsNameDefault(http.StatusInternalServerError)
			}
			return southbound.NewGetWorkflowsNameOK().WithPayload(workflow)
		})

	serverAPI.SouthboundGetWorkflowsHandler = southbound.GetWorkflowsHandlerFunc(
		func(params southbound.GetWorkflowsParams) middleware.Responder {
			pagination := persistence.PaginationParams{Offset: *params.Offset, Limit: *params.Limit}
			ctx := params.HTTPRequest.Context()
			log := logging.LoggerFromCtx(ctx)
			workflows, err := workflow.QueryWorkflows(ctx, storage, pagination)
			if err != nil {
				log.Error().Err(err).Msg("Failed to query workflows")
				return southbound.NewGetWorkflowsDefault(http.StatusInternalServerError)
			}
			return southbound.NewGetWorkflowsOK().WithPayload(workflows)
		})

	serverAPI.SouthboundGetJobsIDTagsHandler = southbound.GetJobsIDTagsHandlerFunc(
		func(params southbound.GetJobsIDTagsParams) middleware.Responder {
			tags, err := tags.Get(params.HTTPRequest.Context(), storage, params.ID)
			if ftag.Get(err) == ftag.NotFound {
				return southbound.NewGetJobsIDTagsNotFound().WithPayload(&model.ErrorResponse{
					Errors: []*model.Error{&JobNotFound},
				})
			}
			if err != nil {
				return southbound.NewGetJobsIDTagsDefault(http.StatusInternalServerError)
			}
			return southbound.NewGetJobsIDTagsOK().WithPayload(tags)
		})

	return serverAPI, nil
}
