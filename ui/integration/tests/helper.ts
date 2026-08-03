// SPDX-FileCopyrightText: 2026 Siemens AG
//
// SPDX-License-Identifier: Apache-2.0
//
// Author: Michael Adler <michael.adler@siemens.com>

import { APIRequestContext } from "@playwright/test";
import { randomUUID } from "crypto";
import { promises as fs } from "fs";
import YAML from "yaml";

/**
 * Marker attached to every resource created by these tests. Jobs carry it as a
 * tag, workflows (which don't support tags) as part of their name. Cleanup only
 * touches resources carrying this marker so that pre-existing data is preserved.
 */
export const TEST_TAG = "playwright-e2e";

export async function createWorkflow(
  baseUrl: string,
  request: APIRequestContext,
): Promise<Record<string, unknown>> {
  const yamlContent = await fs.readFile(
    "../../workflow/dau/wfx.workflow.dau.direct.yml",
    "utf-8",
  );
  const workflow = YAML.parse(yamlContent) as Record<string, unknown>;
  workflow.name = `${workflow.name}.${TEST_TAG}.${randomUUID().substring(0, 8)}`;
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
  baseUrl: string,
  request: APIRequestContext,
  workflow: string,
): Promise<Record<string, unknown>> {
  const jobRequest = {
    clientId: "rpi",
    workflow: workflow,
    tags: [TEST_TAG],
  };

  const response = await request.post(`${baseUrl}/api/wfx/v1/jobs`, {
    data: jobRequest,
  });

  if (!response.ok()) {
    const errorBody = await response.text();
    throw new Error(
      `Failed to create job: ${response.status()} - ${errorBody}`,
    );
  }
  const job = await response.json();
  return job;
}

export async function deleteWorkflow(
  baseUrl: string,
  request: APIRequestContext,
  name: string,
): Promise<void> {
  const response = await request.delete(
    `${baseUrl}/api/wfx/v1/workflows/${name}`,
  );
  if (!response.ok() && response.status() !== 404) {
    const errorBody = await response.text();
    throw new Error(
      `Failed to delete workflow: ${response.status()} - ${errorBody}`,
    );
  }
}

export async function deleteJob(
  baseUrl: string,
  request: APIRequestContext,
  job_id: string,
): Promise<void> {
  const response = await request.delete(`${baseUrl}/api/wfx/v1/jobs/${job_id}`);
  if (!response.ok() && response.status() !== 404) {
    const errorBody = await response.text();
    throw new Error(
      `Failed to delete job: ${response.status()} - ${errorBody}`,
    );
  }
}

async function listAll(
  baseUrl: string,
  request: APIRequestContext,
  path: string,
  params = "",
): Promise<Record<string, unknown>[]> {
  const items: Record<string, unknown>[] = [];
  const limit = 100;
  for (let offset = 0; ; offset += limit) {
    const response = await request.get(
      `${baseUrl}${path}?limit=${limit}&offset=${offset}${params}`,
    );
    if (!response.ok()) {
      const errorBody = await response.text();
      throw new Error(
        `Failed to list ${path}: ${response.status()} - ${errorBody}`,
      );
    }
    const body = (await response.json()) as {
      content?: Record<string, unknown>[];
    };
    const page = body.content ?? [];
    items.push(...page);
    if (page.length < limit) {
      return items;
    }
  }
}

/**
 * Removes the jobs and workflows created by these tests, identified by TEST_TAG.
 * Jobs are deleted first because workflows cannot be removed while jobs still
 * reference them.
 */
export async function cleanupDatabase(
  baseUrl: string,
  request: APIRequestContext,
): Promise<void> {
  const jobs = await listAll(
    baseUrl,
    request,
    "/api/wfx/v1/jobs",
    `&tag=${encodeURIComponent(TEST_TAG)}`,
  );
  for (const job of jobs) {
    await deleteJob(baseUrl, request, job.id as string);
  }
  const workflows = await listAll(baseUrl, request, "/api/wfx/v1/workflows");
  for (const workflow of workflows) {
    const name = workflow.name as string;
    if (!name.includes(`.${TEST_TAG}.`)) {
      continue;
    }
    await deleteWorkflow(baseUrl, request, name);
  }
}
