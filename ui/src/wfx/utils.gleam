// SPDX-FileCopyrightText: 2025 Siemens AG
//
// SPDX-License-Identifier: Apache-2.0
//
// Author: Michael Adler <michael.adler@siemens.com>
import gleam/list
import gleam/option.{type Option, None, Some}
import gleam/set
import gleam/string

import wfx

pub fn workflow_initial_states(workflow: wfx.Workflow) -> List(String) {
  // add all known states
  let state_names =
    workflow.states |> list.map(fn(x) { x.name }) |> set.from_list()
  // remove all states which are the target of a transition
  workflow.transitions
  |> list.fold(state_names, fn(acc, transition) {
    case transition.from == transition.to {
      True -> acc
      False -> acc |> set.delete(transition.to)
    }
  })
  |> set.to_list
  |> list.sort(string.compare)
}

pub fn workflow_final_states(workflow: wfx.Workflow) -> List(String) {
  // add all known states
  let state_names =
    workflow.states |> list.map(fn(x) { x.name }) |> set.from_list()
  // remove all states which are the source of a transition
  workflow.transitions
  |> list.fold(state_names, fn(acc, transition) {
    case transition.from == transition.to {
      True -> acc
      False -> acc |> set.delete(transition.from)
    }
  })
  |> set.to_list
  |> list.sort(string.compare)
}

pub fn page_from_pagination(pagination: wfx.Pagination) -> Int {
  pagination.offset / pagination.limit + 1
}

pub fn job_group(job: wfx.Job) -> Option(wfx.Group) {
  let workflow = job.workflow
  case job.status, workflow.groups {
    Some(status), Some(groups) -> {
      list.find(groups, fn(group) { list.contains(group.states, status.state) })
      |> option.from_result()
    }
    _, _ -> None
  }
}
