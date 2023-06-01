#!/usr/bin/env bash
# SPDX-FileCopyrightText: 2023 Siemens AG
#
# SPDX-License-Identifier: Apache-2.0
#
# Author: Michael Adler <michael.adler@siemens.com>
set -euo pipefail

IN=$1
OUT=$2

mkdir -p "$(dirname "$OUT")"
touch "$OUT"
OUT=$(readlink -f "$OUT")
rm -f "$OUT"

echo ">> Converting: $IN -> $OUT"

mkdir -p "$(dirname "$OUT")"
fullname=$(readlink -f "$IN")
fname=$(basename "$fullname")
luafilter=$(readlink -f filters/fix-links-pre.lua)
dname=$(dirname "$fullname")

set -x
cd "$dname"
pandoc --verbose --columns 120 -f gfm -t gfm --lua-filter "$luafilter" -o "$OUT" <"$fname"
