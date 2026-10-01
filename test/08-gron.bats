#!/usr/bin/env bats
#
# SPDX-FileCopyrightText: 2026 Siemens AG
#
# SPDX-License-Identifier: Apache-2.0

. lib.sh

setup_file() {
    launch_wfx
    wait_wfx_running 2
    wfxctl workflow create ../workflow/dau/wfx.workflow.dau.direct.yml >/dev/null
}

teardown_file() {
    pkill wfx
}

create_job() {
    echo '{}' | wfxctl job create --workflow wfx.workflow.dau.direct \
        --client-id gron --filter='.id' --raw -
}

@test "HTTP API returns gron when requested" {
    local id
    id=$(create_job)

    run curl -sSf -D - "$SOUTHBOUND_HOST$API_BASE_PATH/jobs/$id/status" \
        -H 'Accept: application/gron'
    assert_success
    assert_output --partial 'Content-Type: application/gron'
    assert_output --partial 'json.state = "INSTALL";'
}

@test "wfxctl --format gron returns gron" {
    local id
    id=$(create_job)

    run wfxctl job get-status --id "$id" --format gron
    assert_success
    assert_output --partial 'json.state = "INSTALL";'
}

@test "gron --ungron matches JSON response" {
    local id json gron_response
    id=$(create_job)

    run wfxctl job get-status --id "$id" --format json
    assert_success
    json=$output

    run wfxctl job get-status --id "$id" --format gron
    assert_success
    gron_response=$output

    run gron --ungron <<<"$gron_response"
    assert_success
    assert_equal "$(jq -S . <<<"$json")" "$(jq -S . <<<"$output")"
}
