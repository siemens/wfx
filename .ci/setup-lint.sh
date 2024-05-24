#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

"$SCRIPT_DIR/packages/just.sh"
"$SCRIPT_DIR/packages/staticcheck.sh"
"$SCRIPT_DIR/packages/golangci-lint.sh"
