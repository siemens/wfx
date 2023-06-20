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

@test "default storage" {
    cd "$BATS_TEST_TMPDIR"
    wfx &
    local count
    count=$(wait_wfx_running)
    assert_equal "$count" 2
    assert_file_exists wfx.db
}

@test "SQLite storage" {
    cd "$BATS_TEST_TMPDIR"
    wfx --storage sqlite --storage-opt "file:wfx?mode=memory&cache=shared&_fk=1" &
    local count
    count=$(wait_wfx_running)
    assert_equal "$count" 2
    assert_file_not_exists wfx.db
}

@test "PostgreSQL storage using cli args" {
    wfx --storage postgres \
        --storage-opt "host=${PGHOST:-localhost} port=${PGPORT:-5432} user=${PGUSER:-wfx} password=${PGPASSWORD:-secret} database=${PGDATABASE:-wfx}" &
    local count
    count=$(wait_wfx_running)
    assert_equal "$count" 2
}

@test "PostgreSQL storage using env variables" {
    env PGHOST=${PGHOST:-localhost} \
        PGPORT=${PGPORT:-5432} \
        PGUSER=${PGUSER:-wfx} \
        PGPASSWORD=${PGPASSWORD:-secret} \
        PGDATABASE=${PGDATABASE:-wfx} \
        wfx --storage postgres &
    local count
    count=$(wait_wfx_running)
    assert_equal "$count" 2
}

@test "MySQL storage" {
    wfx --storage mysql \
        --storage-opt "${MYSQL_USER:-root}:${MYSQL_PASSWORD:-root}@tcp(${MYSQL_HOST:-localhost}:${MYSQL_PORT:-3306})/${MYSQL_DATABASE:-wfx}" &
    local count
    count=$(wait_wfx_running)
    assert_equal "$count" 2
}

@test "Fileserver is off by default" {
    launch_wfx
    wait_wfx_running
    run curl -s http://localhost:8080/download
    assert_output '{"code":404,"message":"path /download was not found"}'
}

@test "Fileserver is on if --simple-fileserver is non-empty" {
    mkdir -p "$BATS_TEST_TMPDIR/download"
    echo "hello world" >"$BATS_TEST_TMPDIR"/download/hello
    wfx --storage sqlite \
        --storage-opt "file:wfx?mode=memory&cache=shared&_fk=1" \
        --simple-fileserver "$BATS_TEST_TMPDIR/download" &
    wait_wfx_running
    run curl -s -f http://localhost:8080/download/hello
    assert_output 'hello world'
}

@test "Configuration via env variables" {
    cd "$BATS_TEST_TMPDIR"
    export WFX_LOG_LEVEL=debug
    wfx 1>wfx.log 2>&1 &
    sleep 0.5
    # this message has level debug
    local msg="Setting up persistence storage"
    local count
    count=$(grep -c "$msg" wfx.log)
    assert_equal "$count" 1
}

@test "Configuration via yaml file" {
    cd "$BATS_TEST_TMPDIR"
    echo "log-level: debug" > wfx.yml
    wfx 1>wfx.log 2>&1 &
    sleep 0.5
    # this message has level debug
    local msg="Setting up persistence storage"
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
