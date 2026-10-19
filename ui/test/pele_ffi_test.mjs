// SPDX-FileCopyrightText: 2026 Siemens AG
//
// SPDX-License-Identifier: Apache-2.0
//
// Author: Michael Adler <michael.adler@siemens.com>
import assert from "node:assert/strict";
import { renderDiagram } from "../src/pele.ffi.mjs";

const svg = renderDiagram(`stateDiagram-v2
  state "A & B" as S0
  S0 --> S0:CLIENT
  classDef active fill:#ADD8E6
  class S0 active`);
assert.match(svg, /<svg[^>]*class="pele pele-state"/);
assert.match(svg, /A &amp; B/);
assert.match(svg, /fill:#ADD8E6/);
