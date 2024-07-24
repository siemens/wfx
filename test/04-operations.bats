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
}

@test "Socket-based activation" {
  cd "$BATS_TEST_TMPDIR"
  systemd-socket-activate \
    --listen "$BATS_TEST_TMPDIR/wfx-client.sock" \
    --listen "$BATS_TEST_TMPDIR/wfx-mgmt.sock" \
    wfx --scheme unix &
  local RC
  local i=0
  while [[ $i -lt 30 ]]; do
    RC=0
    wfxctl --client-unix-socket "$BATS_TEST_TMPDIR/wfx-client.sock" version || RC=$?
    if [[ $RC -eq 0 ]]; then
        break
    fi
    sleep 1
    i=$((i+1))
  done
  assert_equal "$RC" 0
}
