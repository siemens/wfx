// SPDX-FileCopyrightText: 2026 Siemens AG
//
// SPDX-License-Identifier: Apache-2.0
//
// Author: Michael Adler <michael.adler@siemens.com>
import { Ok, Error } from "./gleam.mjs";

class AuthenticatedEventSource {
  constructor(url, accessToken) {
    this.url = url;
    this.accessToken = accessToken;
    this.controller = new AbortController();
    this.closed = false;
    this.connect();
  }

  async connect() {
    while (!this.closed) {
      try {
        const response = await fetch(this.url, {
          headers: {
            Accept: "text/event-stream",
            Authorization: `Bearer ${this.accessToken}`,
          },
          signal: this.controller.signal,
        });
        if (!response.ok || !response.body) {
          throw new Error(`SSE request failed with status ${response.status}`);
        }

        const reader = response.body.getReader();
        const decoder = new TextDecoder();
        let buffer = "";
        while (!this.closed) {
          const { value, done } = await reader.read();
          buffer += decoder.decode(value, { stream: !done });
          const messages = buffer.split(/\r?\n\r?\n/);
          buffer = messages.pop() || "";
          for (const message of messages) {
            const data = message
              .split(/\r?\n/)
              .filter((line) => line.startsWith("data:"))
              .map((line) => line.slice(5).trimStart())
              .join("\n");
            if (data) this.onmessage?.({ data });
          }
          if (done) break;
        }
      } catch (err) {
        if (!this.closed) {
          console.error("SSE error:", err);
          this.onerror?.(err);
        }
      }
      if (!this.closed)
        await new Promise((resolve) => setTimeout(resolve, 3000));
    }
  }

  close() {
    this.closed = true;
    this.controller.abort();
  }
}

export function createJobEventsSource(wfx_url, filter, accessToken) {
  // Remove trailing slash
  const base = wfx_url.replace(/\/$/, "");
  const fullPath = `${base}/jobs/events`;
  const url = new URL(fullPath);

  if (filter && filter.trim() !== "") {
    const params = new URLSearchParams(filter);
    params.forEach((value, key) => url.searchParams.append(key, value));
  }
  const finalUrl = url.toString();

  if (
    typeof EventSource !== "undefined" ||
    (accessToken && typeof fetch !== "undefined")
  ) {
    try {
      console.log("SSE: subscribing to URL", finalUrl);
      const source = accessToken
        ? new AuthenticatedEventSource(finalUrl, accessToken)
        : new EventSource(finalUrl);
      return new Ok(source);
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
