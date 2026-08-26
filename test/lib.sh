# SPDX-FileCopyrightText: 2023 Siemens AG
#
# SPDX-License-Identifier: Apache-2.0
#
# Author: Michael Adler <michael.adler@siemens.com>

load 'test_helper/bats-support/load'
load 'test_helper/bats-assert/load'
load 'test_helper/bats-file/load'

SOUTHBOUND_HOST="http://localhost:8080"
NORTHBOUND_HOST="http://localhost:8081"
API_BASE_PATH="/api/wfx/v1"

wait_wfx_running() {
    local expected=$1
    local count
    for _i in {1..20}; do
        count=0
        wfxctl --host "$SOUTHBOUND_HOST" health 2>/dev/null | grep -q $'wfx\tup' && count=$((count+1))
        wfxctl --host "$NORTHBOUND_HOST" health 2>/dev/null | grep -q $'wfx\tup' && count=$((count+1))
        if [[ "$count" -eq "$expected" ]]; then
            break
        fi
        sleep 0.5
    done
    assert_equal "$count" "$expected"
}

launch_wfx() {
    wfx --log-level=debug --log-format=pretty --storage=sqlite --storage-opt="file:wfx?mode=memory&cache=shared&_fk=1" &
}

# vim: ft=bash
