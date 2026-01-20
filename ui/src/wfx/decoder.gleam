// SPDX-FileCopyrightText: 2025 Siemens AG
//
// SPDX-License-Identifier: Apache-2.0
//
// Author: Michael Adler <michael.adler@siemens.com>
import gleam/dict
import gleam/dynamic/decode
import gleam/option.{None}
import gleam/time/timestamp

import wfx

pub fn health_decoder() -> decode.Decoder(wfx.Health) {
  use status <- decode.field("status", decode.string)
  use details <- decode.field("details", details_decoder())
  decode.success(wfx.Health(status: status, details: details))
}

pub fn details_decoder() -> decode.Decoder(wfx.Details) {
  use persistence <- decode.field("persistence", persistence_decoder())
  decode.success(wfx.Details(persistence: persistence))
}

pub fn persistence_decoder() -> decode.Decoder(wfx.Persistence) {
  use status <- decode.field("status", decode.string)
  use timestamp <- decode.field("timestamp", decode.string)
  decode.success(wfx.Persistence(status: status, timestamp: timestamp))
}

pub fn eligible_enum_decoder() -> decode.Decoder(wfx.EligibleEnum) {
  use decoded_string <- decode.then(decode.string)
  case decoded_string {
    // Return succeeding decoders for valid strings
    "CLIENT" -> decode.success(wfx.EligibleClient)
    "WFX" -> decode.success(wfx.EligibleWfx)
    _ -> decode.failure(wfx.EligibleClient, "Invalid EligibleEnum")
  }
}

pub fn action_enum_decoder() -> decode.Decoder(wfx.ActionEnum) {
  use decoded_string <- decode.then(decode.string)
  case decoded_string {
    // Return succeeding decoders for valid strings
    "IMMEDIATE" -> decode.success(wfx.ActionImmediate)
    "WAIT" -> decode.success(wfx.ActionWait)
    _ -> decode.failure(wfx.ActionImmediate, "Invalid ActionEnum")
  }
}

pub fn transition_decoder() -> decode.Decoder(wfx.Transition) {
  use from <- decode.field("from", decode.string)
  use to <- decode.field("to", decode.string)
  use description <- decode.optional_field(
    "description",
    None,
    decode.optional(decode.string),
  )
  use eligible <- decode.field("eligible", eligible_enum_decoder())
  use action <- decode.optional_field(
    "action",
    None,
    decode.optional(action_enum_decoder()),
  )
  decode.success(wfx.Transition(
    from: from,
    to: to,
    description: description,
    eligible: eligible,
    action: action,
  ))
}

pub fn job_status_decoder() -> decode.Decoder(wfx.JobStatus) {
  use state <- decode.field("state", decode.string)
  use client_id <- decode.optional_field(
    "clientId",
    None,
    decode.optional(decode.string),
  )
  use progress <- decode.optional_field(
    "progress",
    None,
    decode.optional(decode.int),
  )
  use message <- decode.optional_field(
    "message",
    None,
    decode.optional(decode.string),
  )
  use definition_hash <- decode.optional_field(
    "definition_hash",
    None,
    decode.optional(decode.string),
  )
  use context <- decode.optional_field(
    "context",
    None,
    decode.optional(decode.dict(decode.string, decode.dynamic)),
  )
  decode.success(wfx.JobStatus(
    state: state,
    client_id: client_id,
    progress: progress,
    message: message,
    definition_hash: definition_hash,
    context: context,
  ))
}

pub fn history_decoder() -> decode.Decoder(wfx.History) {
  use mtime <- decode.field("mtime", decode.string)
  use status <- decode.optional_field(
    "status",
    None,
    decode.optional(job_status_decoder()),
  )
  use definition <- decode.optional_field(
    "definition",
    None,
    decode.optional(decode.dict(decode.string, decode.dynamic)),
  )
  decode.success(wfx.History(
    mtime: mtime,
    status: status,
    definition: definition,
  ))
}

pub fn state_decoder() -> decode.Decoder(wfx.State) {
  use name <- decode.field("name", decode.string)
  use description <- decode.optional_field(
    "description",
    None,
    decode.optional(decode.string),
  )
  decode.success(wfx.State(name: name, description: description))
}

pub fn group_decoder() -> decode.Decoder(wfx.Group) {
  use name <- decode.field("name", decode.string)
  use description <- decode.optional_field(
    "description",
    None,
    decode.optional(decode.string),
  )
  use states <- decode.field("states", decode.list(decode.string))
  decode.success(wfx.Group(name: name, description: description, states: states))
}

pub fn job_decoder() -> decode.Decoder(wfx.Job) {
  use id <- decode.field("id", decode.string)
  use client_id <- decode.field("clientId", decode.string)
  use workflow <- decode.field("workflow", workflow_decoder())
  use tags <- decode.optional_field(
    "tags",
    None,
    decode.optional(decode.list(decode.string)),
  )
  use definition <- decode.optional_field(
    "definition",
    dict.new(),
    decode.dict(decode.string, decode.dynamic),
  )
  use status <- decode.optional_field(
    "status",
    None,
    decode.optional(job_status_decoder()),
  )
  use stime <- decode.optional_field(
    "stime",
    None,
    decode.optional(decode.string),
  )
  use mtime <- decode.optional_field(
    "mtime",
    None,
    decode.optional(decode.string),
  )
  use history <- decode.optional_field(
    "history",
    None,
    decode.optional(decode.list(history_decoder())),
  )
  decode.success(wfx.Job(
    id: id,
    client_id: client_id,
    workflow: workflow,
    tags: tags |> option.unwrap([]),
    definition: definition,
    status: status,
    stime: stime
      |> option.map(timestamp.parse_rfc3339)
      |> option.map(option.from_result)
      |> option.flatten(),
    mtime: mtime
      |> option.map(timestamp.parse_rfc3339)
      |> option.map(option.from_result)
      |> option.flatten(),
    history: history,
  ))
}

pub fn pagination_decoder() -> decode.Decoder(wfx.Pagination) {
  use limit <- decode.field("limit", decode.int)
  use offset <- decode.field("offset", decode.int)
  use total <- decode.field("total", decode.int)
  decode.success(wfx.Pagination(limit: limit, offset: offset, total: total))
}

pub fn paginated_job_list_decoder() -> decode.Decoder(wfx.PaginatedJobList) {
  use pagination <- decode.field("pagination", pagination_decoder())
  use content <- decode.field("content", decode.list(job_decoder()))
  decode.success(wfx.PaginatedJobList(pagination: pagination, content: content))
}

pub fn paginated_workflow_list_decoder() -> decode.Decoder(
  wfx.PaginatedWorkflowList,
) {
  use pagination <- decode.field("pagination", pagination_decoder())
  use content <- decode.field("content", decode.list(workflow_decoder()))
  decode.success(wfx.PaginatedWorkflowList(
    pagination: pagination,
    content: content,
  ))
}

pub fn workflow_decoder() -> decode.Decoder(wfx.Workflow) {
  use name <- decode.field("name", decode.string)
  use description <- decode.optional_field(
    "description",
    None,
    decode.optional(decode.string),
  )
  use states <- decode.optional_field(
    "states",
    None,
    decode.optional(decode.list(state_decoder())),
  )
  use groups <- decode.optional_field(
    "groups",
    None,
    decode.optional(decode.list(group_decoder())),
  )
  use transitions <- decode.optional_field(
    "transitions",
    None,
    decode.optional(decode.list(transition_decoder())),
  )
  decode.success(wfx.Workflow(
    name: name,
    description: description,
    states: states |> option.unwrap([]),
    groups: groups,
    transitions: transitions |> option.unwrap([]),
  ))
}

pub fn job_event_action_decoder() -> decode.Decoder(wfx.JobEventAction) {
  use variant <- decode.then(decode.string)
  case variant {
    "CREATE" -> decode.success(wfx.ActionCreate)
    "DELETE" -> decode.success(wfx.ActionDelete)
    "ADD_TAGS" -> decode.success(wfx.ActionAddTags)
    "DELETE_TAGS" -> decode.success(wfx.ActionDeleteTags)
    "UPDATE_STATUS" -> decode.success(wfx.ActionUpdateStatus)
    "UPDATE_DEFINITION" -> decode.success(wfx.ActionUpdateDefinition)
    _ -> decode.failure(wfx.ActionCreate, "JobEventAction")
  }
}

pub fn job_event_decoder() -> decode.Decoder(wfx.JobEvent) {
  use ctime <- decode.field("ctime", decode.string)
  use action <- decode.field("action", job_event_action_decoder())
  use job <- decode.field("job", job_decoder())
  use tags <- decode.optional_field(
    "tags",
    None,
    decode.optional(decode.list(decode.string)),
  )
  decode.success(wfx.JobEvent(
    ctime: ctime,
    action: action,
    job: job,
    tags: tags |> option.unwrap([]),
  ))
}
