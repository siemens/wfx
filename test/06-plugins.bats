#!/usr/bin/env bats
#
# SPDX-FileCopyrightText: 2024 Siemens AG
#
# SPDX-License-Identifier: Apache-2.0
#
# Author: Michael Adler <michael.adler@siemens.com>

. lib.sh

teardown() {
    if pgrep -x wfx > /dev/null; then
        pkill -9 wfx
    fi
}

@test "plugins work and wfx shuts down cleanly" {
    wfx --log-level=debug --log-format=pretty \
        --storage=sqlite --storage-opt="file:wfx?mode=memory&cache=shared&_fk=1" \
        --client-plugins-dir ../example/plugin &
    wait_wfx_running 2
    run curl -s -o /dev/null -w "%{http_code}" "$BASEURL/workflows"
    assert_output "403"
    kill %1 # wfx, expect a clean shutdown
    local job_count=1
    for _ in {1..10}; do
      job_count=$(jobs -r | wc -l | tr -d " ")
      if [ "$job_count" -eq 0 ]; then
        break
      fi
      sleep 0.3
    done
    assert_equal "$job_count" 0
}

@test "wfx exits if plugin crashes" {
    wfx --log-level=debug --log-format=pretty \
        --storage=sqlite --storage-opt="file:wfx?mode=memory&cache=shared&_fk=1" \
        --client-plugins-dir ../example/plugin &
    WFX_PID=$!
    wait_wfx_running 2
    pkill -9 plugin

    for _ in {1..10}; do
        if [[ ! -d "/proc/$WFX_PID" ]]; then break; fi
        sleep 0.1
    done
    assert [ ! -e "/proc/$WFX_PID" ]
}
