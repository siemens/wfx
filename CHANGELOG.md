# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

### Fixed

- `wfx` would not start if it was built without plugins support

### Changed

- Use zstd instead of xz to compress release tarballs

## [0.3.1] - 2024-07-09

### Added

### Fixed

### Changed

- Use zstd instead of xz to compress release tarballs

## [0.3.0] - 2024-05-02

### Added

- Log HTTP response code
- wfx-viewer: additional output formats mermaid.js and state-machine-cat

### Fixed

- Compilation for Windows (not officially supported though)
- Include workflow definition in response when creating jobs
- Correctly log non-JSON response body of plugins

### Changed

- Log messages from automaxprocs/maxprocs are now seamlessly integrated into existing logging framework

### Removed

## [0.2.0] - 2024-01-15

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
- Build requires Go >= 1.21

### Removed

## [0.1.0] - 2023-02-06

Initial release of wfx.

[0.1.0]: https://github.com/siemens/wfx/releases/tag/v0.1.0
[0.2.0]: https://github.com/siemens/wfx/releases/tag/v0.2.0
[0.3.0]: https://github.com/siemens/wfx/releases/tag/v0.3.0
[0.3.1]: https://github.com/siemens/wfx/releases/tag/v0.3.1
[unreleased]: https://github.com/siemens/wfx/compare/v0.3.1...HEAD
