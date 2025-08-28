#!/usr/bin/env bash
# SPDX-FileCopyrightText: 2025 Siemens AG
#
# SPDX-License-Identifier: Apache-2.0
#
# Author: Michael Adler <michael.adler@siemens.com>
set -euo pipefail

apt-get update -q
apt-get install -q -y --no-install-recommends bats jq systemd
