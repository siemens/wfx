#!/usr/bin/env bats
#
# SPDX-FileCopyrightText: 2026 Siemens AG
#
# SPDX-License-Identifier: Apache-2.0
#
# Author: Michael Adler <michael.adler@siemens.com>

. lib.sh

setup_file() {
    launch_wfx
    wait_wfx_running 2
    wfxctl workflow create ../workflow/dau/wfx.workflow.dau.direct.yml >/dev/null
    for client_id in valid spoof; do
        echo '{}' | wfxctl job create \
            --client-id "$client_id" \
            --workflow wfx.workflow.dau.direct - >/dev/null
    done
}

teardown_file() {
    pkill wfx
}

@test "X-Client-Id validates southbound job queries" {
    local host="http://localhost:8080"

    run wfxctl --host "$host" --header "X-Client-Id: valid" \
        --filter='.content[].clientId' --raw job query
    assert_success
    assert_output "valid"

    run curl -s -o /dev/null -w "%{http_code}" \
        -H "X-Client-Id: valid" "$BASEURL/jobs"
    assert_output "200"

    run wfxctl --host "$host" --header "X-Client-Id: valid" \
        job query --client-id spoof
    assert_failure
    assert_output --partial "wfx.clientIDMismatch"

    run curl -s -o /dev/null -w "%{http_code}" \
        -H "X-Client-Id: valid" "$BASEURL/jobs?clientId=spoof"
    assert_output "400"

    local job_id
    job_id=$(wfxctl --host "$host" --header "X-Client-Id: valid" \
        --filter='.content[0].id' --raw job query)

    run wfxctl --host "$host" --header "X-Client-Id: valid" \
        --filter='.clientId' --raw job get --id "$job_id"
    assert_success
    assert_output "valid"

    run curl -s -o /dev/null -w "%{http_code}" \
        -H "X-Client-Id: valid" "$BASEURL/jobs/$job_id"
    assert_output "200"

    run wfxctl --host "$host" --header "X-Client-Id: spoof" \
        job get --id "$job_id"
    assert_failure
    assert_output --partial "wfx.jobNotFound"

    run curl -s -o /dev/null -w "%{http_code}" \
        -H "X-Client-Id: spoof" "$BASEURL/jobs/$job_id"
    assert_output "404"

    run wfxctl --host "$host" --header "X-Client-Id: valid" \
        job update-status --id "$job_id" --client-id valid --state INSTALLING
    assert_success

    run curl -s -o /dev/null -w "%{http_code}" -X PUT \
        -H "Content-Type: application/json" -H "X-Client-Id: valid" \
        -d '{"clientId":"valid","state":"INSTALLING"}' \
        "$BASEURL/jobs/$job_id/status"
    assert_output "200"

    run wfxctl --host "$host" --header "X-Client-Id: spoof" \
        job update-status --id "$job_id" --client-id valid --state INSTALLED
    assert_failure
    assert_output --partial "wfx.jobNotFound"

    run curl -s -o /dev/null -w "%{http_code}" -X PUT \
        -H "Content-Type: application/json" -H "X-Client-Id: spoof" \
        -d '{"clientId":"valid","state":"INSTALLED"}' \
        "$BASEURL/jobs/$job_id/status"
    assert_output "404"
}
