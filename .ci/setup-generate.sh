#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

apt-get update -q
apt-get install -q -y --no-install-recommends python3-yaml git-lfs apt-transport-https unzip

"$SCRIPT_DIR/packages/just.sh"
"$SCRIPT_DIR/packages/flatbuffers.sh"
