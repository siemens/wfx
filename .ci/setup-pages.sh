#!/usr/bin/env bash
# SPDX-FileCopyrightText: 2025 Siemens AG
#
# SPDX-License-Identifier: Apache-2.0
#
# Author: Michael Adler <michael.adler@siemens.com>
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

apt-get update -q
apt-get install -q -y --no-install-recommends imagemagick librsvg2-bin

"$SCRIPT_DIR/packages/hugo.sh"
"$SCRIPT_DIR/packages/just.sh"
"$SCRIPT_DIR/packages/pandoc.sh"
"$SCRIPT_DIR/packages/lychee.sh"
