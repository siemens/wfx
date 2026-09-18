# SPDX-FileCopyrightText: 2026 Siemens AG
#
# SPDX-License-Identifier: Apache-2.0
#
# Author: Michael Adler <michael.adler@siemens.com>
import importlib.util
from pathlib import Path

spec = importlib.util.spec_from_file_location(
    "check_changelog", Path(__file__).parent / "check-changelog.py"
)
mod = importlib.util.module_from_spec(spec)
spec.loader.exec_module(mod)

GOOD = """\
# Changelog

## [1.0.0] - 2024-02-06

## [0.9.0-rc1] - 2024-01-06

[1.0.0]: https://example.com/1.0.0
[0.9.0-RC1]: https://example.com/0.9.0-rc1
"""


def check(text: str) -> list[str]:
    return mod.check(Path("CHANGELOG.md"), text.encode())


def test_valid():
    assert check(GOOD) == []


def test_no_h2_headings():
    assert check("# Changelog\n") == ["Error: CHANGELOG.md has no level-2 headings"]


def test_unreleased():
    errors = check(GOOD.replace("## [1.0.0]", "## Unreleased\n\n## [1.0.0]"))
    assert len(errors) == 1
    assert "must not contain an Unreleased" in errors[0]


def test_fenced_code_ignored():
    assert check(GOOD + "\n```\n## Unreleased\n```\n") == []


def test_duplicate_version():
    errors = check(
        GOOD.replace("## [0.9.0-rc1] - 2024-01-06", "## [1.0.0] - 2024-01-06")
    )
    assert "Error: duplicate version [1.0.0] in CHANGELOG.md" in errors


def test_order_newest_first():
    errors = check(GOOD.replace("2024-01-06", "2024-03-06"))
    assert any("ordered newest first" in e for e in errors)


def test_missing_link_definition():
    errors = check(GOOD.replace("[1.0.0]: https://example.com/1.0.0\n", ""))
    assert errors == [
        "Error: missing link reference definition for version [1.0.0] at bottom of CHANGELOG.md"
    ]


def test_dangling_link_definition():
    errors = check(GOOD + "[0.1.0]: https://example.com/0.1.0\n")
    assert errors == [
        "Error: dangling link reference definition [0.1.0] in CHANGELOG.md has no matching section"
    ]


def test_unreleased_link_definition_allowed():
    assert check(GOOD + "[unreleased]: https://example.com/compare\n") == []


def test_invalid_date():
    errors = check(GOOD.replace("2024-02-06", "2024-02-31"))
    assert any("invalid date '2024-02-31'" in e for e in errors)


def test_malformed_heading():
    errors = check(GOOD.replace("## [1.0.0] - 2024-02-06", "## 1.0 (2024-02-06)"))
    assert any("does not match" in e for e in errors)
