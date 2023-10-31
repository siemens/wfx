#!/usr/bin/env bash
# SPDX-FileCopyrightText: 2023 Siemens AG
#
# SPDX-License-Identifier: Apache-2.0
#
# Author: Michael Adler <michael.adler@siemens.com>

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" &>/dev/null && pwd)"

# hide the evidence
clear

echo "# waiting for workflow..."
while true; do
    RC=0
    wfxctl workflow get --name=wfx.workflow.config.deployment 1>/dev/null 2>&1 || RC=$?
    if [[ $RC -eq 0 ]]; then
        break
    fi
    sleep 3
done
echo "# found workflow"

########################
# include the magic
########################
source "$SCRIPT_DIR/../demo.sh"
. "$DM" -n
DEMO_COMMENT_COLOR=$WHITE
TYPE_SPEED=15
PROMPT_TIMEOUT=2

while true; do
    pe "jq <definition.json"
    sleep 3
    pe "wfxctl job create --workflow wfx.workflow.config.deployment --client-id foo - <definition.json >/dev/null"
    sleep 10
    clear
done
