#!/usr/bin/env bash
# SPDX-FileCopyrightText: 2024 Siemens AG
#
# SPDX-License-Identifier: Apache-2.0
#
# Author: Michael Adler <michael.adler@siemens.com>
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

. "$SCRIPT_DIR/versions.env"

echo "Installing go-swagger $SWAGGER_VERSION"
curl -Lo /usr/local/bin/swagger "https://github.com/go-swagger/go-swagger/releases/download/v${SWAGGER_VERSION}/swagger_linux_amd64"
chmod 0755 /usr/local/bin/swagger
