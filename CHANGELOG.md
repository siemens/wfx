# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Add optional `description` field to workflows
- Job event notifications via server-sent events (see #11)
- Plugin System for External Middlewares (see #43)

### Fixed

- Send HTTP status code 404 when attempting to access the file server while it is disabled
- Configure TLS for Southbound API (if requested via CLI)
- Connection pool leak due to schema migrations (SQLite, MySQL)

### Changed

- Refactored `wfxctl workflow delete` command to accept workflows as arguments instead of positional parameters
- Prefer cgroup CPU quota over host CPU count
- Empty or `null` arrays are omitted from JSON responses

### Removed

## [0.1.0] - 2023-02-06

Initial release of wfx.

[unreleased]: https://github.com/siemens/wfx/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/siemens/wfx/releases/tag/v0.1.0
