#!/usr/bin/env bash
# SPDX-FileCopyrightText: 2023 Siemens AG
#
# SPDX-License-Identifier: Apache-2.0
#
# Author: Michael Adler <michael.adler@siemens.com>
set -euo pipefail
IN=$1

echo ">> Post-processing $IN"
dir=$(dirname "$IN")
fname=$(basename "$IN")

luafilter=$(readlink -f filters/fix-links-post.lua)
awkscript=$(readlink -f scripts/titleblock.awk)
pushd "$dir" >/dev/null
mv "$fname" "$fname".bak
OUT=$fname
pandoc --verbose --columns 120 -f gfm -t gfm --lua-filter "$luafilter" "$fname".bak |
    awk -f "$awkscript" >"$OUT"
rm "$fname".bak
popd >/dev/null
