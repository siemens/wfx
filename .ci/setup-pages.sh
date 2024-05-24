#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

apt-get update -q
apt-get install -q -y --no-install-recommends npm imagemagick librsvg2-bin

"$SCRIPT_DIR/packages/hugo.sh"
"$SCRIPT_DIR/packages/just.sh"
"$SCRIPT_DIR/packages/pandoc.sh"
"$SCRIPT_DIR/packages/markdown-link-check.sh"
