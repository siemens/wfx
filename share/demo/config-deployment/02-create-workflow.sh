#!/usr/bin/env bash
# SPDX-FileCopyrightText: 2023 Siemens AG
#
# SPDX-License-Identifier: Apache-2.0
#
# Author: Michael Adler <michael.adler@siemens.com>

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" &>/dev/null && pwd)"

# hide the evidence
clear

while true; do
    if pgrep config-deployer >/dev/null; then
        break
    fi
    sleep 1
done
sleep 1

########################
# include the magic
########################
source "$SCRIPT_DIR/../demo.sh"
. "$DM" -n
DEMO_COMMENT_COLOR=$WHITE
TYPE_SPEED=13
PROMPT_TIMEOUT=2

p "# create deployment workflow"
pei "wfxctl workflow create wfx.workflow.config.deployment.yml"
