#!/usr/bin/env bash
# SPDX-FileCopyrightText: 2026 Siemens AG
#
# SPDX-License-Identifier: Apache-2.0
#
# Author: Michael Adler <michael.adler@siemens.com>
set -euo pipefail

options=()
files=()
packages=()
declare -A seen_packages=()

while (($#)); do
    case $1 in
    --)
        shift
        files+=("$@")
        break
        ;;
    -*)
        if (($# < 2)) || [[ $2 == -* ]]; then
            echo "Error: Missing value for key $1" >&2
            exit 1
        fi
        options+=("$1" "$2")
        shift 2
        ;;
    *)
        files+=("$1")
        shift
        ;;
    esac
done

if ((${#files[@]} == 0)); then
    echo "Error: No file arguments provided" >&2
    exit 1
fi

for file in "${files[@]}"; do
    package=$(dirname -- "$file")
    if [[ ! ${seen_packages[$package]+_} ]]; then
        packages+=("./$package")
        seen_packages[$package]=1
    fi
done

staticcheck "${options[@]}" "${packages[@]}"
