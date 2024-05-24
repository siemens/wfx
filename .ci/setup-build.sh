#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

apt-get update -q
apt-get install -q -y --no-install-recommends xz-utils zstd

"$SCRIPT_DIR/packages/just.sh"
"$SCRIPT_DIR/packages/zig.sh"
"$SCRIPT_DIR/packages/goreleaser.sh"
