// SPDX-FileCopyrightText: 2025 Siemens AG
//
// SPDX-License-Identifier: Apache-2.0
//
// Author: Michael Adler <michael.adler@siemens.com>
import gleam/list
import gleam/option.{type Option, None, Some}
import gleam/result
import rsvp

import events
import wfx
import wfx/utils

pub const default_limit = 10

pub type Model {
  Model(
    wfx_url: String,
    base_path: String,
    route: Route,
    paginated_jobs: Option(PaginatedJobs),
    paginated_workflows: Option(PaginatedWorkflows),
    job: Option(Job),
    workflow: Option(Workflow),
    event_source: Option(events.JobsEventSource),
  )
}

pub type ApiError {
  RsvpError(rsvp.Error)
}

pub type Route {
  RouteJobs(page: Int, limit: Int)
  RouteWorkflows(page: Int, limit: Int)
  RouteWorkflowDetails(name: String)
  RouteJobDetails(id: String)
}

pub type PaginatedJobs {
  PaginatedJobs(response: Result(wfx.PaginatedJobList, ApiError))
}

pub type Job {
  Job(
    response: Result(wfx.Job, ApiError),
    mermaid_diagram: Option(String),
    copied_to_clipboard: Bool,
  )
}

pub type PaginatedWorkflows {
  PaginatedWorkflows(response: Result(wfx.PaginatedWorkflowList, ApiError))
}

pub type Workflow {
  Workflow(
    response: Result(wfx.Workflow, ApiError),
    mermaid_diagram: Option(String),
    copied_to_clipboard: Bool,
  )
}

pub fn new(wfx_url wfx_url: String, base_path base_path: String) -> Model {
  Model(
    wfx_url: wfx_url,
    base_path: base_path,
    route: RouteJobs(page: 1, limit: default_limit),
    paginated_jobs: None,
    paginated_workflows: None,
    job: None,
    workflow: None,
    event_source: None,
  )
}

pub fn extract_workflow(model: Model, name: String) -> Option(wfx.Workflow) {
  let from_paginated_workflows = fn() {
    case model.paginated_workflows {
      Some(paginated) ->
        paginated.response
        |> result.map(fn(x) {
          list.find(x.content, fn(wf) { wf.name == name }) |> option.from_result
        })
        |> option.from_result
        |> option.flatten
      _ -> None
    }
  }

  let from_paginated_jobs = fn() {
    case model.paginated_jobs {
      Some(paginated) ->
        paginated.response
        |> result.map(fn(x) {
          list.find(x.content, fn(job) { job.workflow.name == name })
          |> option.from_result
        })
        |> option.from_result
        |> option.flatten
        |> option.map(fn(job) { job.workflow })
      _ -> None
    }
  }

  let from_workflow = fn() {
    case model.workflow |> option.map(fn(x) { x.response }) {
      Some(Ok(workflow)) ->
        case workflow.name == name {
          True -> Some(workflow)
          False -> None
        }
      _ -> None
    }
  }

  let from_job = fn() {
    case model.job |> option.map(fn(x) { x.response }) {
      Some(Ok(job)) ->
        case job.workflow.name == name {
          True -> Some(job.workflow)
          False -> None
        }
      _ -> None
    }
  }

  // try to find workflow from the given data, in this order
  list.find_map(
    [
      from_paginated_workflows,
      from_paginated_jobs,
      from_workflow,
      from_job,
    ],
    fn(f) {
      case f() {
        None -> Error(Nil)
        Some(wf) -> Ok(wf)
      }
    },
  )
  |> option.from_result
}

pub fn merge_job_event(model: Model, event: wfx.JobEvent) -> Model {
  let model = case model.paginated_jobs {
    Some(paginated) ->
      case paginated.response {
        Error(_) -> model
        Ok(paginated) ->
          Model(
            ..model,
            paginated_jobs: Some(
              PaginatedJobs(
                response: Ok(merge_job_event_into_jobs(paginated, event)),
              ),
            ),
          )
      }
    _ -> model
  }

  let model = case model.job {
    Some(job_model) -> {
      let new_job =
        job_model.response |> result.map(merge_single_event(_, event))
      Model(..model, job: Some(Job(..job_model, response: new_job)))
    }
    _ -> model
  }

  model
}

fn merge_job_event_into_jobs(
  jobs: wfx.PaginatedJobList,
  ev: wfx.JobEvent,
) -> wfx.PaginatedJobList {
  let pagination = case ev.action {
    wfx.ActionCreate ->
      wfx.Pagination(
        limit: jobs.pagination.limit,
        offset: jobs.pagination.offset,
        total: jobs.pagination.total + 1,
      )
    wfx.ActionDelete ->
      wfx.Pagination(
        limit: jobs.pagination.limit,
        offset: jobs.pagination.offset,
        total: jobs.pagination.total - 1,
      )
    _ -> jobs.pagination
  }

  let content = case ev.action {
    wfx.ActionCreate -> {
      // if we are on the first page, prepend the new job to the top
      case utils.page_from_pagination(jobs.pagination) == 1 {
        True -> {
          let other_jobs = jobs.content |> list.take(jobs.pagination.limit)
          [ev.job, ..other_jobs]
        }
        False ->
          // ignore new job
          jobs.content
      }
    }
    wfx.ActionDelete ->
      jobs.content |> list.filter(fn(job) { job.id != ev.job.id })
    _ ->
      jobs.content
      |> list.map(fn(job) { merge_single_event(job, ev) })
  }

  wfx.PaginatedJobList(pagination: pagination, content: content)
}

fn merge_single_event(job: wfx.Job, ev: wfx.JobEvent) -> wfx.Job {
  case job.id == ev.job.id, ev.action {
    True, wfx.ActionAddTags ->
      wfx.Job(..job, mtime: ev.job.mtime, tags: ev.job.tags)
    True, wfx.ActionDeleteTags ->
      wfx.Job(..job, mtime: ev.job.mtime, tags: ev.job.tags)
    True, wfx.ActionUpdateDefinition ->
      wfx.Job(..job, mtime: ev.job.mtime, definition: ev.job.definition)
    True, wfx.ActionUpdateStatus ->
      wfx.Job(..job, mtime: ev.job.mtime, status: ev.job.status)
    // nothing to do in these cases:
    _, _ -> job
  }
}
