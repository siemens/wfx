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
	"fmt"
	"strings"
	"time"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/ftag"
	"github.com/alexliesenfeld/health"
	"github.com/rs/zerolog/log"
	"github.com/siemens/wfx/cmd/wfx/cmd/config"
	"github.com/siemens/wfx/cmd/wfx/metadata"
	"github.com/siemens/wfx/generated/api"
	"github.com/siemens/wfx/internal/handler/job"
	"github.com/siemens/wfx/internal/handler/job/definition"
	"github.com/siemens/wfx/internal/handler/job/events"
	"github.com/siemens/wfx/internal/handler/job/status"
	"github.com/siemens/wfx/internal/handler/job/tags"
	"github.com/siemens/wfx/internal/handler/workflow"
	"github.com/siemens/wfx/middleware/logging"
	"github.com/siemens/wfx/middleware/sse"
	"github.com/siemens/wfx/persistence"
)

const (
	defaultPageLimit = 10
)

type WfxServer struct {
	storage persistence.Storage
	checker health.Checker
	sseOpts SSEOpts
}

type SSEOpts struct {
	PingInterval  time.Duration
	GraceInterval time.Duration
}

func NewWfxServer(storage persistence.Storage) *WfxServer {
	checker := health.NewChecker(
		health.WithTimeout(10*time.Second),
		health.WithPeriodicCheck(30*time.Second, 0, health.Check{
			Name: "persistence",
			Check: func(ctx context.Context) error {
				return fault.Wrap(storage.CheckHealth(ctx))
			},
		}),
		health.WithStatusListener(healthStatusListener),
		health.WithDisabledAutostart())
	wfx := &WfxServer{
		storage: storage,
		checker: checker,
		sseOpts: SSEOpts{
			PingInterval:  config.DefaultSSEPingInterval,
			GraceInterval: config.DefaultSSEGraceInterval,
		},
	}
	return wfx
}

func (server *WfxServer) WithSSEOpts(sseOpts SSEOpts) *WfxServer {
	server.sseOpts = sseOpts
	return server
}

func (server WfxServer) Start() {
	server.checker.Start()
}

func (server WfxServer) Stop() {
	if server.checker.IsStarted() {
		server.checker.Stop()
	}
}

func healthStatusListener(_ context.Context, state health.CheckerState) {
	logFn := log.Warn
	switch state.Status {
	case health.StatusDown:
		logFn = log.Error
	case health.StatusUp:
		logFn = log.Info
	case health.StatusUnknown:
		logFn = log.Warn
	}

	childLog := logFn()
	for k, v := range state.CheckState {
		childLog.Str(k, string(v.Status))
	}

	childLog.Str("overall", string(state.Status)).Msg("Health status changed")
}

//revive:disable:var-naming
func (server WfxServer) GetJobs(ctx context.Context, request api.GetJobsRequestObject) (api.GetJobsResponseObject, error) {
	filter := persistence.FilterParams{
		ClientID: request.Params.ParamClientID,
		State:    request.Params.ParamState,
		Workflow: request.Params.ParamWorkflow,
	}
	if request.Params.ParamGroup != nil {
		filter.Group = *request.Params.ParamGroup
	}
	if request.Params.ParamTag != nil {
		filter.Tags = *request.Params.ParamTag
	}

	pagination := persistence.PaginationParams{Offset: 0, Limit: defaultPageLimit}
	if request.Params.ParamOffset != nil {
		pagination.Offset = *request.Params.ParamOffset
	}
	if request.Params.ParamLimit != nil {
		pagination.Limit = *request.Params.ParamLimit
	}

	jobs, err := job.QueryJobs(ctx, server.storage, filter, pagination, (*string)(request.Params.ParamSort))
	if err != nil {
		return nil, fault.Wrap(err)
	}
	if request.Params.XResponseFilter != nil {
		return NewJQFilter(*request.Params.XResponseFilter, *jobs), nil
	}
	return api.GetJobs200JSONResponse(*jobs), nil
}

func (server WfxServer) PostJobs(ctx context.Context, request api.PostJobsRequestObject) (api.PostJobsResponseObject, error) {
	job, err := job.CreateJob(ctx, server.storage, request.Body)
	if err != nil {
		switch ftag.Get(err) {
		case ftag.NotFound:
			err2 := WorkflowNotFound
			err2.Message = err.Error()
			return api.PostJobs400JSONResponse(api.ErrorResponse{
				Errors: &[]api.Error{err2},
			}), nil
		default:
			return nil, fault.Wrap(err)
		}
	}
	if request.Params.XResponseFilter != nil {
		return NewJQFilter(*request.Params.XResponseFilter, *job), nil
	}
	return api.PostJobs201JSONResponse(*job), nil
}

func (server WfxServer) GetJobsEvents(ctx context.Context, request api.GetJobsEventsRequestObject) (api.GetJobsEventsResponseObject, error) {
	var filter events.FilterParams
	if ids := request.Params.JobIds; ids != nil {
		filter.JobIDs = strings.Split(*ids, ",")
	}
	if ids := request.Params.ClientIDs; ids != nil {
		filter.ClientIDs = strings.Split(*ids, ",")
	}
	if wfs := request.Params.Workflows; wfs != nil {
		filter.Workflows = strings.Split(*wfs, ",")
	}

	var tags []string
	if s := request.Params.Tags; s != nil {
		tags = strings.Split(*s, ",")
	}
	subscriber := events.AddSubscriber(ctx, server.sseOpts.GraceInterval, filter, tags)
	return sse.NewResponder(ctx, server.sseOpts.PingInterval, subscriber), nil
}

func (server WfxServer) DeleteJobsId(ctx context.Context, request api.DeleteJobsIdRequestObject) (api.DeleteJobsIdResponseObject, error) {
	if err := job.DeleteJob(ctx, server.storage, request.Id); err != nil {
		if ftag.Get(err) == ftag.NotFound {
			return api.DeleteJobsId404JSONResponse(api.ErrorResponse{}), nil
		}
		return nil, fault.Wrap(err)
	}
	return api.DeleteJobsId204Response{}, nil
}

func (server WfxServer) GetJobsId(ctx context.Context, request api.GetJobsIdRequestObject) (api.GetJobsIdResponseObject, error) {
	history := false
	if request.Params.ParamHistory != nil {
		history = *request.Params.ParamHistory
	}
	job, err := job.GetJob(ctx, server.storage, request.Id, history)
	if err != nil {
		if ftag.Get(err) == ftag.NotFound {
			return api.GetJobsId404JSONResponse(api.ErrorResponse{
				Errors: &[]api.Error{JobNotFound},
			}), nil
		}
		return nil, fault.Wrap(err)
	}
	if request.Params.XResponseFilter != nil {
		return NewJQFilter(*request.Params.XResponseFilter, *job), nil
	}
	return api.GetJobsId200JSONResponse(*job), nil
}

func (server WfxServer) GetJobsIdDefinition(ctx context.Context, request api.GetJobsIdDefinitionRequestObject) (api.GetJobsIdDefinitionResponseObject, error) {
	definition, err := definition.Get(ctx, server.storage, request.Id)
	if err != nil {
		if ftag.Get(err) == ftag.NotFound {
			return api.GetJobsIdDefinition404JSONResponse(api.ErrorResponse{
				Errors: &[]api.Error{JobNotFound},
			}), nil
		}
		return nil, fault.Wrap(err)
	}
	if request.Params.XResponseFilter != nil {
		return NewJQFilter(*request.Params.XResponseFilter, definition), nil
	}
	return api.GetJobsIdDefinition200JSONResponse(definition), nil
}

func (server WfxServer) PutJobsIdDefinition(ctx context.Context, request api.PutJobsIdDefinitionRequestObject) (api.PutJobsIdDefinitionResponseObject, error) {
	var def map[string]any
	if request.Body != nil {
		def = *request.Body
	}
	definition, err := definition.Update(ctx, server.storage, request.Id, def)
	if err != nil {
		if ftag.Get(err) == ftag.NotFound {
			return api.PutJobsIdDefinition404JSONResponse(api.ErrorResponse{
				Errors: &[]api.Error{JobNotFound},
			}), nil
		}
		return nil, fault.Wrap(err)
	}
	if request.Params.XResponseFilter != nil {
		return NewJQFilter(*request.Params.XResponseFilter, definition), nil
	}
	return api.PutJobsIdDefinition200JSONResponse(definition), nil
}

func (server WfxServer) GetJobsIdStatus(ctx context.Context, request api.GetJobsIdStatusRequestObject) (api.GetJobsIdStatusResponseObject, error) {
	status, err := status.Get(ctx, server.storage, request.Id)
	if err != nil {
		if ftag.Get(err) == ftag.NotFound {
			return api.GetJobsIdStatus404JSONResponse(api.ErrorResponse{
				Errors: &[]api.Error{JobNotFound},
			}), nil
		}
		return nil, fault.Wrap(err)
	}
	if request.Params.XResponseFilter != nil {
		return NewJQFilter(*request.Params.XResponseFilter, *status), nil
	}
	return api.GetJobsIdStatus200JSONResponse(*status), nil
}

func (server WfxServer) PutJobsIdStatus(ctx context.Context, request api.PutJobsIdStatusRequestObject, eligible api.EligibleEnum) (api.PutJobsIdStatusResponseObject, error) {
	status, err := status.Update(ctx, server.storage, request.Id, request.Body, eligible)
	if err != nil {
		switch ftag.Get(err) {
		case ftag.NotFound:
			return api.PutJobsIdStatus404JSONResponse(api.ErrorResponse{
				Errors: &[]api.Error{JobNotFound},
			}), nil
		case ftag.InvalidArgument:
			err2 := InvalidRequest
			err2.Message = err.Error()
			return api.PutJobsIdStatus400JSONResponse(api.ErrorResponse{
				Errors: &[]api.Error{err2},
			}), nil
		default:
			return nil, fault.Wrap(err)
		}
	}
	if request.Params.XResponseFilter != nil {
		return NewJQFilter(*request.Params.XResponseFilter, *status), nil
	}
	return api.PutJobsIdStatus200JSONResponse(*status), nil
}

func (server WfxServer) DeleteJobsIdTags(ctx context.Context, request api.DeleteJobsIdTagsRequestObject) (api.DeleteJobsIdTagsResponseObject, error) {
	var body []string
	if request.Body != nil {
		body = *request.Body
	}
	tags, err := tags.Delete(ctx, server.storage, request.Id, body)
	if err != nil {
		if ftag.Get(err) == ftag.NotFound {
			return api.DeleteJobsIdTags404JSONResponse(api.ErrorResponse{
				Errors: &[]api.Error{JobNotFound},
			}), nil
		}
		return nil, fault.Wrap(err)
	}
	if request.Params.XResponseFilter != nil {
		return NewJQFilter(*request.Params.XResponseFilter, tags), nil
	}
	return api.DeleteJobsIdTags200JSONResponse(tags), nil
}

func (server WfxServer) GetJobsIdTags(ctx context.Context, request api.GetJobsIdTagsRequestObject) (api.GetJobsIdTagsResponseObject, error) {
	tags, err := tags.Get(ctx, server.storage, request.Id)
	if err != nil {
		if ftag.Get(err) == ftag.NotFound {
			return api.GetJobsIdTags404JSONResponse(api.ErrorResponse{
				Errors: &[]api.Error{JobNotFound},
			}), nil
		}
		return nil, fault.Wrap(err)
	}
	if request.Params.XResponseFilter != nil {
		return NewJQFilter(*request.Params.XResponseFilter, tags), nil
	}
	return api.GetJobsIdTags200JSONResponse(tags), nil
}

func (server WfxServer) PostJobsIdTags(ctx context.Context, request api.PostJobsIdTagsRequestObject) (api.PostJobsIdTagsResponseObject, error) {
	var body []string
	if request.Body != nil {
		body = *request.Body
	}
	tags, err := tags.Add(ctx, server.storage, request.Id, body)
	if err != nil {
		if ftag.Get(err) == ftag.NotFound {
			return api.PostJobsIdTags404JSONResponse(api.ErrorResponse{
				Errors: &[]api.Error{JobNotFound},
			}), nil
		}
		return nil, fault.Wrap(err)
	}
	if request.Params.XResponseFilter != nil {
		return NewJQFilter(*request.Params.XResponseFilter, tags), nil
	}
	return api.PostJobsIdTags200JSONResponse(tags), nil
}

func (server WfxServer) GetWorkflows(ctx context.Context, request api.GetWorkflowsRequestObject) (api.GetWorkflowsResponseObject, error) {
	var offset int64
	if request.Params.ParamOffset != nil {
		offset = *request.Params.ParamOffset
	}
	var limit int32 = defaultPageLimit
	if request.Params.ParamLimit != nil {
		limit = *request.Params.ParamLimit
	}
	pagination := persistence.PaginationParams{Offset: offset, Limit: limit}
	log := logging.LoggerFromCtx(ctx)
	workflows, err := workflow.QueryWorkflows(ctx, server.storage, pagination, (*string)(request.Params.ParamSort))
	if err != nil {
		log.Error().Err(err).Msg("Failed to query workflows")
		return nil, fault.Wrap(err)
	}
	if request.Params.XResponseFilter != nil {
		return NewJQFilter(*request.Params.XResponseFilter, *workflows), nil
	}
	return api.GetWorkflows200JSONResponse(*workflows), nil
}

func (server WfxServer) PostWorkflows(ctx context.Context, request api.PostWorkflowsRequestObject) (api.PostWorkflowsResponseObject, error) {
	wf, err := workflow.CreateWorkflow(ctx, server.storage, request.Body)
	if err != nil {
		switch ftag.Get(err) {
		case ftag.InvalidArgument:
			err2 := WorkflowInvalid
			err2.Message = err.Error()
			return api.PostWorkflows400JSONResponse(api.ErrorResponse{
				Errors: &[]api.Error{err2},
			}), nil
		case ftag.AlreadyExists:
			err2 := WorkflowNotUnique
			err2.Message = fmt.Sprintf("Workflow with name '%s' already exists", request.Body.Name)
			return api.PostWorkflows400JSONResponse(api.ErrorResponse{
				Errors: &[]api.Error{err2},
			}), nil
		default:
			return nil, fault.Wrap(err)
		}
	}
	if request.Params.XResponseFilter != nil {
		return NewJQFilter(*request.Params.XResponseFilter, wf), nil
	}
	return api.PostWorkflows201JSONResponse(*wf), nil
}

func (server WfxServer) DeleteWorkflowsName(ctx context.Context, request api.DeleteWorkflowsNameRequestObject) (api.DeleteWorkflowsNameResponseObject, error) {
	err := workflow.DeleteWorkflow(ctx, server.storage, request.Name)
	if err != nil {
		if ftag.Get(err) == ftag.NotFound {
			err2 := WorkflowNotFound
			err2.Message = fmt.Sprintf("Workflow '%s' not found", request.Name)
			return api.DeleteWorkflowsName404JSONResponse(api.ErrorResponse{
				Errors: &[]api.Error{err2},
			}), nil
		}
		return nil, fault.Wrap(err)
	}
	return api.DeleteWorkflowsName204Response{}, nil
}

func (server WfxServer) GetWorkflowsName(ctx context.Context, request api.GetWorkflowsNameRequestObject) (api.GetWorkflowsNameResponseObject, error) {
	workflow, err := workflow.GetWorkflow(ctx, server.storage, request.Name)
	if err != nil {
		if ftag.Get(err) == ftag.NotFound {
			return api.GetWorkflowsName404JSONResponse(api.ErrorResponse{
				Errors: &[]api.Error{WorkflowNotFound},
			}), nil
		}
		return nil, fault.Wrap(err)
	}
	if request.Params.XResponseFilter != nil {
		return NewJQFilter(*request.Params.XResponseFilter, *workflow), nil
	}
	return api.GetWorkflowsName200JSONResponse(*workflow), nil
}

func (server WfxServer) GetHealth(ctx context.Context, _ api.GetHealthRequestObject) (api.GetHealthResponseObject, error) {
	result := server.checker.Check(ctx)
	details := make(map[string]api.CheckResult, len(result.Details))
	for component, result := range result.Details {
		componentResult := api.CheckResult{
			Status:    (api.AvailabilityStatus)(result.Status),
			Timestamp: result.Timestamp,
		}
		if result.Error != nil {
			componentResult.Error = result.Error.Error()
		}
		details[component] = componentResult
	}

	checkerResult := api.CheckerResult{
		Status: api.AvailabilityStatus(result.Status),
	}
	if len(result.Info) > 0 {
		checkerResult.Info = &result.Info
	}
	if len(result.Details) > 0 {
		checkerResult.Details = &details
	}

	if result.Status == health.StatusUp {
		return api.GetHealth200JSONResponse{
			Headers: api.GetHealth200ResponseHeaders{
				// avoid caching
				CacheControl: "no-cache",
				Pragma:       "no-cache",
				Expires:      "Thu, 01 Jan 1970 00:00:00 GMT",
			},
			Body: checkerResult,
		}, nil
	}
	return api.GetHealth503JSONResponse{
		Headers: api.GetHealth503ResponseHeaders{
			// avoid caching
			CacheControl: "no-cache",
			Pragma:       "no-cache",
			Expires:      "Thu, 01 Jan 1970 00:00:00 GMT",
		},
		Body: checkerResult,
	}, nil
}

func (server WfxServer) GetVersion(context.Context, api.GetVersionRequestObject) (api.GetVersionResponseObject, error) {
	return api.GetVersion200JSONResponse{
		Version:    metadata.Version,
		Commit:     metadata.Commit,
		ApiVersion: metadata.APIVersion,
	}, nil
}
