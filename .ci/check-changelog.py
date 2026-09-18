#!/usr/bin/env python3
# SPDX-FileCopyrightText: 2026 Siemens AG
#
# SPDX-License-Identifier: Apache-2.0
#
# Author: Michael Adler <michael.adler@siemens.com>
import datetime
import re
import sys
from pathlib import Path

import tree_sitter_markdown
from tree_sitter import Language, Parser, Query, QueryCursor

HEADING_QUERY = """
[
  (atx_heading
    (atx_h2_marker)
    heading_content: (inline) @heading)
  (setext_heading
    heading_content: (paragraph) @heading
    (setext_h2_underline))
]
"""

LINK_QUERY = """
(link_reference_definition
  (link_label) @label)
"""

SEMVER_PATTERN = (
    r"(?:0|[1-9]\d*)\.(?:0|[1-9]\d*)\.(?:0|[1-9]\d*)"
    r"(?:-(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*)?"
    r"(?:\+[0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*)?"
)
HEADING_REGEX = re.compile(rf"\[({SEMVER_PATTERN})\]\s+-\s+(\d{{4}}-\d{{2}}-\d{{2}})$")
UNRELEASED_REGEX = re.compile(r"\[?unreleased", re.IGNORECASE)


def capture(language, tree, source: str, name: str) -> list[str]:
    cursor = QueryCursor(Query(language, source))
    nodes = cursor.captures(tree.root_node).get(name, [])
    nodes.sort(key=lambda node: node.start_byte)
    return [node.text.decode("utf-8").strip() for node in nodes]


def check(path: Path, content: bytes) -> list[str]:
    language = Language(tree_sitter_markdown.language())
    tree = Parser(language).parse(content)

    headings = capture(language, tree, HEADING_QUERY, "heading")
    if not headings:
        return [f"Error: {path} has no level-2 headings"]

    # CommonMark link labels are case-insensitive
    link_labels = {
        label.strip("[]").casefold()
        for label in capture(language, tree, LINK_QUERY, "label")
    }

    errors = []
    versions: set[str] = set()
    previous_date = None

    for heading in headings:
        if UNRELEASED_REGEX.match(heading):
            errors.append(
                f"Error: {path} must not contain an Unreleased section: '{heading}'"
            )
            continue

        m = HEADING_REGEX.match(heading)
        if not m:
            errors.append(
                f"Error: heading '{heading}' in {path} does not match '[<semver>] - YYYY-MM-DD'"
            )
            continue

        version, date_str = m.group(1), m.group(2)

        if version.casefold() in versions:
            errors.append(f"Error: duplicate version [{version}] in {path}")
        versions.add(version.casefold())

        try:
            date = datetime.date.fromisoformat(date_str)
        except ValueError:
            errors.append(
                f"Error: invalid date '{date_str}' in heading '{heading}' in {path}"
            )
        else:
            if previous_date is not None and date > previous_date:
                errors.append(
                    f"Error: heading '{heading}' in {path} is newer than the section above it; "
                    "sections must be ordered newest first"
                )
            previous_date = date

        if version.casefold() not in link_labels:
            errors.append(
                f"Error: missing link reference definition for version [{version}] at bottom of {path}"
            )

    for label in sorted(link_labels - versions - {"unreleased"}):
        errors.append(
            f"Error: dangling link reference definition [{label}] in {path} has no matching section"
        )

    return errors


def main() -> None:
    path = Path(sys.argv[1] if len(sys.argv) > 1 else "CHANGELOG.md")
    if not path.is_file():
        print(f"Error: file not found: {path}", file=sys.stderr)
        sys.exit(1)

    errors = check(path, path.read_bytes())
    if errors:
        for err in errors:
            print(err, file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
