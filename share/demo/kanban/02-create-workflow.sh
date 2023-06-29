#!/usr/bin/env bash
# SPDX-FileCopyrightText: 2023 Siemens AG
#
# SPDX-License-Identifier: Apache-2.0
#
# Author: Michael Adler <michael.adler@siemens.com>

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" &>/dev/null && pwd)"

# hide the evidence
clear

echo "# waiting for wfx..."
while true; do
    UP_COUNT=$(wfxctl health | grep -c up)
    if [[ "$UP_COUNT" -ge 2 ]]; then
        break
    fi
    sleep 0.2
done
wfxctl health
sleep 3

########################
# include the magic
########################
source "$SCRIPT_DIR"/demo.sh
. "$DM" -n
DEMO_COMMENT_COLOR=$WHITE
TYPE_SPEED=13
PROMPT_TIMEOUT=2

GIT_ROOT=$(git rev-parse --show-toplevel)

p "# create kanban workflow"
pei "wfxctl workflow create wfx.workflow.kanban.yml"

p "# visualization:"
cat <<EOF
 BACKLOG
    │
    ▼
   NEW ───────┐
    │         │
    ├◀─┐      │
    ▼  │      │
PROGRESS ─────┤
    │         │
    ▼         │
VALIDATE ─────┤
    │         │
    ▼         ▼
  DONE    DISCARDED
EOF
