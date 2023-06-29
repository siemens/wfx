#!/usr/bin/env bash
# SPDX-FileCopyrightText: 2023 Siemens AG
#
# SPDX-License-Identifier: Apache-2.0
#
# Author: Michael Adler <michael.adler@siemens.com>

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" &>/dev/null && pwd)"

# hide the evidence
clear

echo "# waiting for kanban workflow..."
while true; do
    RC=0
    wfxctl workflow get --name=wfx.workflow.kanban 1>/dev/null 2>&1 || RC=$?
    if [[ $RC -eq 0 ]]; then
        break
    fi
    sleep 3
done
echo "# found kanban workflow"

########################
# include the magic
########################
source "$SCRIPT_DIR"/demo.sh
. "$DM" -n
DEMO_COMMENT_COLOR=$WHITE
TYPE_SPEED=15
PROMPT_TIMEOUT=2

CLIENTS=("$@")
TASKID=1
while true; do
    for client in "${CLIENTS[@]}"; do
        clear
        p "# create a task for developer $client"
        p "echo '{ \"title\": \"Task $TASKID\" }' | wfxctl job create --workflow wfx.workflow.kanban --client-id $client --filter 'del(.workflow)'"
        echo "{ \"title\": \"Task $TASKID\" }" | wfxctl job create --workflow wfx.workflow.kanban --client-id "$client" --filter 'del(.workflow)' -
        sleep 3
        TASKID=$((TASKID + 1))
    done
done
