#!/bin/sh
# SPDX-FileCopyrightText: 2023 Siemens AG
#
# SPDX-License-Identifier: Apache-2.0
#
# Author: Michael Adler <michael.adler@siemens.com>
set -eux
VERSION="1.13.0"
curl -Ls "https://github.com/casey/just/releases/download/$VERSION/just-$VERSION-x86_64-unknown-linux-musl.tar.gz" |
    tar --no-same-owner -C /usr/local/bin/ -xzv -f - just
just --version
