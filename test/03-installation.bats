#!/usr/bin/env bats
#
# SPDX-FileCopyrightText: 2023 Siemens AG
#
# SPDX-License-Identifier: Apache-2.0
#
# Author: Michael Adler <michael.adler@siemens.com>

. lib.sh

@test "DESTDIR and prefix are respected" {
    make -C.. DESTDIR="$BATS_TEST_TMPDIR/" prefix=local install
    assert_file_executable "$BATS_TEST_TMPDIR"/local/bin/wfx
    assert_file_executable "$BATS_TEST_TMPDIR"/local/bin/wfxctl
    assert_file_executable "$BATS_TEST_TMPDIR"/local/bin/wfx-viewer
    assert_file_executable "$BATS_TEST_TMPDIR"/local/bin/wfx-loadtest
}
