// SPDX-FileCopyrightText: 2026 Siemens AG
//
// SPDX-License-Identifier: Apache-2.0
//
// Author: Michael Adler <michael.adler@siemens.com>
import gleam/option.{type Option}

pub type OAuthConfig {
  OAuthConfig(issuer_url: String, client_id: String, scope: String)
}

pub type Config {
  Config(wfx_url: String, base_path: String, oauth: Option(OAuthConfig))
}

@external(javascript, "./config.ffi.mjs", "loadConfig")
pub fn load_config() -> Config
