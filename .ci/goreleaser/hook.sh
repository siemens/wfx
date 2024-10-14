#!/usr/bin/env bash
# SPDX-FileCopyrightText: 2023 Siemens AG
#
# SPDX-License-Identifier: Apache-2.0
#
# Author: Michael Adler <michael.adler@siemens.com>
set -euxo pipefail

# save stdout and stderr
exec 3>&1 4>&2

# restore stdout and stderr on exit
trap 'exec 2>&4 1>&3' 0 1 2 3

# redirect stdout and stderr to a file
exec 1>>dist/hook.log 2>&1

# full path to the binary
BINARY=$1
FNAME=$(basename "$BINARY")
DNAME=$(dirname "$BINARY")

"$BINARY" completion bash >"$DNAME/$FNAME".bash
"$BINARY" completion fish >"$DNAME/$FNAME".fish
"$BINARY" completion zsh >"${DNAME}/_${FNAME}"

"$BINARY" man --dir "${DNAME}"
find "${DNAME}" -name "*.1" -exec gzip -9 {} \;
