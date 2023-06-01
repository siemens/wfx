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

@test "check /swagger.json available" {
    curl -s -f "$BASEURL/swagger.json"
}

@test "response filters" {
    create_kanban_workflow
    local id
    id=$(create_kanban_job)
    run curl -s -f "$BASEURL/jobs/$id/status" -H "X-Response-Filter: .state"
    assert_output '"NEW"'
}

@test "check health" {
    curl http://localhost:8080/health
}

@test "check version" {
    run sh -c "curl -s -f http://localhost:8080/version | jq -r .apiVersion"
    assert_output 'v1'
}
