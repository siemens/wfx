// SPDX-FileCopyrightText: 2025 Siemens AG
//
// SPDX-License-Identifier: Apache-2.0
//
// Author: Michael Adler <michael.adler@siemens.com>
@external(javascript, "./highlightjs.ffi.mjs", "highlightCode")
pub fn highlight_code(code: String, lang: String) -> String
