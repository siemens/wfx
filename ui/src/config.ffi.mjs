// SPDX-FileCopyrightText: 2025 Siemens AG
//
// SPDX-License-Identifier: Apache-2.0
//
// Author: Michael Adler <michael.adler@siemens.com>
export function loadConfig() {
  if (window.loadConfig) {
    return window.loadConfig();
  }
  console.log("Error: no config provided");
}
