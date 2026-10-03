# Review checklist

Architecture rules live in `AGENTS.md`. This file is the extra judgment the scripts do not encode.

## Architecture (`architecture_check.go`)

A domain package is any direct child of `internal/` other than `core`, `api`, and `health`.

The script fails when a domain is missing `model.go`, `repository.go`, `service.go`, or `handler.go`; when the domain is not imported from `internal/api/router.go`; when the domain tests declare no fake repository type; when `repository_test.go` or `handler_test.go` does not use `pgtest`; when `.Commit(` or `.Rollback(` appears outside `internal/core/httpx/httpx.go`, `internal/core/db/migrate.go`, or `internal/core/pgtest/pgtest.go`; when `fmt.Print*` is used; or when a file under `internal/core/` imports a domain package.

Still judge by hand:

- Request flow stays handler -> service -> repository -> `db.Tx` from the context.
- The repository exposes an interface plus a Postgres implementation. The fake implements that interface.
- Domain failures have `Kind() string` and `PublicDetail() string`. `httpx` maps `not_found` to 404 and `validation` to 400; anything else becomes 500 and is logged.
- Services do not touch transactions, HTTP, or the router.
- Domains do not import each other unless the user accepts the coupling.

## Orphans (`orphan_check.go`)

Fails when a package under `internal/` is never imported from a package other than its own, counting files under `cmd/`, `internal/`, and `migrations/`.

Bugbot covers unused code inside packages that are imported.

## Consistency (Bugbot, step 2)

No script covers these. Flag them when the diff introduces them:

- Race: check-then-act (read, then insert or update without a unique constraint or row lock) and lost updates on the same row.
- Dual write: one operation writes to the database and to a second store or external call, with no fallback, outbox, or compensation.
- Retry: a retry around a create or other non-idempotent write that has no idempotency key.

A unique constraint that closes an insert race still belongs in the merge-risk index check.

## Env (`env_example_check.go`)

Every `getenv("KEY", ...)` call in `internal/core/config/config.go` must have a key in `.env.example`.

## Merge risk (`merge_risk_check.go`)

Scope is the diff from the merge-base with `main` or `master` through the working tree, plus untracked files. With no commits, the scope is untracked files.

| Signal | Script | Reviewer |
| --- | --- | --- |
| Changed migration runs DROP, ALTER COLUMN, or RENAME SQL | FAIL | Call out lock risk, data loss, and expand/contract |
| `model.go` or `repository.go` changed and no `migrations/*.sql` changed | FAIL | Model drift |
| `handler.go` or `internal/api/router.go` changed | WARN | Removed fields, new required fields, path or status changes, renamed JSON keys |
| `repository.go` or `model.go` changed | WARN | New filters or foreign keys without an index in the migration |
| `internal/core/config/config.go` changed | WARN | New setting missing from deploy env, not only `.env.example` |

Breaking API examples: deleting a response field, changing a path, making an optional request field required, changing a documented status code. Adding an optional response field is compatible.

Index examples: a new `WHERE` column, a new foreign key, or a new `ORDER BY` on a large table without an index in the same migration.
