// SPDX-FileCopyrightText: 2025 Siemens AG
//
// SPDX-License-Identifier: Apache-2.0
//
// Author: Michael Adler <michael.adler@siemens.com>
import mermaid from "mermaid";
mermaid.initialize({ startOnLoad: false, useMaxWidth: true });

export async function renderDiagram(graphDefinition) {
  const { svg } = await mermaid.render("graphDiv", graphDefinition);
  return svg;
}
