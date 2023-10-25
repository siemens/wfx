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

user="admin"
pass="secret"
timeout="5m"

echo "Creating new job for client $CLIENT_ID with timeout $timeout"
printf '{ "timeout": "%s", "credential": "%s:%s" }' "$timeout" "$user" "$pass" |
    wfxctl job create --workflow wfx.workflow.remote.access \
        --client-id "$CLIENT_ID" \
        --filter='.id' --raw -
