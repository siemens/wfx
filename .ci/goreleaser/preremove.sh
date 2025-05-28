#!/bin/sh
# SPDX-FileCopyrightText: 2025 Siemens AG
#
# SPDX-License-Identifier: Apache-2.0

if command -v systemctl &> /dev/null; then
    systemctl disable wfx.service
fi

if [ -d /run/systemd/system ]; then
    systemctl stop wfx.service
fi
