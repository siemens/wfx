// SPDX-FileCopyrightText: 2025 Siemens AG
//
// SPDX-License-Identifier: Apache-2.0
//
// Author: Michael Adler <michael.adler@siemens.com>
pub type Config {
  Config(wfx_url: String, base_path: String)
}

@external(javascript, "./config.ffi.mjs", "loadConfig")
pub fn load_config() -> Config
