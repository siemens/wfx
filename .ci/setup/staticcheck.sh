#!/bin/sh
# SPDX-FileCopyrightText: 2023 Siemens AG
#
# SPDX-License-Identifier: Apache-2.0
#
# Author: Michael Adler <michael.adler@siemens.com>
set -eux
VERSION="2023.1.3"
echo "Installing staticcheck $VERSION"
curl -Ls "https://github.com/dominikh/go-tools/releases/download/$VERSION/staticcheck_linux_amd64.tar.gz" |
    tar --no-same-owner -C /usr/local/bin/ --strip-components=1 -xzv -f - staticcheck/staticcheck
