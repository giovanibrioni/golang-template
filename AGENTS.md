# AGENTS.md

Go HTTP template for a small API. Business domains live in their own packages under `internal/`. Shared infrastructure lives in `internal/core`. One example domain, `items`, creates, lists, and fetches records in Postgres.

This cut does not include auth, CORS, a queue, OpenAPI, OpenTelemetry, or a collector.

## Layout

```text
cmd/api/main.go                     process entrypoint, migrations, HTTP server
internal/api/router.go              Chi router and /api/v1
internal/core/config/config.go      environment configuration
internal/core/logging/logging.go    slog setup and request id on log records
internal/core/db/db.go              pool, transaction context, Beginner
internal/core/db/migrate.go         embedded SQL migrations
internal/core/httpx/httpx.go        request transaction, JSON, HTTP error mapping
internal/core/pgtest/pgtest.go      Postgres helper for integration tests
internal/health/handler.go          GET /api/v1/health
internal/items/                     example feature
migrations/                         SQL files and embed.FS
```

`internal/core` holds the cross-cutting infrastructure shared by every feature. Every other package under `internal/` is a business domain that owns its model, repository, service, and HTTP handler.

Request flow:

```text
Chi -> httpx.Wrap -> handler -> service -> repository -> pgx.Tx
```

`httpx.Wrap` opens the transaction, stores it on the context, and commits only when the handler returns nil. Rollback happens on any error. The handler returns `httpx.Response` and an `error`. It does not write status codes itself.

## New feature

Copy `internal/items` to `internal/<feature>` and adapt it.

1. Add the model, a `Repository` interface, a Postgres implementation, a service, and a handler with `Mount`.
2. Mount the package from `internal/api/router.go` under `/api/v1`.
3. Add a new SQL file in `migrations/`. Names sort lexicographically. Do not edit a migration that has already been applied.
4. Add a unit test with an in-memory fake of the repository.
5. Add an integration test for the repository and one for the HTTP handler.

Rules:

- Do not call `Commit` or `Rollback` outside `httpx.Wrap`. Repository methods use `db.Tx` from the context.
- Return a response struct from the handler. Do not pass driver rows to HTTP.
- Keep domain errors in the feature package. Give them `Kind() string` and `PublicDetail() string`. `httpx` maps `not_found` to 404 and `validation` to 400. Anything else becomes 500 and is logged.
- Log with `slog` and the default logger (`InfoContext`, `ErrorContext`). Do not use `fmt.Println`. Do not log secrets.
- Register routes in the feature `Mount` function so a new endpoint stays next to its service.

## Tests

`make test-unit` runs `go test -short ./...` and does not need Docker.

`make test` needs `make up`. It points `DATABASE_URL` at the `app_test` database. `internal/core/pgtest` applies migrations once, then each test rolls back an outer transaction. HTTP tests use `pgtest.RollbackBeginner` so the commit inside `httpx.Wrap` is only a savepoint. Repository tests use `pgtest.TxContext`.

The fake repository in the unit test must implement the feature `Repository` interface.

## Commands

- `make help`
- `make install`
- `make up` / `make down`
- `make dev`
- `make migrate`
- `make test-unit`
- `make test`
- `make lint`
- `make review`

## Adapt this template

- Change the module path in `go.mod` and in every import.
- Change `APP_NAME` if the service name should differ.
- Change `POSTGRES_DB`, the user, and the password if you do not want the local defaults.
- Rename the Compose volume if you want a different Docker volume.
- Delete `internal/items` and `migrations/001_create_items.sql` when items is no longer the example. Remove the items route from `internal/api/router.go`.
