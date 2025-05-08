# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- `wfxctl workflow query` now accepts a `sort` param
- Added (existing but undocumented) `/health` and `/version` endpoint to OpenAPI spec
- OpenAPI v3 spec is served at `/api/wfx/v1/openapi.json`
- Add build tags to `--version` output

### Fixed

- `wfx`: implemented sort functionality for `/workflows` endpoint
- `wfx`: fixed an issue where job event connections were prematurely closed due to inactivity

### Changed

- Start and enable wfx service on installation and stop and disable it on removal
- Migrated from Swagger to OpenAPI v3
- The previous Swagger (OpenAPI v2) specification is still available at `/api/wfx/v1/swagger.json` to _ensure compatibility_ with older clients (e.g., SWUpdate <= 2024.12). This endpoint will be removed in a future release.
- The top-level `/swagger.json` is no longer served, as no known clients make use of it.
- `wfxctl workflow get` uses southbound API by default
- `wfxctl health` validates the certificate chain when using TLS
- Forbbiden requests (e.g. job creation via southbound API) now return HTTP status code 403 instead of 405
- System certificates will be loaded automatically for TLS communication
- Access log level was changed from `INFO` to `DEBUG` to avoid logging a message for every poll by each client.
  To restore the old behavior, start wfx with `--log-level=DEBUG` (note that this will enable additional log messages
  though).

## [0.3.3] - 2024-12-23

### Added

- Generate SBOMs for release artifacts
- Generate separate man pages for subcommands

## [0.3.2] - 2024-09-03

### Fixed

- `wfx` would not start if it was built without plugins support

## [0.3.1] - 2024-07-09

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

## [0.1.0] - 2023-02-06

Initial release of wfx.

[0.1.0]: https://github.com/siemens/wfx/releases/tag/v0.1.0
[0.2.0]: https://github.com/siemens/wfx/releases/tag/v0.2.0
[0.3.0]: https://github.com/siemens/wfx/releases/tag/v0.3.0
[0.3.1]: https://github.com/siemens/wfx/releases/tag/v0.3.1
[0.3.2]: https://github.com/siemens/wfx/releases/tag/v0.3.2
[0.3.3]: https://github.com/siemens/wfx/releases/tag/v0.3.3
[unreleased]: https://github.com/siemens/wfx/compare/v0.3.3...HEAD
