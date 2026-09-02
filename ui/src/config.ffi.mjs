// SPDX-FileCopyrightText: 2026 Siemens AG
//
// SPDX-License-Identifier: Apache-2.0
//
// Author: Michael Adler <michael.adler@siemens.com>
import { None, Some } from "../gleam_stdlib/gleam/option.mjs";

const oauthDefaults = {
  issuer: "",
  client_id: "",
  scope: "openid email profile",
};

const defaults = {
  wfx_url: "http://127.0.0.1:8081/api/wfx/v1",
  base_path: "",
  oauth: null,
};

export function loadConfig() {
  const config = window.loadConfig
    ? { ...defaults, ...window.loadConfig() }
    : defaults;
  if (!window.loadConfig) {
    console.warn("No config provided; using default config.");
  }
  return {
    ...config,
    oauth:
      config.oauth === null
        ? new None()
        : new Some({ ...oauthDefaults, ...config.oauth }),
  };
}
