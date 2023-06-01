# Development

Development notes.

Unit- and integration tests are triggered automatically for each commit.

## Build tags

To make tools like `gopls` (the Go language server) understand build tags,
set the `GOFLAGS` environment variable accordingly, e.g.

```sh
export GOFLAGS="-tags=integration,sqlite,mysql,postgres"
```

## Persistence

### entgo

The database migration is managed using [atlas](https://atlasgo.io/).
See [here](https://entgo.io/docs/versioned-migrations/) for an in-depth explanation.

The schema is defined programmatically in `generated/ent/schema/`.

**Example**: Adding a new column to jobs

1. Modify `./generated/ent/schema/job.go`.
2. Generate code, e.g. `go generate ./generated/ent`.
3. Generate `*.sql` migrations: `just postgres-generate-schema added-column`

**Example**: Update atlas.sum

```bash
$ docker run --rm -v $(pwd):/git docker.io/arigaio/atlas migrate hash --dir file:///git
```

This is necessary if you edit the sql files.
