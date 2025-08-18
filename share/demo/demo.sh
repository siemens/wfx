# SPDX-FileCopyrightText: 2025 Siemens AG
#
# SPDX-License-Identifier: Apache-2.0
#
# Author: Michael Adler <michael.adler@siemens.com>
DM=$(command -v demo-magic.sh || echo "")
[ "$DM" != "" ] || {
    DM=$(mktemp)
    curl -Ls -f -o "$DM" https://github.com/paxtonhare/demo-magic/raw/2a2f439c26a93286dc2adc6ef2a81755af83f36e/demo-magic.sh
    SHASUM=$(sha256sum "$DM" | awk '{print $1}')
    expected="ecaa937c89fe657668651610a62df6808daa3449248cdcbfe2be982b011dfb17"
    [ "$SHASUM" = "$expected" ] || {
        echo "checksum failed! actual: $SHASUM, expected: $expected"
        exit 1
    }
}

GIT_ROOT=$(git rev-parse --show-toplevel 2>/dev/null)
if [[ "$GIT_ROOT" != "" ]]; then
    # extend PATH, in case wfx and wfxctl are stored in $GIT_ROOT
    export PATH="$GIT_ROOT:$PATH"
fi

for binary in wfx wfxctl; do
    command -v "$binary" >/dev/null || {
        echo "FATAL: binary '$binary' not found"
        exit 1
    }
done

# vim: ft=sh
