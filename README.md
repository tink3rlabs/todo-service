# todo-service

todo-service is the canonical reference service for the
[`github.com/tink3rlabs/magic`](https://github.com/tink3rlabs/magic) library. It is
a small but complete microservice — a CRUD API for todo items — that demonstrates
magic's features end to end.

A step-by-step tutorial that builds this service is in the magic documentation.

## Features showcased

- **Storage adapters with startup migrations** — `storage.StorageAdapterFactory`
  selects a memory, SQL (postgresql, mysql, sqlite), or dynamodb adapter from
  configuration; migrations run at startup via `storage.NewDatabaseMigration`.
- **Listing and search** — `GET /todos` supports both a structured list and a
  Lucene `?filter=` search, cursor-paginated with `limit` and `next`.
- **Full CRUD including JSON Patch** — GET/POST/PUT/PATCH/DELETE on `/todos`, with
  PATCH accepting JSON Patch documents.
- **Typed errors mapped to HTTP status** — `magic/errors` values are translated to
  status codes by the `ErrorHandler` middleware.
- **Request validation** — `Validator` middleware backed by JSON schemas.
- **Health probes** — `GET /health/liveness` and `GET /health/readiness`.
- **Observability** — magic's `observability` package exports Prometheus or OTLP
  metrics and traces, including a custom `todo_service_todos_created_total`
  counter; metrics are served at `/metrics`.
- **Config-gated JWT auth** — `EnsureValidToken` plus `RequireRole` guard the write
  routes when enabled; reads are public.
- **Config-gated pub/sub** — publishes `todo.created` and `todo.updated` events via
  magic's `pubsub` package over SNS.
- **Leadership election and scheduling** — the `leadership` package elects a leader
  that runs a gocron scheduler.
- **OpenAPI generation** — the spec is generated with `openapi-godoc` and served at
  `/api-docs`.
- **cobra CLI + viper config** — a `server` subcommand with configuration loaded by
  viper from an embedded `embed.FS`.

## Running it locally

Prerequisites: Go 1.25.

```bash
go run . server --config config/development.yaml
```

The service defaults to the in-memory storage adapter, so it runs with no external
services. Auth and pub/sub are disabled by default, so no tokens and no AWS
credentials are needed.

`config/openapi.json` is a generated artifact. Regenerate it with:

```bash
go generate ./...
```

## Configuration

Configuration lives in `config/development.yaml`. The configurable blocks are:

- `storage` — adapter selection and connection settings
- `auth` — JWT validation and role requirements (disabled by default)
- `observability` — metrics and tracing backends
- `pubsub` — SNS event publishing (disabled by default)
- `leadership` — leader election
- `health` — health probe behavior
- `logger` — log level and format

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/todos` | List or Lucene-search todos (cursor-paginated) |
| POST | `/todos` | Create a todo |
| GET | `/todos/{id}` | Get a todo |
| PUT | `/todos/{id}` | Replace a todo |
| PATCH | `/todos/{id}` | Update a todo with a JSON Patch document |
| DELETE | `/todos/{id}` | Delete a todo |
| GET | `/health/liveness` | Liveness probe |
| GET | `/health/readiness` | Readiness probe |
| GET | `/metrics` | Prometheus metrics |
| GET | `/api-docs` | OpenAPI specification |
