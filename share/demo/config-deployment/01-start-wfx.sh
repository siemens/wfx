#!/usr/bin/env bash
# SPDX-FileCopyrightText: 2023 Siemens AG
#
# SPDX-License-Identifier: Apache-2.0
#
# Author: Michael Adler <michael.adler@siemens.com>

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" &>/dev/null && pwd)"
source "$SCRIPT_DIR/../demo.sh"

########################
# include the magic
########################
. "$DM" -n
DEMO_COMMENT_COLOR=$WHITE
TYPE_SPEED=15

GIT_ROOT=$(git rev-parse --show-toplevel)

# hide the evidence
clear

PROMPT_TIMEOUT=2

pe "wfx --storage sqlite --storage-opt 'file:wfx?mode=memory&cache=shared&_fk=1' --simple-fileserver files"
