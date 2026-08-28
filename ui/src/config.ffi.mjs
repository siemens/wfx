// SPDX-FileCopyrightText: 2026 Siemens AG
//
// SPDX-License-Identifier: Apache-2.0
//
// Author: Michael Adler <michael.adler@siemens.com>
export function loadConfig() {
  if (window.loadConfig) {
    return window.loadConfig();
  }
  console.warn("No config provided; using default config.");
  return { wfx_url: "http://127.0.0.1:8081/api/wfx/v1", base_path: "" };
}
