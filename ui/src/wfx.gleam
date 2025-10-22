// SPDX-FileCopyrightText: 2025 Siemens AG
//
// SPDX-License-Identifier: Apache-2.0
//
// Author: Michael Adler <michael.adler@siemens.com>
import gleam/dict.{type Dict}
import gleam/dynamic.{type Dynamic}
import gleam/option.{type Option}
import gleam/time/timestamp

pub type Job {
  Job(
    id: String,
    client_id: String,
    workflow: Workflow,
    tags: List(String),
    definition: Dict(String, Dynamic),
    status: Option(JobStatus),
    stime: Option(timestamp.Timestamp),
    mtime: Option(timestamp.Timestamp),
    history: Option(List(History)),
  )
}

pub type JobEvent {
  JobEvent(ctime: String, action: JobEventAction, job: Job, tags: List(String))
}

pub type JobEventAction {
  ActionCreate
  ActionDelete
  ActionAddTags
  ActionDeleteTags
  ActionUpdateStatus
  ActionUpdateDefinition
}

pub type Workflow {
  Workflow(
    name: String,
    description: Option(String),
    states: List(State),
    groups: Option(List(Group)),
    transitions: List(Transition),
  )
}

pub type Group {
  Group(name: String, description: Option(String), states: List(String))
}

pub type Transition {
  Transition(
    from: String,
    to: String,
    description: Option(String),
    eligible: EligibleEnum,
    action: Option(ActionEnum),
  )
}

pub type EligibleEnum {
  EligibleClient
  EligibleWfx
}

pub type ActionEnum {
  ActionImmediate
  ActionWait
}

pub type State {
  State(name: String, description: Option(String))
}

pub type JobStatus {
  JobStatus(
    state: String,
    client_id: Option(String),
    progress: Option(Int),
    message: Option(String),
    definition_hash: Option(String),
    context: Option(Dict(String, Dynamic)),
  )
}

pub type History {
  History(
    mtime: String,
    status: Option(JobStatus),
    definition: Option(Dict(String, Dynamic)),
  )
}

pub type Health {
  Health(details: Details, status: String)
}

pub type Details {
  Details(persistence: Persistence)
}

pub type Persistence {
  Persistence(status: String, timestamp: String)
}

pub type Pagination {
  Pagination(limit: Int, offset: Int, total: Int)
}

pub type PaginatedJobList {
  PaginatedJobList(pagination: Pagination, content: List(Job))
}

pub type PaginatedWorkflowList {
  PaginatedWorkflowList(pagination: Pagination, content: List(Workflow))
}
