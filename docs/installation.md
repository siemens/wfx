# Build and Installation

Pre-built binaries, as well as Debian and RPM packages, are available [here](https://github.com/siemens/wfx/releases) for Linux, specifically [x86_64](https://github.com/golang/go/wiki/MinimumRequirements#amd64) and [arm64](https://github.com/golang/go/wiki/MinimumRequirements#arm64) architectures.

If pre-built binaries are not available (refer to `go tool dist list` for alternative platforms and architectures, such
as Windows or macOS), or if specific features need to be disabled during compilation, building wfx from source is
necessary.

## Building wfx

A recent [Go compiler](https://go.dev/) >= 1.18 as well as [GNU make](https://www.gnu.org/software/make/) wrapping the `go build` commands is required to build wfx and its associated tools.

wfx requires a persistent storage to save workflows and jobs.
Go [Build Tags](https://pkg.go.dev/go/build) are used to select compiled-in support for different persistent storage options.
The following persistent storage selection build tags are available:

| Build Tag    | Description                                                      |
| :----------- | :--------------------------------------------------------------- |
| `sqlite`     | Enable built-in [SQLite](https://www.sqlite.org/) support        |
| `libsqlite3` | Dynamically link against `libsqlite3`                            |
| `postgres`   | Enable built-in [PostgreSQL](https://www.postgresql.org) support |
| `mysql`      | Enable built-in [MySQL](https://www.mysql.com/) support          |

By default, all built-in persistent storage options are enabled.

Note that the selection of build tags can impact the size of the `wfx` binary file and may as well have implications for the software clearing process, including obligations that must be met.

Building wfx via

```bash
make
```

results in the following binaries:

- `wfx`: The server component providing the RESTful APIs for managing workflows and jobs.
- `wfxctl`: Command line client for interacting with the wfx.
- `wfx-loadtest`: Command line tool for load-testing a wfx instance.
- `wfx-viewer`: Convenience tool to visualize workflows in PlantUML or SVG format.

All binaries have extensive help texts when invoked with `--help`.

To build and compile-in, e.g., SQLite persistent storage support only, according `GO_TAGS` must be given:

```bash
make GO_TAGS=sqlite
```

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
