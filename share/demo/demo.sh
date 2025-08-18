# SPDX-FileCopyrightText: 2023 Siemens AG
#
# SPDX-License-Identifier: Apache-2.0
#
# Author: Michael Adler <michael.adler@siemens.com>
DM=$(command -v demo-magic.sh)
[ "$DM" != "" ] || {
    DM=$(mktemp)
    echo "Saving demo-magic.sh to $DM"
    curl -Ls -o "$DM" https://github.com/paxtonhare/demo-magic/raw/2a2f439c26a93286dc2adc6ef2a81755af83f36e/demo-magic.sh
    SHASUM=$(sha256sum "$DM" | awk '{print $1}')
    expected="ecaa937c89fe657668651610a62df6808daa3449248cdcbfe2be982b011dfb17"
    [ "$SHASUM" = "$expected" ] || {
        echo "checksum failed! actual: $SHASUM, expected: $expected"
        exit 1
    }
}

# vim: ft=sh
