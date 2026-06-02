// SPDX-FileCopyrightText: 2026 Siemens AG
//
// SPDX-License-Identifier: Apache-2.0
//
// Author: Michael Adler <michael.adler@siemens.com>
import rsvp

import model
import wfx

pub type Msg {
  // Subject Verb Object
  UserClickedRefresh

  DocumentChangedRoute(model.Route)
  CopyToClipboard(String)

  MermaidGeneratedWorklowDiagram(String)
  MermaidGeneratedJobDiagram(String)

  WfxSentJobs(Result(wfx.PaginatedJobList, rsvp.Error(String)))
  WfxSentWorkflows(Result(wfx.PaginatedWorkflowList, rsvp.Error(String)))
  WfxSentSingleWorkflow(Result(wfx.Workflow, rsvp.Error(String)))
  WfxSentSingleJob(Result(wfx.Job, rsvp.Error(String)))
  WfxSentJobEvent(String)
}
