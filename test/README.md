# Tests

This directory tests the **command-line interfaces**.
In particular, it aims to test the examples used in the [documentation](../docs) directory.
This means, if a test here fails, it is likely that the documentation must be updated as well.

## Prerequisites

You need to install [bats](https://github.com/bats-core/bats) in order to run the tests.

The tests further assume you have a PostgreSQL and MySQL database running locally.
If you have either [podman](https://podman.io/) or [docker](https://www.docker.com/) installed, you can use the provided
`justfile` to do so:

```bash
just postgres-start mysql-start
```

To shut down the databases, running

```bash
just postgres-stop mysql-stop
```

## Running the tests locally

```bash
go test ./...
```

will execute all tests.

Note: Make sure you have the git submodules checked out (`git submodule update --init`).

Alternatively, you can run the individual `.bats` files.
Individual test cases can be run like that: `bats . --filter "TLS mixed-mode"`
