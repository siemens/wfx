// SPDX-FileCopyrightText: 2025 Siemens AG
//
// SPDX-License-Identifier: Apache-2.0
//
// Author: Michael Adler <michael.adler@siemens.com>

import { APIRequestContext } from "@playwright/test";
import { promises as fs } from "fs";
import YAML from "yaml";

export async function createWorkflow(
  baseUrl: String,
  request: APIRequestContext,
): Promise<Record<string, unknown>> {
  const yamlContent = await fs.readFile(
    "../../workflow/dau/wfx.workflow.dau.direct.yml",
    "utf-8",
  );
  const workflow = YAML.parse(yamlContent) as Record<string, unknown>;
  const response = await request.post(`${baseUrl}/api/wfx/v1/workflows`, {
    data: workflow,
  });

  if (!response.ok()) {
    const errorBody = await response.text();
    throw new Error(
      `Failed to create workflow: ${response.status()} - ${errorBody}`,
    );
  }
  return workflow;
}

export async function createJob(
  baseUrl: String,
  request: APIRequestContext,
  workflow: String,
): Promise<Record<string, unknown>> {
  const jobRequest = {
    clientId: "rpi",
    workflow: workflow,
  };

  const response = await request.post(`${baseUrl}/api/wfx/v1/jobs`, {
    data: jobRequest,
  });

  if (!response.ok()) {
    const errorBody = await response.text();
    throw new Error(
      `Failed to create workflow: ${response.status()} - ${errorBody}`,
    );
  }
  const job = await response.json();
  return job;
}

export async function deleteWorkflow(
  baseUrl: String,
  request: APIRequestContext,
  name: string,
): Promise<void> {
  const response = await request.delete(
    `${baseUrl}/api/wfx/v1/workflows/${name}`,
  );
  if (!response.ok()) {
    const errorBody = await response.text();
    throw new Error(
      `Failed to delete workflow: ${response.status()} - ${errorBody}`,
    );
  }
}

export async function deleteJob(
  baseUrl: String,
  request: APIRequestContext,
  job_id: string,
): Promise<void> {
  const response = await request.delete(`${baseUrl}/api/wfx/v1/jobs/${job_id}`);
  if (!response.ok()) {
    const errorBody = await response.text();
    throw new Error(
      `Failed to delete job: ${response.status()} - ${errorBody}`,
    );
  }
}
