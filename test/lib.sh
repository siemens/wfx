# SPDX-FileCopyrightText: 2023 Siemens AG
#
# SPDX-License-Identifier: Apache-2.0
#
# Author: Michael Adler <michael.adler@siemens.com>

load 'test_helper/bats-support/load'
load 'test_helper/bats-assert/load'
load 'test_helper/bats-file/load'

# shellcheck disable=SC2034
BASEURL="http://localhost:8080/api/wfx/v1"

wait_wfx_running() {
    local expected=$1
    local count
    for _i in {1..20}; do
        count=$(wfxctl health 2>/dev/null || echo "")
        count=$(echo "$count" | grep -c up || true)
        if [[ "${count:-0}" -eq "$expected" ]]; then
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
