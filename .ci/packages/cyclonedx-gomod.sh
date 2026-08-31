#!/usr/bin/env bash
# SPDX-FileCopyrightText: 2026 Siemens AG
#
# SPDX-License-Identifier: Apache-2.0
#
# Author: Michael Adler <michael.adler@siemens.com>
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

. "$SCRIPT_DIR/versions.env"

echo "Installing cyclonedx-gomod $CYCLONEDX_GOMOD_VERSION"
curl -Ls "https://github.com/CycloneDX/cyclonedx-gomod/releases/download/v${CYCLONEDX_GOMOD_VERSION}/cyclonedx-gomod_${CYCLONEDX_GOMOD_VERSION}_linux_amd64.tar.gz" |
    tar --extract --gzip --directory=/usr/local/bin cyclonedx-gomod
chmod 0755 /usr/local/bin/cyclonedx-gomod
