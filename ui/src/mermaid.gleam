// SPDX-FileCopyrightText: 2025 Siemens AG
//
// SPDX-License-Identifier: Apache-2.0
//
// Author: Michael Adler <michael.adler@siemens.com>
import gleam/javascript/promise.{type Promise}
import gleam/list
import gleam/option.{None, Some}
import gleam/string_tree.{type StringTree}
import lustre/effect.{type Effect}
import msg.{type Msg, MermaidGeneratedJobDiagram, MermaidGeneratedWorklowDiagram}
import wfx
import wfx/utils

@external(javascript, "./mermaid.ffi.mjs", "renderDiagram")
fn render_diagram(graph_definition: String) -> Promise(String)

pub fn get_workflow_diagram(workflow: wfx.Workflow) -> Effect(Msg) {
  let graph_definition = workflow_to_mermaid(workflow) |> string_tree.to_string
  effect.from(do_generate_workflow_diagram(graph_definition, _))
}

pub fn get_job_diagram(job: wfx.Job) -> Effect(Msg) {
  let graph_definition = job_to_mermaid(job) |> string_tree.to_string
  effect.from(do_generate_job_diagram(graph_definition, _))
}

fn do_generate_workflow_diagram(
  graph_definition: String,
  dispatch: fn(Msg) -> Nil,
) -> Nil {
  render_diagram(graph_definition)
  |> promise.map(fn(html) { MermaidGeneratedWorklowDiagram(html) })
  |> promise.tap(dispatch)
  Nil
}

fn do_generate_job_diagram(
  graph_definition: String,
  dispatch: fn(Msg) -> Nil,
) -> Nil {
  render_diagram(graph_definition)
  |> promise.map(fn(html) { MermaidGeneratedJobDiagram(html) })
  |> promise.tap(dispatch)
  Nil
}

fn workflow_to_mermaid(workflow: wfx.Workflow) -> StringTree {
  let header = string_tree.new() |> string_tree.append("stateDiagram-v2")

  let initials =
    utils.workflow_initial_states(workflow)
    |> list.map(fn(x) { "    [*] --> " <> x })
    |> string_tree.from_strings

  let finals =
    utils.workflow_final_states(workflow)
    |> list.map(fn(x) { "    " <> x <> " --> [*]" })
    |> string_tree.from_strings

  let transitions =
    workflow.transitions
    |> list.map(fn(x) {
      string_tree.new()
      |> string_tree.append("    ")
      |> string_tree.append(x.from)
      |> string_tree.append(" --> ")
      |> string_tree.append(x.to)
      |> string_tree.append(":")
      |> string_tree.append({
        case x.eligible {
          wfx.EligibleClient -> "CLIENT"
          wfx.EligibleWfx -> "WFX"
        }
      })
    })
    |> string_tree.join("\n")

  string_tree.join([header, initials, transitions, finals], "\n")
}

fn job_to_mermaid(job: wfx.Job) -> StringTree {
  let color = case job.status {
    Some(status) ->
      string_tree.from_string("\n    classDef cl_")
      |> string_tree.append(status.state)
      |> string_tree.append(" fill:#ADD8E6")
      |> string_tree.append("\n    class ")
      |> string_tree.append(status.state)
      |> string_tree.append(" cl_")
      |> string_tree.append(status.state)
    None -> string_tree.new()
  }

  workflow_to_mermaid(job.workflow) |> string_tree.append_tree(color)
}
