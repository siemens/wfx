#!/usr/bin/env -S awk -f
# SPDX-FileCopyrightText: 2023 Siemens AG
#
# SPDX-License-Identifier: Apache-2.0
#
# Author: Michael Adler <michael.adler@siemens.com>

BEGIN { found_header = 0 }

NR == 1 && /^\+\+\+/ {    # If it's the first line and it starts with "+++"
    found_header = 1      # the document has already been processed, so do nothing
}

/^# / && !found_header {
    gsub(/^# /, "") # remove the hash and space
    printf "+++\ntitle = \"%s\"\n+++\n", $0 # print the transformed header
    found_header = 1
    next # skip to the next line
}

{ print } # print all lines except the transformed header
