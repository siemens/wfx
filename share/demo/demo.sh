#!/usr/bin/env bash
# SPDX-FileCopyrightText: 2023 Siemens AG
#
# SPDX-License-Identifier: Apache-2.0
#
# Author: Michael Adler <michael.adler@siemens.com>

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" &>/dev/null && pwd)"
DM=$(command -v demo-magic.sh)
[ "$DM" != "" ] || {
    DM=$(mktemp)
    echo "Saving demo-magic.sh to $DM"
    curl -Ls -o "$DM" https://github.com/paxtonhare/demo-magic/raw/a938137035b73105d09347a91f9fd5e9722b617a/demo-magic.sh
    SHASUM=$(sha256sum "$DM" | awk '{print $1}')
    expected="2f4f93fc8bc3c7708d51a12547c7f95024b4a49612191cdbc59233024d4b1cd3"
    [ "$SHASUM" = "$expected" ] || {
        echo "checksum failed! actual: $SHASUM, expected: $expected"
        exit 1
    }
}

########################
# include the magic
########################
. "$DM" -n
DEMO_COMMENT_COLOR=$WHITE
TYPE_SPEED=15

cd "$(git rev-parse --show-toplevel)"

# hide the evidence
clear

echo "Press enter to start..."
PROMPT_TIMEOUT=0
wait
clear
PROMPT_TIMEOUT=2

source "$SCRIPT_DIR"/wfx.inc.sh
source "$SCRIPT_DIR"/wfxctl.inc.sh
