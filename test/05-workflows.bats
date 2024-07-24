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
    wait_wfx_running 2
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
             --filter='.id' --raw -)

    # update job using wfxctl
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

    # update state to VALIDATE
    wfxctl job update-status \
        --actor=client \
        "--id=$ID" \
        --state=VALIDATE

    # update state to DONE
    wfxctl job update-status \
        --actor=client \
        "--id=$ID" \
        --state=DONE
}

@test "query jobs by group" {
    # NOTE: this test assumes a job whose state is in the CLOSED group;
    # such a job is created by the previous test

    # Create a second job in NEW state
    wfxctl job create --workflow wfx.workflow.kanban \
        --client-id Dana \
        --filter='.id'

    # Query all jobs
    jobs_no_filter=$(wfxctl job query --filter ".content | length")
    assert_equal "$jobs_no_filter" 2

    # Query jobs only in group=OPEN states
    jobs_open=$(wfxctl job query --group OPEN --filter ".content | length")
    assert_equal "$jobs_open" 1

    # Query jobs in OPEN or CLOSED states
    jobs_open_closed=$(wfxctl job query --group OPEN,CLOSED --filter ".content | length")
    assert_equal "$jobs_open_closed" 2

    # Use alternative syntax
    jobs_open_closed=$(wfxctl job query --group OPEN --group CLOSED --filter ".content | length")
    assert_equal "$jobs_open_closed" 2
}

@test "Subscribe to job events" {
    cd "$BATS_TEST_TMPDIR"
    ID=$(echo '{ "title": "Expose Job API" }' |
        wfxctl job create --workflow wfx.workflow.kanban \
            --client-id Dana \
            --filter='.id' --raw - 2>/dev/null)

    curl -s --no-buffer "localhost:8080/api/wfx/v1/jobs/events?jobIds=$ID&tags=bats" > curl.out &
    sleep 1
    for state in PROGRESS VALIDATE DONE; do
        wfxctl job update-status \
            --actor=client \
            --id "$ID" \
            --state "$state" 1>/dev/null 2>&1
    done
    for i in {1..30}; do
        if grep -q DONE curl.out; then
            break
        fi
        sleep 1
    done

    assert_file_contains curl.out '"state":"PROGRESS"'
    assert_file_contains curl.out '"state":"VALIDATE"'
    assert_file_contains curl.out '"state":"DONE"'
}
