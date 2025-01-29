#!/usr/bin/env python3
# SPDX-FileCopyrightText: 2025 Siemens AG
#
# SPDX-License-Identifier: Apache-2.0
#
# Author: Michael Adler <michael.adler@siemens.com>
import subprocess
import json
import sys

result = subprocess.run(
    ["lychee", "--format=json", "."], capture_output=True, text=True
)
output = json.loads(result.stdout)
has_broken_urls = False
for fname, errs in output["error_map"].items():
    has_broken_urls = True
    with open(fname, "r") as f:
        lines = f.readlines()
    for entry in errs:
        for row, line in enumerate(lines, start=1):
            url = entry["url"]
            if url in line:
                print(f"::warning file={fname},line={row},col=1::Broken URL: {url}")

if has_broken_urls:
    sys.exit(1)
