#!/usr/bin/env bash
# SPDX-FileCopyrightText: 2024 Siemens AG
#
# SPDX-License-Identifier: Apache-2.0
#
# Author: Michael Adler <michael.adler@siemens.com>
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

. "$SCRIPT_DIR/versions.env"

echo "Installing mockery $MOCKERY_VERSION"
curl -Ls "https://github.com/vektra/mockery/releases/download/v${MOCKERY_VERSION}/mockery_${MOCKERY_VERSION}_Linux_x86_64.tar.gz" |
    tar --extract --gzip --directory=/usr/local/bin mockery
