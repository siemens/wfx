// SPDX-FileCopyrightText: 2025 Siemens AG
//
// SPDX-License-Identifier: Apache-2.0
//
// Author: Michael Adler <michael.adler@siemens.com>
import gleam/json
import gleam/list
import gleam/option.{Some}
import simplifile

import wfx
import wfx/decoder
import wfx/utils

pub fn job_event_decoder_test() {
  let assert Ok(content) = simplifile.read(from: "test_data/job_event.json")
  let assert Ok(ev) = json.parse(content, decoder.job_event_decoder())
  assert ev.action == wfx.ActionUpdateStatus
  let assert Some(status) = ev.job.status
  assert status.state == "PROGRESS"
}

pub fn paginated_job_list_decoder_test() {
  let assert Ok(content) = simplifile.read(from: "test_data/jobs.json")
  let assert Ok(parsed) =
    json.parse(content, decoder.paginated_job_list_decoder())
  assert parsed.pagination.total == 1
}

pub fn paginated_workflow_list_decoder_test() {
  let assert Ok(content) = simplifile.read(from: "test_data/workflows.json")
  let assert Ok(parsed) =
    json.parse(content, decoder.paginated_workflow_list_decoder())
  assert parsed.pagination.total == 1
}

pub fn workflow_initial_state_test() {
  let assert Ok(content) = simplifile.read(from: "test_data/workflows.json")
  let assert Ok(parsed) =
    json.parse(content, decoder.paginated_workflow_list_decoder())
  let assert Ok(workflow) = parsed.content |> list.first()
  let states = utils.workflow_initial_states(workflow)
  assert states == ["BACKLOG"]
}

pub fn workflow_final_state_test() {
  let assert Ok(content) = simplifile.read(from: "test_data/workflows.json")
  let assert Ok(parsed) =
    json.parse(content, decoder.paginated_workflow_list_decoder())
  let assert Ok(workflow) = parsed.content |> list.first()
  let states = utils.workflow_final_states(workflow)
  assert states == ["DISCARDED", "DONE"]
}
