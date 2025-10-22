// SPDX-FileCopyrightText: 2025 Siemens AG
//
// SPDX-License-Identifier: Apache-2.0
//
// Author: Michael Adler <michael.adler@siemens.com>
import hljs from "highlight.js";
import json from "highlight.js/lib/languages/json";

hljs.registerLanguage("json", json);

export function highlightCode(code, lang) {
  const highlightedCode = hljs.highlight(code, { language: lang }).value;
  return highlightedCode;
}
