# golang-template

Small Go HTTP API organized by domain. Shared infrastructure lives in `internal/core`; every other package under `internal/` is a domain. Postgres runs in Docker. The API runs on the host.

## Prerequisites

- Go 1.27
- Docker

## Run

```sh
make up
make dev
```

- Health: `http://127.0.0.1:8081/api/v1/health`
- Create an item: `POST /api/v1/items` with `{"name":"widget"}`
- List items: `GET /api/v1/items`
- Get one item: `GET /api/v1/items/{id}`

## Checks

```sh
make test-unit
make test
make lint
```

`make test-unit` does not need Docker. `make test` needs `make up`.

Copy `.env.example` if you want to export configuration yourself. The `dev` and `migrate` targets set `DATABASE_URL` for you.

See `AGENTS.md` for the layout and the checklist to adapt this template.
