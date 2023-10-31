#!/usr/bin/env bash
# SPDX-FileCopyrightText: 2023 Siemens AG
#
# SPDX-License-Identifier: Apache-2.0
#
# Author: Michael Adler <michael.adler@siemens.com>
set -euo pipefail

CLIENT_ID=""

usage() {
    echo "Usage: $0 -c CLIENT_ID"
}

while getopts "c:" option; do
    case "$option" in
        c)
            CLIENT_ID="$OPTARG"
            ;;
        *)
            usage
            exit 1
            ;;
    esac
done

if [[ "$CLIENT_ID" = "" ]]; then
    usage
    exit 1
fi

url="http://localhost:8080/download/example.ini"
preinstall="echo 'stopping service'; sleep 3"
postinstall="echo 'starting service'; sleep 3"
destination="/tmp/example.ini"

echo "Creating new job for client $CLIENT_ID"
printf '{ "url": "%s", "preinstall": "%s", "postinstall": "%s", "destination": "%s" }' "$url" "$preinstall" "$postinstall" "$destination" |
    wfxctl job create --workflow wfx.workflow.config.deployment \
        --client-id "$CLIENT_ID" \
        --filter='.id' --raw -
