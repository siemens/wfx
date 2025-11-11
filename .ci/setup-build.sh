#!/usr/bin/env bash
# SPDX-FileCopyrightText: 2025 Siemens AG
#
# SPDX-License-Identifier: Apache-2.0
#
# Author: Michael Adler <michael.adler@siemens.com>
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

apt-get update -q
apt-get install -q -y --no-install-recommends xz-utils zstd

"$SCRIPT_DIR/packages/just.sh"
"$SCRIPT_DIR/packages/zig.sh"
"$SCRIPT_DIR/packages/goreleaser.sh"
"$SCRIPT_DIR/packages/syft.sh"

# Make zcc wrapper executable
chmod +x "$SCRIPT_DIR/zcc"

# Ensure Go dependencies are up to date
go mod tidy
