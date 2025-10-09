// SPDX-FileCopyrightText: 2025 Siemens AG
//
// SPDX-License-Identifier: Apache-2.0
//
// Author: Michael Adler <michael.adler@siemens.com>
import gleam/list
import gleam/option.{type Option, None, Some}
import gleam/string_tree

pub type JobsEventSource

@external(javascript, "./events.ffi.mjs", "createJobEventsSource")
pub fn do_create_job_events_source(
  _wfx_url: String,
  _filter: String,
) -> Result(JobsEventSource, Nil) {
  Error(Nil)
}

pub fn create_job_events_source(
  wfx_url: String,
  job_ids: Option(List(String)),
) -> Option(JobsEventSource) {
  let filter = case job_ids {
    Some(job_ids) ->
      "jobIds="
      <> {
        job_ids
        |> list.map(fn(s) { string_tree.from_string(s) })
        |> string_tree.join(",")
        |> string_tree.to_string
      }
    None -> ""
  }
  do_create_job_events_source(wfx_url, filter) |> option.from_result
}

@external(javascript, "./events.ffi.mjs", "start")
pub fn start(
  _source: JobsEventSource,
  _on_message: fn(String) -> Nil,
  _on_error: fn() -> Nil,
) -> Nil {
  Nil
}

@external(javascript, "./events.ffi.mjs", "stop")
pub fn stop(_source: JobsEventSource) -> Nil {
  Nil
}
