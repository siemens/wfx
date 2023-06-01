#!/usr/bin/env bash
##############################################################################
# SPDX-FileCopyrightText: 2023 Siemens AG
#
# SPDX-License-Identifier: Apache-2.0
#
# Author: Michael Adler <michael.adler@siemens.com>

pkill wfx >/dev/null || true

# Put your stuff here
p "# usage:"
pe "wfx --help"
p "# tl;dr: storage options, tls config, timeouts"

pe "wfx >/dev/null &"
wait
p ""
