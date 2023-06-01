#!/usr/bin/env bats
#
# SPDX-FileCopyrightText: 2023 Siemens AG
#
# SPDX-License-Identifier: Apache-2.0
#
# Author: Michael Adler <michael.adler@siemens.com>

. lib.sh

PRIVATE_KEY="$BATS_FILE_TMPDIR/private.pem"
PUBLIC_KEY="$BATS_FILE_TMPDIR/public.pem"

setup_file() {
  openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout "$PRIVATE_KEY" -out "$PUBLIC_KEY" -subj "/CN=example.com"
}

teardown_file() {
  pkill wfx
  rm -f "$PUBLIC_KEY" "$PRIVATE_KEY"
}

@test "TLS only" {
  cd "$BATS_TEST_TMPDIR"
  wfx --scheme=https \
    --tls-certificate="$PUBLIC_KEY" \
    --tls-key="$PRIVATE_KEY" &
    local count
    count=$(wait_wfx_running)
    assert_equal "$count" 2
}

@test "Unix-Domain Sockets" {
  cd "$BATS_TEST_TMPDIR"
  wfx --scheme unix \
    --client-unix-socket "$BATS_TEST_TMPDIR/wfx-client.sock" \
    --mgmt-unix-socket "$BATS_TEST_TMPDIR/wfx-mgmt.sock" &
  local count
  count=$(wait_wfx_running)
  assert_equal "$count" 2
}

@test "TLS mixed-mode" {
  cd "$BATS_TEST_TMPDIR"
  wfx --scheme=http,https \
    --client-host=localhost \
    --client-tls-host=0.0.0.0 \
    --tls-certificate="$PUBLIC_KEY" \
    --tls-key="$PRIVATE_KEY" &
    local count
    count=$(wait_wfx_running 4)
    assert_equal "$count" 4
}
