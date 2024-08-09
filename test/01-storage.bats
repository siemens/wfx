#!/usr/bin/env bats
#
# SPDX-FileCopyrightText: 2024 Siemens AG
#
# SPDX-License-Identifier: Apache-2.0
#
# Author: Michael Adler <michael.adler@siemens.com>

. lib.sh

teardown() {
    pkill wfx
    rm -f wfx.db*
}

@test "Default storage" {
    cd "$BATS_TEST_TMPDIR"
    wfx &
    local count
    wait_wfx_running 2
    assert_file_exists wfx.db
}

@test "SQLite in-memory storage" {
    cd "$BATS_TEST_TMPDIR"
    wfx --storage sqlite --storage-opt "file:wfx?mode=memory&cache=shared&_fk=1" &
    local count
    wait_wfx_running 2
    assert_file_not_exists wfx.db
}

@test "PostgreSQL storage using CLI args" {
    wfx --storage postgres \
        --storage-opt "host=${PGHOST:-localhost} port=${PGPORT:-5432} user=${PGUSER:-wfx} password=${PGPASSWORD:-secret} database=${PGDATABASE:-wfx}" &
    wait_wfx_running 2
}

@test "PostgreSQL storage using env variables" {
    env PGHOST=${PGHOST:-localhost} \
        PGPORT=${PGPORT:-5432} \
        PGUSER=${PGUSER:-wfx} \
        PGPASSWORD=${PGPASSWORD:-secret} \
        PGDATABASE=${PGDATABASE:-wfx} \
        wfx --storage postgres &
    wait_wfx_running 2
}

@test "MySQL storage using CLI args" {
    wfx --storage mysql \
        --storage-opt "${MYSQL_USER:-root}:${MYSQL_PASSWORD:-root}@tcp(${MYSQL_HOST:-localhost}:${MYSQL_PORT:-3306})/${MYSQL_DATABASE:-wfx}" &
    wait_wfx_running 2
}
