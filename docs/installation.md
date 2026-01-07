# Build and Installation

Pre-built binaries, as well as Debian and RPM packages, are available [here](https://github.com/siemens/wfx/releases) for Linux, specifically [x86_64](https://go.dev/wiki/MinimumRequirements#amd64) and [arm64](https://go.dev/wiki/MinimumRequirements#arm64) architectures.

To start a container hosting wfx, follow these commands:

```bash
# create a named volume to persist data (only needed the first time)
docker volume create wfx-db
docker run --rm -v wfx-db:/home/nonroot \
    -p 8080:8080 -p 8081:8081 \
    ghcr.io/siemens/wfx:latest
```

If pre-built binaries are not available (refer to `go tool dist list` for alternative platforms and architectures, such
as Windows or macOS), or if specific features need to be disabled during compilation, building wfx from source is
necessary.

## Building wfx

A recent [Go compiler](https://go.dev/) (see `go.mod`):

```bash
# build wfx
go build

# build the rest (optional)
go build ./cmd/wfxctl
go build ./cmd/wfx-loadtest
go build ./cmd/wfx-viewer
```

The above commands produce the following binaries:

- `wfx`: The server component providing the RESTful APIs for managing workflows and jobs.
- `wfxctl`: Command line client for interacting with the wfx.
- `wfx-loadtest`: Command line tool for load-testing a wfx instance.
- `wfx-viewer`: Convenience tool to visualize workflows in different formats (e.g. PlantUML, Mermaid).

All binaries have extensive help texts when invoked with `--help`.

### Build Tags

Go [build tags](https://pkg.go.dev/go/build) can be used to opt-out from various features at compile-time.
By default, all features are enabled. The following build tags are available:

| Build Tag     | Description                                                                                           |
| :------------ | :---------------------------------------------------------------------------------------------------- |
| `ui`          | Enable built-in WebUI (must have been built separately before, see [building WebUI](#building-webui)) |
| `no_sqlite`   | Disable built-in [SQLite](https://www.sqlite.org/) support                                            |
| `libsqlite3`  | Dynamically link against `libsqlite3`                                                                 |
| `no_postgres` | Disable built-in [PostgreSQL](https://www.postgresql.org) support                                     |
| `no_mysql`    | Disable built-in [MySQL](https://www.mysql.com/) support                                              |
| `no_plugin`   | Disable support for [external plugins](operations.md#plugins)                                         |
| `no_swagger`  | Disable legacy `swagger.json` endpoint                                                                |

Example:

```bash
go build -tags no_mysql,no_sqlite
```

Note:

- wfx requires at least one persistent storage to save workflows and jobs
- build tags can impact the size of the `wfx` binary file and may as well have implications for the software clearing process, including obligations that must be met

### Debian

The Go toolchain provided by Debian _stable_ is often outdated; it's typically end-of-life upstream but still maintained
by Debian's security team. Therefore, to compile wfx from source in Debian _stable_, the `-backports` repository is
necessary. In contrast, for Debian _testing_, it usually works out of the box since it ships with a recent version of
the Go toolchain.

## Building WebUI

Prerequisites:

- [Gleam](https://gleam.run/)
- [Rebar3](https://rebar3.org/)
- [npm](https://www.npmjs.com/)

Then run

```sh
$ cd ui

# install deps
$ npm install
$ gleam deps download

# build ui
$ gleam run -m lustre/dev build --minify app

# generate files needed for go build
$ cd ..
$ go generate -tags ui ./ui
# build wfx and embed ui files
$ go build -tags ui
```

## Installing wfx

wfx's release binaries are statically linked and self-contained.
Hence, an installation isn't strictly necessary, although if available, it's recommended to pick the distro packages (e.g. `*.deb` for Debian-based distros).

```bash
go install github.com/siemens/wfx@latest
```

Alternatively, a pre-built [Debian](https://www.debian.org) package is [provided](https://github.com/siemens/wfx/releases).

For convenience and ease of use, all binaries come with shell completions available for [Bash](https://www.gnu.org/software/bash/), [Fish](https://fishshell.com) and [Zsh](https://www.zsh.org).
To install the completions, refer to the binary's `completion --help` output, e.g. `wfx completion bash --help`.
