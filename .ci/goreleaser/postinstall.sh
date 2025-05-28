#!/bin/sh
# SPDX-FileCopyrightText: 2023 Siemens AG
#
# SPDX-License-Identifier: Apache-2.0
set -eu

if ! getent passwd wfx >/dev/null; then
    adduser --system --group --shell /usr/sbin/nologin --disabled-login \
        --home /var/lib/wfx --no-create-home wfx
fi

mkdir -p /var/lib/wfx
chown -R wfx:wfx /var/lib/wfx

if command -v systemctl &> /dev/null; then
    systemctl enable wfx.service
fi

if [ -d /run/systemd/system ]; then
    systemctl --system daemon-reload
    systemctl start wfx.service
fi