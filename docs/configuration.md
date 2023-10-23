# Configuration

wfx is configured in the following order of precedence using

1. command line parameters (e.g. `--log-level`),
2. environment variables (prefixed with `WFX_`, e.g. `WFX_LOG_LEVEL`),
3. configuration files in [YAML](https://github.com/go-yaml/yaml#compatibility) format (either via the `--config` command line parameter or present in one of the default search locations, see `wfx --help`).

Note that wfx supports configuration file live reloading so that a running wfx instance can be reconfigured without the need for restarting it.

Without configuration,

```bash
/usr/bin/wfx
```

starts wfx using the built-in SQLite-based persistent storage.
The northbound (management) API is available at [http://127.0.0.1:8081/api/wfx/v1/](http://127.0.0.1:8081/api/wfx/v1/),
whereas the southbound (client) API is available at a different port: [http://127.0.0.1:8080/api/wfx/v1](http://127.0.0.1:8080/api/wfx/v1).

## Systemd Integration

For production deployments, it's recommended to run wfx under a service supervisor such as [systemd](https://systemd.io).
The `share/systemd` directory provides pre-configured systemd service units.
These units are also included in the distribution packages available with wfx releases.

```
systemctl enable --now wfx@foo.socket
# multiple instances of wfx can be running at the same time, e.g.
systemctl enable --now wfx@bar.socket
```

The wfx services launch on-demand, i.e., they start when a client connects, such as when retrieving the wfx version:

```bash
wfxctl --client-unix-socket /var/run/wfx/foo/client.sock version
```

## Persistent Storage

wfx requires a persistent storage to save workflows and jobs.
The default persistent storage is [SQLite](#sqlite) and requires no further configuration.

The command line argument `--storage={sqlite,postgres,mysql}` is available to choose a persistent storage backend out of the compiled-in available ones at run-time.
Each persistent storage allows further individual configuration via `--storage-opt=<options>`.

Note that wfx needs to manage the database schema and hence needs appropriate permissions to, e.g., create tables.
This is in particular important for the PostgreSQL and MySQL persistent storage options as they're connecting to an external database service which has to be setup accordingly.

### SQLite

[SQLite](https://www.sqlite.org/) is the default persistent storage and automatically selected if no other persistent storage option is given.
It can be further configured with the `--storage-opt` configuration option, see the [go-sqlite3 Wiki](https://github.com/mattn/go-sqlite3/wiki/DSN) for available Data Source Name (DSN) options.

As an example, the following command runs a wfx instance with an ephemeral in-memory SQLite database:

```bash
wfx --storage sqlite --storage-opt "file:wfx?mode=memory&cache=shared&_fk=1"
```

Note that all state is lost on wfx exiting so it's advised to use it for testing purposes only.

### PostgreSQL

[PostgreSQL](https://www.postgresql.org/) is a well-known open source object-relational database.
Via environment variables or a Data Source Name (DSN) passed as `--storage-opt`, the link to a PostgreSQL instance is configured, see PostgreSQL's [parameter key word names](https://www.postgresql.org/docs/15/libpq-connect.html#LIBPQ-PARAMKEYWORDS) for DSN and [environment variables](http://www.postgresql.org/docs/15/static/libpq-envars.html) documentation for details and available options.

As an example, the following two commands each run a wfx instance connecting to the same PostgreSQL instance but with different configuration means:

```bash
# Configuration via DSN key=value string
wfx --storage postgres \
    --storage-opt "host=localhost port=5432 user=wfx password=secret database=wfx" &

# Configuration using environment variables
env PGHOST=localhost \
    PGPORT=5432 \
    PGUSER=wfx \
    PGPASSWORD=secret \
    PGDATABASE=wfx \
    wfx --storage postgres
```

### MySQL

[MySQL](https://www.mysql.com/) is another well-known open source relational database.

Note that [MariaDB](https://mariadb.org) is currently unsupported due to the lack of certain JSON features (specifically the inability to directly index JSON data).

With the Data Source Name (DSN) passed as `--storage-opt`, the link to a MySQL instance is configured, see [Go's SQL Driver](https://github.com/go-sql-driver/mysql#dsn-data-source-name) for available options and as reference.

As an example, the following command runs a wfx instance connecting to a MySQL instance using similar configuration options as for PostgreSQL:

```bash
# Configuration via DSN URL string
wfx --storage mysql \
    --storage-opt "wfx:secret@tcp(localhost:3306)/wfx"
```

## Communication Channels

wfx currently supports the following network communication channels:

- `http`: Unencrypted HTTP
- `https`: HTTP over TLS (Transport Layer Security)
- `unix`: Unix-domain sockets

With the `--scheme` configuration option, one or multiple from the preceding list are enabled.
For instance, to use wfx in HTTPS-only mode:

```bash
wfx --scheme=https \
    --tls-certificate=localhost/cert.pem \
    --tls-key=localhost/key.pem
```

To enable both, HTTP and HTTPS, simultaneously, bound to different hosts:

```bash
wfx --scheme=http,https \
    --client-host=localhost \
    --client-tls-host=0.0.0.0 \
    --tls-certificate=localhost/cert.pem \
    --tls-key=localhost/key.pem
```

To exclusively use Unix-domain sockets:

```bash
wfx --scheme unix \
    --client-unix-socket /tmp/wfx-client.sock \
    --mgmt-unix-socket /tmp/wfx-mgmt.sock
```

The following connectivity parameters are available:

| Parameter           | Description                                                                       |
| :------------------ | :-------------------------------------------------------------------------------- |
| `--scheme`          | One or multiple communication schemes to be used for client-server communication. |
| `--client-host`     | The address to listen on for client HTTP requests                                 |
| `--client-port`     | The port to listen on for client HTTP requests                                    |
| `--client-tls-host` | Same as `--client-host` but for HTTP over TLS                                     |
| `--client-tls-port` | Same as `--client-port` but for HTTP over TLS                                     |
| `--mgmt-host`       | The address to listen on for wfx management / operator HTTP requests              |
| `--mgmt-port`       | The port to listen on for wfx management /operator HTTP requests                  |
| `--mgmt-tls-host`   | Same as `--mgmt-host` but for HTTP over TLS                                       |
| `--mgmt-tls-port`   | Same as `--mgmt-port` but for HTTP over TLS                                       |
| `--tls-certificate` | The location of the TLS certificate file                                          |
| `--tls-key`         | The location of the TLS key file                                                  |
| `--tls-ca`          | The certificate authority certificate file for mutual TLS authentication          |

## File Server

wfx comes with a built-in file server that serves artifacts at `http://<wfx host:{client,mgmt} port>/download/`.
This feature is particularly useful for dynamic deployments or when an external file storage solution is unavailable.
To configure the directory that backs the file server URL's contents, use the `--simple-fileserver=/path/to/folder` option.

Note that this feature is disabled by default and must be explicitly enabled at run-time using this option.
