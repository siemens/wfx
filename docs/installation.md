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

A recent [Go compiler](https://go.dev/) (see `go.mod`) as well as [GNU make](https://www.gnu.org/software/make/) wrapping the `go build` commands is required to build wfx and its associated tools:

```bash
make
```

The above command produces the following binaries:

- `wfx`: The server component providing the RESTful APIs for managing workflows and jobs.
- `wfxctl`: Command line client for interacting with the wfx.
- `wfx-loadtest`: Command line tool for load-testing a wfx instance.
- `wfx-viewer`: Convenience tool to visualize workflows in different formats (e.g. PlantUML, Mermaid).

All binaries have extensive help texts when invoked with `--help`.

### Build Tags

Go [build tags](https://pkg.go.dev/go/build) are used to select compiled-in support for various features.
The following persistent storage selection build tags are available:

| Build Tag    | Description                                                      |
| :----------- | :--------------------------------------------------------------- |
| `sqlite`     | Enable built-in [SQLite](https://www.sqlite.org/) support        |
| `libsqlite3` | Dynamically link against `libsqlite3`                            |
| `postgres`   | Enable built-in [PostgreSQL](https://www.postgresql.org) support |
| `mysql`      | Enable built-in [MySQL](https://www.mysql.com/) support          |
| `plugin`     | Enable support for [external plugins](operations.md#plugins)     |

By default, all built-in persistent storage options are enabled (wfx requires at least one persistent storage to save workflows and jobs).

Note that the selection of build tags can impact the size of the `wfx` binary file and may as well have implications for the software clearing process, including obligations that must be met.

To build and compile-in, e.g., SQLite persistent storage support only, according `GO_TAGS` must be given:

```bash
make GO_TAGS=sqlite
```

### Debian

The Go toolchain provided by Debian _stable_ is often outdated; it's typically end-of-life upstream but still maintained
by Debian's security team. Therefore, to compile wfx from source in Debian _stable_, the `-backports` repository is
necessary. In contrast, for Debian _testing_, it usually works out of the box since it ships with a recent version of
the Go toolchain.

## Installing wfx

wfx's release binaries are statically linked and self-contained.
Hence, an installation isn't strictly necessary, although if available, it's recommended to pick the distro packages (e.g. `*.deb` for Debian-based distros).

Nevertheless, for convenience on UNIXy systems,

```bash
make DESTDIR= prefix= install
```

installs the binaries to `/bin`.
Giving a different `DESTDIR` and/or `prefix` allows to adjust to other locations.

Alternatively, a pre-built [Debian](https://www.debian.org) package is [provided](https://github.com/siemens/wfx/releases).

For convenience and ease of use, all binaries come with shell completions available for [Bash](https://www.gnu.org/software/bash/), [Fish](https://fishshell.com) and [Zsh](https://www.zsh.org).
To install the completions, refer to the binary's `completion --help` output, e.g. `wfx completion bash --help`.
