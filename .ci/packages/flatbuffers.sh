#!/usr/bin/env bash
# SPDX-FileCopyrightText: 2024 Siemens AG
#
# SPDX-License-Identifier: Apache-2.0
#
# Author: Michael Adler <michael.adler@siemens.com>
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

. "$SCRIPT_DIR/versions.env"

echo "Installing flatbuffers $FLATBUFFERS_VERSION"
curl -Lo /tmp/flatc.zip "https://github.com/google/flatbuffers/releases/download/v${FLATBUFFERS_VERSION}/Linux.flatc.binary.clang++-15.zip"
unzip /tmp/flatc.zip flatc -d /usr/local/bin/
