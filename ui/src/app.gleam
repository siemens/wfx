// SPDX-FileCopyrightText: 2025 Siemens AG
//
// SPDX-License-Identifier: Apache-2.0
//
// Author: Michael Adler <michael.adler@siemens.com>
import events
import gleam/bool
import gleam/int
import gleam/io
import gleam/json
import gleam/list
import gleam/option.{type Option, None, Some}
import gleam/result
import gleam/string
import gleam/string_tree
import gleam/uri.{type Uri}
import lustre
import lustre/effect.{type Effect}
import mermaid
import modem
import plinth/browser/clipboard
import rsvp

import config.{type Config}
import model.{type Model, Model}
import msg.{type Msg}
import view.{view}
import wfx/decoder

pub fn main() {
  let cfg = config.load_config()
  let app = lustre.application(fn(_) { init(cfg) }, update, view)
  let assert Ok(_) = lustre.start(app, "#app", Nil)
  Nil
}

pub fn init(cfg: Config) -> #(Model, Effect(Msg)) {
  let initial_uri = modem.initial_uri()
  let wfx_url = case cfg.wfx_url |> string.starts_with("/") {
    True -> {
      let assert Ok(origin) =
        initial_uri
        |> result.map(uri.origin)
        |> result.flatten
      origin <> cfg.wfx_url
    }
    False -> cfg.wfx_url
  }
  let model = model.new(wfx_url: wfx_url, base_path: cfg.base_path)

  let #(route, initial_eff) =
    initial_uri
    |> result.map(fn(uri) {
      #(
        uri.path_segments(remove_prefix(uri.path, cfg.base_path)),
        uri.query
          |> option.map(uri.parse_query)
          |> option.map(option.from_result)
          |> option.flatten
          |> option.unwrap([]),
      )
    })
    |> fn(path) {
      case path {
        Ok(#(["jobs"], query)) -> {
          let page = int_from_query(query, "page") |> option.unwrap(1)
          let limit =
            int_from_query(query, "limit") |> option.unwrap(model.default_limit)
          #(
            model.RouteJobs(page: page, limit: limit),
            get_jobs(wfx_url: model.wfx_url, page: page, limit: limit),
          )
        }
        Ok(#(["jobs", id], _)) -> #(
          model.RouteJobDetails(id: id),
          get_single_job(wfx_url: model.wfx_url, id: id, history: True),
        )
        Ok(#(["workflows"], query)) -> {
          let page = int_from_query(query, "page") |> option.unwrap(1)
          let limit =
            int_from_query(query, "limit") |> option.unwrap(model.default_limit)
          #(
            model.RouteWorkflows(page: page, limit: limit),
            get_workflows(wfx_url: model.wfx_url, page: page, limit: limit),
          )
        }
        Ok(#(["workflows", name], _)) -> #(
          model.RouteWorkflowDetails(name),
          get_single_workflow(wfx_url: model.wfx_url, name: name),
        )
        // fallback to jobs overview
        _ -> #(
          model.RouteJobs(page: 1, limit: model.default_limit),
          get_jobs(wfx_url: model.wfx_url, page: 1, limit: model.default_limit),
        )
      }
    }

  // subscribe to all job events because we want to be informed about new jobs and updates for those new jobs
  let event_source = events.create_job_events_source(model.wfx_url, None)
  let model = model.Model(..model, route: route, event_source: event_source)

  #(
    model,
    effect.batch([
      modem.init(on_url_change(_, model.base_path)),
      initial_eff,
      event_source |> option.map(get_job_events) |> option.unwrap(effect.none()),
    ]),
  )
}

fn on_url_change(uri: Uri, base_path: String) -> Msg {
  case
    uri.path_segments(remove_prefix(uri.path, base_path)),
    uri.query
    |> option.map(uri.parse_query)
    |> option.map(option.from_result)
    |> option.flatten
    |> option.unwrap([])
  {
    ["jobs"], query ->
      msg.DocumentChangedRoute(model.RouteJobs(
        page: int_from_query(query, "page") |> option.unwrap(1),
        limit: int_from_query(query, "limit")
          |> option.unwrap(model.default_limit),
      ))
    ["jobs", id], _ -> msg.DocumentChangedRoute(model.RouteJobDetails(id: id))
    ["workflows"], query ->
      msg.DocumentChangedRoute(model.RouteWorkflows(
        page: int_from_query(query, "page") |> option.unwrap(1),
        limit: int_from_query(query, "limit")
          |> option.unwrap(model.default_limit),
      ))
    ["workflows", name], _ ->
      msg.DocumentChangedRoute(model.RouteWorkflowDetails(name: name))
    _, _ ->
      // fallback to jobs overview
      msg.DocumentChangedRoute(model.RouteJobs(
        page: 1,
        limit: model.default_limit,
      ))
  }
}

fn int_from_query(query: List(#(String, String)), name: String) -> Option(Int) {
  query
  |> list.find_map(fn(x) {
    let #(key, val) = x
    case key == name {
      True -> Ok(val)
      False -> Error(Nil)
    }
  })
  |> result.try(int.parse)
  |> option.from_result
}

pub fn update(model: Model, msg: Msg) -> #(Model, Effect(Msg)) {
  case msg {
    msg.DocumentChangedRoute(route) -> {
      let model = Model(..model, route: route)
      case route {
        model.RouteJobs(page: page, limit: limit) -> #(
          model,
          get_jobs(wfx_url: model.wfx_url, page: page, limit: limit),
        )
        model.RouteJobDetails(id: id) -> #(
          model,
          get_single_job(wfx_url: model.wfx_url, id: id, history: True),
        )

        model.RouteWorkflows(page: page, limit: limit) -> #(
          model,
          get_workflows(wfx_url: model.wfx_url, page: page, limit: limit),
        )
        model.RouteWorkflowDetails(name) ->
          // check if we already know the workflow, e.g. from a job's properties
          case model.extract_workflow(model, name) {
            Some(workflow) -> #(
              Model(
                ..model,
                workflow: Some(model.Workflow(
                  response: Ok(workflow),
                  mermaid_diagram: None,
                  copied_to_clipboard: False,
                )),
              ),
              mermaid.get_workflow_diagram(workflow),
            )
            _ -> #(
              model,
              get_single_workflow(wfx_url: model.wfx_url, name: name),
            )
          }
      }
    }

    msg.UserClickedRefresh -> {
      case model.route {
        model.RouteJobs(page, limit) -> #(
          model,
          get_jobs(wfx_url: model.wfx_url, page: page, limit: limit),
        )
        model.RouteWorkflows(page, limit) -> #(
          model,
          get_workflows(wfx_url: model.wfx_url, page: page, limit: limit),
        )
        model.RouteWorkflowDetails(name) -> #(
          model,
          get_single_workflow(wfx_url: model.wfx_url, name: name),
        )
        model.RouteJobDetails(id) -> #(
          model,
          get_single_job(wfx_url: model.wfx_url, id: id, history: True),
        )
      }
    }

    // API responses
    msg.WfxSentJobs(response) -> #(
      Model(
        ..model,
        paginated_jobs: Some(model.PaginatedJobs(
          response: response
          |> result.map_error(fn(err) { model.RsvpError(err) }),
        )),
      ),
      effect.none(),
    )

    msg.WfxSentWorkflows(response) -> #(
      Model(
        ..model,
        paginated_workflows: Some(model.PaginatedWorkflows(
          response: response
          |> result.map_error(fn(err) { model.RsvpError(err) }),
        )),
      ),
      effect.none(),
    )

    msg.WfxSentSingleWorkflow(response) -> #(
      model.Model(
        ..model,
        workflow: Some(model.Workflow(
          response: response
            |> result.map_error(fn(err) { model.RsvpError(err) }),
          mermaid_diagram: None,
          copied_to_clipboard: False,
        )),
      ),
      case response {
        Ok(workflow) -> mermaid.get_workflow_diagram(workflow)
        Error(_) -> effect.none()
      },
    )

    msg.WfxSentSingleJob(response) -> #(
      model.Model(
        ..model,
        job: Some(model.Job(
          response: response
            |> result.map_error(fn(err) { model.RsvpError(err) }),
          mermaid_diagram: None,
          copied_to_clipboard: False,
        )),
      ),
      case response {
        Ok(job) -> mermaid.get_job_diagram(job)
        Error(_) -> effect.none()
      },
    )

    msg.WfxSentJobEvent(event_json) ->
      case json.parse(event_json, decoder.job_event_decoder()) {
        Ok(event) -> {
          let old_state =
            model.job
            |> option.map(fn(x) { x.response |> option.from_result })
            |> option.flatten
            |> option.map(fn(x) { x.status })
            |> option.flatten
          let new_model = model.merge_job_event(model, event)
          let new_job =
            new_model.job
            |> option.map(fn(x) { x.response |> option.from_result })
            |> option.flatten
          let new_state =
            new_job
            |> option.map(fn(x) { x.status })
            |> option.flatten
          #(new_model, case new_job, old_state != new_state {
            Some(job), True -> mermaid.get_job_diagram(job)
            _, _ -> effect.none()
          })
        }
        Error(_) -> {
          io.println("[BUG] Failed to decode job event: " <> event_json)
          #(model, effect.none())
        }
      }
    msg.MermaidGeneratedJobDiagram(diagram) ->
      case model.job {
        Some(job_model) -> #(
          Model(
            ..model,
            job: Some(model.Job(..job_model, mermaid_diagram: Some(diagram))),
          ),
          effect.none(),
        )
        None -> #(model, effect.none())
      }
    msg.MermaidGeneratedWorklowDiagram(diagram) ->
      case model.workflow {
        Some(workflow_model) -> #(
          Model(
            ..model,
            workflow: Some(
              model.Workflow(..workflow_model, mermaid_diagram: Some(diagram)),
            ),
          ),
          effect.none(),
        )
        None -> #(model, effect.none())
      }

    msg.CopyToClipboard(value) -> {
      let _ = clipboard.write_text(value)
      case model.route {
        model.RouteJobDetails(_) -> {
          let new_model = case model.job {
            None -> model
            Some(job) ->
              Model(
                ..model,
                job: Some(model.Job(..job, copied_to_clipboard: True)),
              )
          }
          #(new_model, effect.none())
        }
        model.RouteWorkflowDetails(_) -> {
          let new_model = case model.workflow {
            None -> model
            Some(workflow) ->
              Model(
                ..model,
                workflow: Some(
                  model.Workflow(..workflow, copied_to_clipboard: True),
                ),
              )
          }
          #(new_model, effect.none())
        }
        _ -> #(model, effect.none())
      }
    }
  }
}

fn get_jobs(
  wfx_url wfx_url: String,
  page page: Int,
  limit limit: Int,
) -> Effect(Msg) {
  let append = string_tree.append
  let url =
    string_tree.from_string(wfx_url)
    |> append("/jobs?sort=desc&offset=")
    |> append(int.to_string({ page - 1 } * limit))
    |> append("&limit=")
    |> append(int.to_string(limit))
    |> string_tree.to_string
  let handler =
    rsvp.expect_json(decoder.paginated_job_list_decoder(), msg.WfxSentJobs)
  rsvp.get(url, handler)
}

fn get_workflows(
  wfx_url wfx_url: String,
  page page: Int,
  limit limit: Int,
) -> Effect(Msg) {
  let append = string_tree.append
  let url =
    string_tree.from_string(wfx_url)
    |> append("/workflows?sort=asc&offset=")
    |> append(int.to_string({ page - 1 } * limit))
    |> append("&limit=")
    |> append(int.to_string(limit))
    |> string_tree.to_string
  let handler =
    rsvp.expect_json(
      decoder.paginated_workflow_list_decoder(),
      msg.WfxSentWorkflows,
    )
  rsvp.get(url, handler)
}

fn get_single_workflow(
  wfx_url wfx_url: String,
  name name: String,
) -> Effect(Msg) {
  let url = wfx_url <> "/workflows/" <> name
  let handler =
    rsvp.expect_json(decoder.workflow_decoder(), msg.WfxSentSingleWorkflow)
  rsvp.get(url, handler)
}

fn get_single_job(
  wfx_url wfx_url: String,
  id id: String,
  history history: Bool,
) -> Effect(Msg) {
  let append = string_tree.append
  let url =
    string_tree.from_string(wfx_url)
    |> append("/jobs/")
    |> append(id)
    |> append("?history=")
    |> append(bool.to_string(history))
    |> string_tree.to_string
  let handler = rsvp.expect_json(decoder.job_decoder(), msg.WfxSentSingleJob)
  rsvp.get(url, handler)
}

pub fn get_job_events(source: events.JobsEventSource) -> Effect(Msg) {
  effect.from(fn(dispatch) {
    events.start(source, fn(data) { dispatch(msg.WfxSentJobEvent(data)) }, fn() {
      // ignore errors as the browser will try to fix errors on its own
      Nil
    })
    Nil
  })
}

fn remove_prefix(s: String, prefix: String) -> String {
  case string.starts_with(s, prefix) {
    True -> string.drop_start(s, string.length(prefix))
    False -> s
  }
}
