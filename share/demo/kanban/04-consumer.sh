#!/usr/bin/env bash
# SPDX-FileCopyrightText: 2023 Siemens AG
#
# SPDX-License-Identifier: Apache-2.0
#
# Author: Michael Adler <michael.adler@siemens.com>
[[ $# -gt 0 ]] || {
    echo "Usage: $0 CLIENT_ID"
    exit 1
}

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" &>/dev/null && pwd)"

echo "# waiting for wfx..."
while true; do
    UP_COUNT=$(wfxctl health | grep -c up)
    if [[ "$UP_COUNT" -ge 2 ]]; then
        break
    fi
    sleep 0.2
done

########################
# include the magic
########################
source "$SCRIPT_DIR"/demo.sh
. "$DM" -n
DEMO_COMMENT_COLOR=$WHITE
TYPE_SPEED=20
PROMPT_TIMEOUT=2

# hide the evidence
clear

CLIENT_ID=$1

while true; do
    clear
    echo "# polling for new task for developer '$CLIENT_ID'"
    JOB_ID=
    while [[ "$JOB_ID" = "" ]]; do
        JOB_ID=$(wfxctl job query --limit 1 --client-id "$CLIENT_ID" --state=NEW --filter '.content.[].id' --raw)
        sleep 1
    done
    p "wfxctl job query --limit 1 --client-id $CLIENT_ID --state=NEW --filter '.content.[].id' --raw"
    p "# found job with id $JOB_ID"
    for state in PROGRESS VALIDATE DONE; do
        pe "wfxctl job update-status --id=$JOB_ID --state=$state"
        sleep 4
    done
    p "# task is done"
    sleep 10
done
