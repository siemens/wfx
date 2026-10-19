// SPDX-FileCopyrightText: 2026 Siemens AG
//
// SPDX-License-Identifier: Apache-2.0
//
// Author: Michael Adler <michael.adler@siemens.com>
import { register, render } from "pele/core";
import state from "pele/diagrams/state";

register(state);

export function renderDiagram(graphDefinition) {
  return render(graphDefinition).svg;
}
