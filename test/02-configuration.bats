#!/usr/bin/env bats
#
# SPDX-FileCopyrightText: 2023 Siemens AG
#
# SPDX-License-Identifier: Apache-2.0
#
# Author: Michael Adler <michael.adler@siemens.com>

. lib.sh

teardown() {
    pkill wfx
    rm -f wfx.db*
}

@test "Configuration via yaml file" {
    cd "$BATS_TEST_TMPDIR"
    echo "log-level: debug" > wfx.yml
    wfx 1>wfx.log 2>&1 &
    sleep 0.5
    # this message has level debug
    local msg="Setting up persistent storage"
    local count
    count=$(grep -c "$msg" wfx.log)
    assert_equal "$count" 1
}

@test "Command-line parameter has priority over env variables" {
    cd "$BATS_TEST_TMPDIR"
    export WFX_LOG_LEVEL=debug
    wfx --log-level info 1>wfx.log 2>&1 &
    sleep 0.5
    # this message has level debug
    local msg="Setting up persistence storage"
    if grep -q "$msg" wfx.log; then
      local count
      count=$(grep -c "$msg" wfx.log)
      assert_equal "$count" 0
    fi
}

@test "Command-line parameter has priority over config file" {
    cd "$BATS_TEST_TMPDIR"
    echo "log-level: debug" > wfx.yml
    wfx --log-level info 1>wfx.log 2>&1 &
    sleep 0.5
    # this message has level debug
    local msg="Setting up persistence storage"
    if grep -q "$msg" wfx.log; then
      local count
      count=$(grep -c "$msg" wfx.log)
      assert_equal "$count" 0
    fi
}

@test "Env variable has priority over config file" {
    cd "$BATS_TEST_TMPDIR"
    echo "log-level: debug" > wfx.yml
    export WFX_LOG_LEVEL=info
    wfx 1>wfx.log 2>&1 &
    sleep 0.5
    # this message has level debug
    local msg="Setting up persistence storage"
    if grep -q "$msg" wfx.log; then
      local count
      count=$(grep -c "$msg" wfx.log)
      assert_equal "$count" 0
    fi
}
