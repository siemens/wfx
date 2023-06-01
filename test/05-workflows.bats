#!/usr/bin/env bats
#
# SPDX-FileCopyrightText: 2023 Siemens AG
#
# SPDX-License-Identifier: Apache-2.0
#
# Author: Michael Adler <michael.adler@siemens.com>

. lib.sh

setup_file() {
    launch_wfx
    wait_wfx_running
}

teardown_file() {
    pkill wfx
}

@test "offline validation" {
    wfxctl workflow validate ../workflow/dau/wfx.workflow.dau.direct.yml
    wfxctl workflow validate ../workflow/dau/wfx.workflow.dau.direct.yml
}

@test "create kanban workflow" {
    wfxctl workflow create --filter=.transitions ../share/demo/kanban/wfx.workflow.kanban.yml
}

@test "create new kanban job" {
    ID=$(echo '{ "title": "Expose Job API" }' |
           wfxctl job create --workflow wfx.workflow.kanban \
             --client-id Dana \
             --filter='.id' --raw)

    # update jobu using wfxctl
    wfxctl job update-status \
        --actor=client \
        "--id=$ID" \
        --state=PROGRESS

    # update job again using curl
    curl -X PUT \
        "http://localhost:8080/api/wfx/v1/jobs/$ID/status" \
        -H 'Content-Type: application/json' \
        -H 'Accept: application/json' \
        -d '{"state":"PROGRESS"}'

    # update progress
    wfxctl job update-status \
        --actor=client \
        "--id=$ID" \
        --state=PROGRESS \
        --progress $((RANDOM % 100))

    # update state
    wfxctl job update-status \
        --actor=client \
        "--id=$ID" \
        --state=VALIDATE
}
