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
	"github.com/siemens/wfx/generated/northbound/restapi"
	"github.com/siemens/wfx/generated/northbound/restapi/operations"
	"github.com/siemens/wfx/generated/northbound/restapi/operations/northbound"
	"github.com/siemens/wfx/internal/handler/job"
	"github.com/siemens/wfx/internal/handler/job/definition"
	"github.com/siemens/wfx/internal/handler/job/status"
	"github.com/siemens/wfx/internal/handler/job/tags"
	"github.com/siemens/wfx/internal/handler/workflow"
	"github.com/siemens/wfx/middleware/logging"
	"github.com/siemens/wfx/persistence"
)

func NewNorthboundAPI(storage persistence.Storage) (*operations.WorkflowExecutorAPI, error) {
	swaggerSpec, err := loads.Embedded(restapi.SwaggerJSON, restapi.FlatSwaggerJSON)
	if err != nil {
		return nil, fault.Wrap(err)
	}
	serverAPI := operations.NewWorkflowExecutorAPI(swaggerSpec)

	serverAPI.NorthboundGetJobsHandler = northbound.GetJobsHandlerFunc(
		func(params northbound.GetJobsParams) middleware.Responder {
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
				return northbound.NewGetJobsDefault(http.StatusInternalServerError)
			}
			return northbound.NewGetJobsOK().WithPayload(jobs)
		})

	serverAPI.NorthboundGetJobsIDHandler = northbound.GetJobsIDHandlerFunc(
		func(params northbound.GetJobsIDParams) middleware.Responder {
			history := false
			if params.History != nil {
				history = *params.History
			}
			job, err := job.GetJob(params.HTTPRequest.Context(), storage, params.ID, history)
			if ftag.Get(err) == ftag.NotFound {
				return northbound.NewGetJobsIDNotFound().WithPayload(&model.ErrorResponse{
					Errors: []*model.Error{&JobNotFound},
				})
			}
			if err != nil {
				return northbound.NewGetJobsIDDefault(http.StatusInternalServerError)
			}
			return northbound.NewGetJobsIDOK().WithPayload(job)
		})

	serverAPI.NorthboundGetJobsIDStatusHandler = northbound.GetJobsIDStatusHandlerFunc(
		func(params northbound.GetJobsIDStatusParams) middleware.Responder {
			status, err := status.Get(params.HTTPRequest.Context(), storage, params.ID)
			if ftag.Get(err) == ftag.NotFound {
				return northbound.NewGetJobsIDStatusNotFound().WithPayload(&model.ErrorResponse{
					Errors: []*model.Error{&JobNotFound},
				})
			}
			if err != nil {
				return northbound.NewGetJobsIDStatusDefault(http.StatusInternalServerError)
			}
			return northbound.NewGetJobsIDStatusOK().WithPayload(status)
		})

	serverAPI.NorthboundPutJobsIDStatusHandler = northbound.PutJobsIDStatusHandlerFunc(
		func(params northbound.PutJobsIDStatusParams) middleware.Responder {
			status, err := status.Update(params.HTTPRequest.Context(), storage, params.ID, params.NewJobStatus, model.EligibleEnumWFX)
			if err != nil {
				switch ftag.Get(err) {
				case ftag.NotFound:
					return northbound.NewPutJobsIDStatusNotFound().WithPayload(&model.ErrorResponse{
						Errors: []*model.Error{&JobNotFound},
					})
				case ftag.InvalidArgument:
					err2 := InvalidRequest
					err2.Message = err.Error()
					return northbound.NewPutJobsIDStatusBadRequest().WithPayload(&model.ErrorResponse{Errors: []*model.Error{&err2}})
				default:
					return northbound.NewPutJobsIDStatusDefault(http.StatusInternalServerError)
				}
			}
			return northbound.NewPutJobsIDStatusOK().WithPayload(status)
		})

	serverAPI.NorthboundGetJobsIDDefinitionHandler = northbound.GetJobsIDDefinitionHandlerFunc(
		func(params northbound.GetJobsIDDefinitionParams) middleware.Responder {
			definition, err := definition.Get(params.HTTPRequest.Context(), storage, params.ID)
			if ftag.Get(err) == ftag.NotFound {
				return northbound.NewGetJobsIDDefinitionNotFound().WithPayload(&model.ErrorResponse{
					Errors: []*model.Error{&JobNotFound},
				})
			}
			if err != nil {
				return northbound.NewGetJobsIDDefinitionDefault(http.StatusInternalServerError)
			}
			return northbound.NewGetJobsIDDefinitionOK().WithPayload(definition)
		})
	serverAPI.NorthboundPutJobsIDDefinitionHandler = northbound.PutJobsIDDefinitionHandlerFunc(
		func(params northbound.PutJobsIDDefinitionParams) middleware.Responder {
			definition, err := definition.Update(params.HTTPRequest.Context(), storage, params.ID, params.JobDefinition)
			if ftag.Get(err) == ftag.NotFound {
				return northbound.NewPutJobsIDDefinitionNotFound().WithPayload(&model.ErrorResponse{
					Errors: []*model.Error{&JobNotFound},
				})
			}
			if err != nil {
				return northbound.NewPutJobsIDDefinitionDefault(http.StatusInternalServerError)
			}

			return northbound.NewPutJobsIDDefinitionOK().WithPayload(definition)
		})

	serverAPI.NorthboundGetWorkflowsNameHandler = northbound.GetWorkflowsNameHandlerFunc(
		func(params northbound.GetWorkflowsNameParams) middleware.Responder {
			workflow, err := workflow.GetWorkflow(params.HTTPRequest.Context(), storage, params.Name)
			if ftag.Get(err) == ftag.NotFound {
				return northbound.NewGetWorkflowsNameNotFound().WithPayload(
					&model.ErrorResponse{
						Errors: []*model.Error{&WorkflowNotFound},
					})
			}
			if err != nil {
				return northbound.NewGetWorkflowsNameDefault(http.StatusInternalServerError)
			}
			return northbound.NewGetWorkflowsNameOK().WithPayload(workflow)
		})

	serverAPI.NorthboundGetWorkflowsHandler = northbound.GetWorkflowsHandlerFunc(
		func(params northbound.GetWorkflowsParams) middleware.Responder {
			pagination := persistence.PaginationParams{Offset: *params.Offset, Limit: *params.Limit}
			ctx := params.HTTPRequest.Context()
			log := logging.LoggerFromCtx(ctx)
			workflows, err := workflow.QueryWorkflows(ctx, storage, pagination)
			if err != nil {
				log.Error().Err(err).Msg("Failed to query workflows")
				return northbound.NewGetWorkflowsDefault(http.StatusInternalServerError)
			}
			return northbound.NewGetWorkflowsOK().WithPayload(workflows)
		})

	// not available southbound
	serverAPI.NorthboundDeleteWorkflowsNameHandler = northbound.DeleteWorkflowsNameHandlerFunc(
		func(params northbound.DeleteWorkflowsNameParams) middleware.Responder {
			err := workflow.DeleteWorkflow(params.HTTPRequest.Context(), storage, params.Name)
			if err != nil {
				switch ftag.Get(err) {
				case ftag.NotFound:
					return northbound.NewDeleteWorkflowsNameNotFound()
				default:
					return northbound.NewDeleteWorkflowsNameDefault(http.StatusInternalServerError)
				}
			}
			return northbound.NewDeleteWorkflowsNameNoContent()
		})
	serverAPI.NorthboundPostWorkflowsHandler = northbound.PostWorkflowsHandlerFunc(
		func(params northbound.PostWorkflowsParams) middleware.Responder {
			wf, err := workflow.CreateWorkflow(params.HTTPRequest.Context(), storage, params.Workflow)
			if err != nil {
				switch ftag.Get(err) {
				case ftag.InvalidArgument:
					err2 := WorkflowInvalid
					err2.Message = err.Error()
					return northbound.NewPostWorkflowsBadRequest().WithPayload(&model.ErrorResponse{Errors: []*model.Error{&err2}})
				case ftag.AlreadyExists:
					return northbound.NewPostWorkflowsBadRequest().WithPayload(&model.ErrorResponse{Errors: []*model.Error{&WorkflowNotUnique}})
				default:
					return northbound.NewPostWorkflowsDefault(http.StatusInternalServerError)
				}
			}
			return northbound.NewPostWorkflowsCreated().WithPayload(wf)
		})
	serverAPI.NorthboundPostJobsHandler = northbound.PostJobsHandlerFunc(
		func(params northbound.PostJobsParams) middleware.Responder {
			job, err := job.CreateJob(params.HTTPRequest.Context(), storage, params.Job)
			if err != nil {
				switch ftag.Get(err) {
				case ftag.NotFound:
					err2 := WorkflowNotFound
					err2.Message = err.Error()
					return northbound.NewPostJobsBadRequest().WithPayload(&model.ErrorResponse{Errors: []*model.Error{&err2}})
				default:
					return northbound.NewPostJobsDefault(http.StatusInternalServerError)
				}
			}
			return northbound.NewPostJobsCreated().WithPayload(job)
		})
	serverAPI.NorthboundDeleteJobsIDHandler = northbound.DeleteJobsIDHandlerFunc(
		func(params northbound.DeleteJobsIDParams) middleware.Responder {
			err := job.DeleteJob(params.HTTPRequest.Context(), storage, params.ID)
			if err != nil {
				switch ftag.Get(err) {
				case ftag.NotFound:
					return northbound.NewDeleteJobsIDNotFound()
				default:
					return northbound.NewDeleteJobsIDDefault(http.StatusInternalServerError)
				}
			}
			return northbound.NewDeleteJobsIDNoContent()
		})

	// tags
	serverAPI.NorthboundGetJobsIDTagsHandler = northbound.GetJobsIDTagsHandlerFunc(
		func(params northbound.GetJobsIDTagsParams) middleware.Responder {
			tags, err := tags.Get(params.HTTPRequest.Context(), storage, params.ID)
			if ftag.Get(err) == ftag.NotFound {
				return northbound.NewGetJobsIDTagsNotFound().WithPayload(&model.ErrorResponse{
					Errors: []*model.Error{&JobNotFound},
				})
			}
			if err != nil {
				return northbound.NewGetJobsIDTagsDefault(http.StatusInternalServerError)
			}
			return northbound.NewGetJobsIDTagsOK().WithPayload(tags)
		})
	serverAPI.NorthboundPostJobsIDTagsHandler = northbound.PostJobsIDTagsHandlerFunc(
		func(params northbound.PostJobsIDTagsParams) middleware.Responder {
			tags, err := tags.Add(params.HTTPRequest.Context(), storage, params.ID, params.Tags)
			if ftag.Get(err) == ftag.NotFound {
				return northbound.NewPostJobsIDTagsNotFound().WithPayload(&model.ErrorResponse{
					Errors: []*model.Error{&JobNotFound},
				})
			}
			if err != nil {
				return northbound.NewPostJobsIDTagsDefault(http.StatusInternalServerError)
			}
			return northbound.NewPostJobsIDTagsOK().WithPayload(tags)
		})
	serverAPI.NorthboundDeleteJobsIDTagsHandler = northbound.DeleteJobsIDTagsHandlerFunc(
		func(params northbound.DeleteJobsIDTagsParams) middleware.Responder {
			tags, err := tags.Delete(params.HTTPRequest.Context(), storage, params.ID, params.Tags)
			if ftag.Get(err) == ftag.NotFound {
				return northbound.NewDeleteJobsIDTagsNotFound().WithPayload(&model.ErrorResponse{Errors: []*model.Error{&JobNotFound}})
			}
			if err != nil {
				return northbound.NewDeleteJobsIDTagsDefault(http.StatusInternalServerError)
			}
			return northbound.NewDeleteJobsIDTagsOK().WithPayload(tags)
		})

	return serverAPI, nil
}
