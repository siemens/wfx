#!/bin/sh
# SPDX-FileCopyrightText: 2023 Siemens AG
#
# SPDX-License-Identifier: Apache-2.0
#
# Author: Michael Adler <michael.adler@siemens.com>
set -eux
VERSION="v0.30.4"
echo "Installing swagger $VERSION"
curl -Ls -o /usr/local/bin/swagger "https://github.com/go-swagger/go-swagger/releases/download/${VERSION}/swagger_linux_amd64"
chmod +x /usr/local/bin/swagger
swagger version
