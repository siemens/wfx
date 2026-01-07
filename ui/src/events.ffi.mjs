// SPDX-FileCopyrightText: 2025 Siemens AG
//
// SPDX-License-Identifier: Apache-2.0
//
// Author: Michael Adler <michael.adler@siemens.com>
import { Ok, Error } from "./gleam.mjs";

export function createJobEventsSource(wfx_url, filter) {
  // Remove trailing slash
  const base = wfx_url.replace(/\/$/, "");
  const fullPath = `${base}/jobs/events`;
  const url = new URL(fullPath);

  if (filter && filter.trim() !== "") {
    const params = new URLSearchParams(filter);
    params.forEach((value, key) => url.searchParams.append(key, value));
  }
  const finalUrl = url.toString();

  if (typeof EventSource !== "undefined") {
    try {
      console.log("SSE: subscribing to URL", finalUrl);
      return new Ok(new EventSource(finalUrl));
    } catch (err) {
      // Handle instantiation errors (e.g., invalid URL, security issues)
      console.error("Failed to create EventSource:", err);
      return new Error(undefined);
    }
  }
  console.warn("EventSource is not supported in this environment.");
  return new Error(undefined);
}

export function start(source, on_message, on_error) {
  source.onmessage = (event) => {
    if (event.data) {
      on_message(event.data);
    }
  };

  source.onerror = (err) => {
    console.error("SSE error:", JSON.stringify(err));
    on_error();
  };
}

export function stop(source) {
  console.log("SSE: closing");
  source.close();
}
