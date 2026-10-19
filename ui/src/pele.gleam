// SPDX-FileCopyrightText: 2026 Siemens AG
//
// SPDX-License-Identifier: Apache-2.0
//
// Author: Michael Adler <michael.adler@siemens.com>
import gleam/dict
import gleam/int
import gleam/list
import gleam/option.{None, Some}
import gleam/result
import gleam/string
import gleam/string_tree.{type StringTree}
import lustre/effect.{type Effect}
import msg.{type Msg, PeleGeneratedJobDiagram, PeleGeneratedWorkflowDiagram}
import wfx
import wfx/utils

@external(javascript, "./pele.ffi.mjs", "renderDiagram")
fn render_diagram(graph_definition: String) -> String

pub fn get_workflow_diagram(workflow: wfx.Workflow) -> Effect(Msg) {
  let graph_definition = workflow_diagram_source(workflow)
  effect.from(fn(dispatch) {
    dispatch(PeleGeneratedWorkflowDiagram(render_diagram(graph_definition)))
    Nil
  })
}

pub fn get_job_diagram(job: wfx.Job) -> Effect(Msg) {
  let graph_definition = job_to_pele(job) |> string_tree.to_string
  effect.from(fn(dispatch) {
    dispatch(PeleGeneratedJobDiagram(render_diagram(graph_definition)))
    Nil
  })
}

pub fn workflow_diagram_source(workflow: wfx.Workflow) -> String {
  workflow_to_pele(workflow) |> string_tree.to_string
}

fn workflow_to_pele(workflow: wfx.Workflow) -> StringTree {
  let state_ids =
    workflow.states
    |> list.index_map(fn(state, index) {
      #(state.name, "S" <> int.to_string(index))
    })
  let aliases = dict.from_list(state_ids)
  let states =
    workflow.states
    |> list.index_map(fn(state, index) {
      "    state \""
      <> escape_state_name(state.name)
      <> "\" as S"
      <> int.to_string(index)
    })
    |> string.join("\n")
    |> string_tree.from_string

  let initials =
    utils.workflow_initial_states(workflow)
    |> list.map(fn(name) {
      "    [*] --> " <> result.unwrap(dict.get(aliases, name), "")
    })
    |> string.join("\n")
    |> string_tree.from_string

  let transitions =
    workflow.transitions
    |> list.map(fn(transition) {
      "    "
      <> result.unwrap(dict.get(aliases, transition.from), "")
      <> " --> "
      <> result.unwrap(dict.get(aliases, transition.to), "")
      <> ":"
      <> {
        case transition.eligible {
          wfx.EligibleClient -> "CLIENT"
          wfx.EligibleWfx -> "WFX"
        }
      }
    })
    |> string.join("\n")
    |> string_tree.from_string

  let finals =
    utils.workflow_final_states(workflow)
    |> list.map(fn(name) {
      "    " <> result.unwrap(dict.get(aliases, name), "") <> " --> [*]"
    })
    |> string.join("\n")
    |> string_tree.from_string

  string_tree.join(
    [
      string_tree.from_string("stateDiagram-v2"),
      states,
      initials,
      transitions,
      finals,
    ],
    "\n",
  )
}

fn escape_state_name(name: String) -> String {
  name
  |> string.replace(each: "&", with: "&amp;")
  |> string.replace(each: "\"", with: "&quot;")
  |> string.replace(each: "<", with: "&lt;")
  |> string.replace(each: ">", with: "&gt;")
}

fn job_to_pele(job: wfx.Job) -> StringTree {
  let color = case job.status {
    Some(status) -> {
      let state_id =
        job.workflow.states
        |> list.index_map(fn(state, index) {
          #(state.name, "S" <> int.to_string(index))
        })
        |> dict.from_list
        |> dict.get(status.state)
        |> result.unwrap("")
      string_tree.from_string("\n    classDef cl_")
      |> string_tree.append(state_id)
      |> string_tree.append(" fill:#ADD8E6")
      |> string_tree.append("\n    class ")
      |> string_tree.append(state_id)
      |> string_tree.append(" cl_")
      |> string_tree.append(state_id)
    }
    None -> string_tree.new()
  }

  workflow_to_pele(job.workflow) |> string_tree.append_tree(color)
}
