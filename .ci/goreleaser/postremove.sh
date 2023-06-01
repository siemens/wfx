#!/bin/sh
# SPDX-FileCopyrightText: 2023 Siemens AG
#
# SPDX-License-Identifier: Apache-2.0
if getent passwd wfx >/dev/null; then
    deluser wfx
fi

if getent group wfx >/dev/null; then
    delgroup wfx
fi

exit 0
