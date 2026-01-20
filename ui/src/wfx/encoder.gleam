// SPDX-FileCopyrightText: 2025 Siemens AG
//
// SPDX-License-Identifier: Apache-2.0
//
// Author: Michael Adler <michael.adler@siemens.com>
import gleam/dict.{type Dict}
import gleam/dynamic.{type Dynamic}
import gleam/json
import gleam/time/calendar
import gleam/time/timestamp

import wfx

pub fn job(job: wfx.Job) -> json.Json {
  json.object([
    #("id", json.string(job.id)),
    #("client_id", json.string(job.client_id)),
    #("status", json.nullable(job.status, job_status)),
    #("definition", job.definition |> encode_dict),
    #("workflow", workflow(job.workflow)),
    #("tags", json.array(job.tags, json.string)),
    #("stime", json.nullable(job.stime, timestamp)),
    #("mtime", json.nullable(job.mtime, timestamp)),
    #("history", json.nullable(job.history, fn(xs) { json.array(xs, history) })),
  ])
}

pub fn job_status(status: wfx.JobStatus) -> json.Json {
  json.object([
    #("state", json.string(status.state)),
    #("client_id", json.nullable(status.client_id, json.string)),
    #("progress", json.nullable(status.progress, json.int)),
    #("message", json.nullable(status.message, json.string)),
    #("definition_hash", json.nullable(status.definition_hash, json.string)),
    #("context", json.nullable(status.context, encode_dict)),
  ])
}

pub fn history(history: wfx.History) -> json.Json {
  json.object([
    #("mtime", json.string(history.mtime)),
    #("status", json.nullable(history.status, job_status)),
    #("definition", json.nullable(history.definition, encode_dict)),
  ])
}

pub fn workflow(workflow: wfx.Workflow) -> json.Json {
  json.object([
    #("name", json.string(workflow.name)),
    #("description", json.nullable(workflow.description, json.string)),
    #("states", json.array(workflow.states, state)),
    #("groups", json.nullable(workflow.groups, fn(x) { json.array(x, group) })),
    #("transitions", json.array(workflow.transitions, transition)),
  ])
}

pub fn state(state: wfx.State) -> json.Json {
  json.object([
    #("name", json.string(state.name)),
    #("description", json.nullable(state.description, json.string)),
  ])
}

pub fn group(group: wfx.Group) -> json.Json {
  json.object([
    #("name", json.string(group.name)),
    #("description", json.nullable(group.description, json.string)),
    #("states", json.array(group.states, json.string)),
  ])
}

pub fn transition(transition: wfx.Transition) -> json.Json {
  json.object([
    #("from", json.string(transition.from)),
    #("to", json.string(transition.to)),
    #("description", json.nullable(transition.description, json.string)),
    #("eligible", eligible_enum(transition.eligible)),
    #("action", json.nullable(transition.action, action_enum)),
  ])
}

pub fn eligible_enum(e: wfx.EligibleEnum) -> json.Json {
  case e {
    wfx.EligibleClient -> json.string("CLIENT")
    wfx.EligibleWfx -> json.string("WFX")
  }
}

pub fn action_enum(a: wfx.ActionEnum) -> json.Json {
  case a {
    wfx.ActionImmediate -> json.string("IMMEDIATE")
    wfx.ActionWait -> json.string("WAIT")
  }
}

fn timestamp(ts: timestamp.Timestamp) -> json.Json {
  ts |> timestamp.to_rfc3339(calendar.local_offset()) |> json.string
}

fn encode_dict(dict: Dict(String, Dynamic)) -> json.Json {
  json.dict(dict, fn(x) { x }, encode_dynamic)
}

@external(javascript, "./encoder.ffi.mjs", "encodeDynamic")
fn encode_dynamic(dyn: Dynamic) -> json.Json
