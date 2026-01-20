// SPDX-FileCopyrightText: 2025 Siemens AG
//
// SPDX-License-Identifier: Apache-2.0
//
// Author: Michael Adler <michael.adler@siemens.com>
import birdie
import config
import gleam/dict
import gleam/option.{None, Some}
import gleam/time/timestamp
import gleeunit
import lustre/dev/simulate
import lustre/element
import model
import msg
import wfx

import app
import view

const cfg = config.Config(wfx_url: "http://localhost", base_path: "/ui")

pub fn main() -> Nil {
  gleeunit.main()
}

pub fn start_empty_test() {
  simulate.application(fn(_) { app.init(cfg) }, app.update, view.view)
  |> simulate.start(Nil)
  |> simulate.view
  |> element.to_readable_string
  |> birdie.snap("Empty Paginated Jobs")
}

pub fn given_empty_app_when_wfx_sends_jobs_then_jobs_displayed_test() {
  let example_job =
    wfx.Job(
      id: "1",
      client_id: "john_doe",
      workflow: wfx.Workflow(
        name: "workflow.test",
        description: None,
        states: [wfx.State(name: "NEW", description: None)],
        groups: Some([
          wfx.Group(name: "OPEN", description: None, states: ["NEW"]),
        ]),
        transitions: [],
      ),
      tags: [],
      definition: dict.new(),
      status: Some(wfx.JobStatus(
        state: "NEW",
        client_id: None,
        progress: None,
        message: None,
        definition_hash: None,
        context: None,
      )),
      stime: Some(timestamp.from_unix_seconds(0)),
      mtime: Some(timestamp.from_unix_seconds(0)),
      history: None,
    )

  simulate.application(fn(_) { app.init(cfg) }, app.update, view.view)
  |> simulate.start(Nil)
  |> simulate.message(
    msg.WfxSentJobs(
      Ok(
        wfx.PaginatedJobList(
          pagination: wfx.Pagination(limit: 10, offset: 0, total: 1),
          content: [example_job],
        ),
      ),
    ),
  )
  |> simulate.view
  |> element.to_readable_string
  |> birdie.snap("Paginated Jobs")
}

pub fn given_jobs_when_wfx_sends_job_event_then_jobs_are_updated_test() {
  let example_job =
    wfx.Job(
      id: "1",
      client_id: "john_doe",
      workflow: wfx.Workflow(
        name: "workflow.test",
        description: None,
        states: [wfx.State(name: "NEW", description: None)],
        groups: Some([
          wfx.Group(name: "OPEN", description: None, states: ["NEW"]),
        ]),
        transitions: [],
      ),
      tags: [],
      definition: dict.new(),
      status: Some(wfx.JobStatus(
        state: "NEW",
        client_id: None,
        progress: None,
        message: None,
        definition_hash: None,
        context: None,
      )),
      stime: Some(timestamp.from_unix_seconds(0)),
      mtime: Some(timestamp.from_unix_seconds(0)),
      history: None,
    )

  simulate.application(fn(_) { app.init(cfg) }, app.update, view.view)
  |> simulate.start(Nil)
  |> simulate.message(
    msg.WfxSentJobs(
      Ok(
        wfx.PaginatedJobList(
          pagination: wfx.Pagination(limit: 10, offset: 0, total: 1),
          content: [example_job],
        ),
      ),
    ),
  )
  |> simulate.message(msg.WfxSentJobEvent(
    "{\"ctime\": \"2025-10-01T14:39:05.155+02:00\", \"action\":\"UPDATE_STATUS\",\"job\":{\"id\":\"1\", \"clientId\": \"Dana\", \"workflow\": { \"name\": \"wfx.workflow.kanban\"}, \"status\":{\"state\":\"PROGRESS\"}}}",
  ))
  |> simulate.view
  |> element.to_readable_string
  |> birdie.snap("Paginated Jobs After Job Event")
}

pub fn given_app_when_view_job_then_job_is_shown_test() {
  let example_job =
    wfx.Job(
      id: "1",
      client_id: "john_doe",
      workflow: wfx.Workflow(
        name: "workflow.test",
        description: None,
        states: [wfx.State(name: "NEW", description: None)],
        groups: Some([
          wfx.Group(name: "OPEN", description: None, states: ["NEW"]),
        ]),
        transitions: [],
      ),
      tags: [],
      definition: dict.new(),
      status: Some(wfx.JobStatus(
        state: "NEW",
        client_id: None,
        progress: None,
        message: None,
        definition_hash: None,
        context: None,
      )),
      stime: Some(timestamp.from_unix_seconds(0)),
      mtime: Some(timestamp.from_unix_seconds(0)),
      history: None,
    )

  simulate.application(fn(_) { app.init(cfg) }, app.update, view.view)
  |> simulate.start(Nil)
  |> simulate.message(
    msg.DocumentChangedRoute(model.RouteJobDetails(example_job.id)),
  )
  |> simulate.message(msg.WfxSentSingleJob(Ok(example_job)))
  |> simulate.view
  |> element.to_readable_string
  |> birdie.snap("Job Details")
}

pub fn given_app_when_view_workflows_then_workflow_table_shown_test() {
  let example_workflow =
    wfx.Workflow(
      name: "workflow.test",
      description: None,
      states: [wfx.State(name: "NEW", description: None)],
      groups: Some([
        wfx.Group(name: "OPEN", description: None, states: ["NEW"]),
      ]),
      transitions: [],
    )

  simulate.application(fn(_) { app.init(cfg) }, app.update, view.view)
  |> simulate.start(Nil)
  |> simulate.message(
    msg.DocumentChangedRoute(model.RouteWorkflows(
      page: 1,
      limit: model.default_limit,
    )),
  )
  |> simulate.message(
    msg.WfxSentWorkflows(
      Ok(
        wfx.PaginatedWorkflowList(
          pagination: wfx.Pagination(limit: 10, offset: 0, total: 1),
          content: [example_workflow],
        ),
      ),
    ),
  )
  |> simulate.view
  |> element.to_readable_string
  |> birdie.snap("Paginated Workflows")
}

pub fn given_app_when_view_workflow_then_job_is_shown_test() {
  let example_workflow =
    wfx.Workflow(
      name: "workflow.test",
      description: None,
      states: [wfx.State(name: "NEW", description: None)],
      groups: Some([
        wfx.Group(name: "OPEN", description: None, states: ["NEW"]),
      ]),
      transitions: [],
    )

  simulate.application(fn(_) { app.init(cfg) }, app.update, view.view)
  |> simulate.start(Nil)
  |> simulate.message(
    msg.DocumentChangedRoute(model.RouteWorkflowDetails(example_workflow.name)),
  )
  |> simulate.message(msg.WfxSentSingleWorkflow(Ok(example_workflow)))
  |> simulate.view
  |> element.to_readable_string
  |> birdie.snap("Workflow Details")
}
