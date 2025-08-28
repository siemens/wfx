#!/usr/bin/env bash
# SPDX-FileCopyrightText: 2025 Siemens AG
#
# SPDX-License-Identifier: Apache-2.0
#
# Author: Michael Adler <michael.adler@siemens.com>
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

. "$SCRIPT_DIR/versions.env"

command -v npm >/dev/null || {
    apt-get update -q
    apt-get install -q -y --no-install-recommends npm
}

npm install -g -q "prettier@$PRETTIER_VERSION"
