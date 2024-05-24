#!/usr/bin/env bash
# SPDX-FileCopyrightText: 2024 Siemens AG
#
# SPDX-License-Identifier: Apache-2.0
#
# Author: Michael Adler <michael.adler@siemens.com>
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

. "$SCRIPT_DIR/versions.env"

echo "Installing gofumpt $GOFUMPT_VERSION"
curl -Lso /usr/local/bin/gofumpt "https://github.com/mvdan/gofumpt/releases/download/v${GOFUMPT_VERSION}/gofumpt_v${GOFUMPT_VERSION}_linux_amd64"
chmod 0755 /usr/local/bin/gofumpt
