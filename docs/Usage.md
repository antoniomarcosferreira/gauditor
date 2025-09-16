# Usage Guide

This guide shows how to use Gauditor as a library, run the HTTP server, integrate with Gin, and choose a storage backend.

## Install

```bash
go get github.com/antoniomarcosferreira/gauditor
```

## Library: Record and Query

```go
import (
  "context"
  g "github.com/antoniomarcosferreira/gauditor/pkg/gauditor"
)

func recordLogin(ctx context.Context) error {
  rec := g.NewRecorder(g.NewMemoryStorage())
  _, err := rec.Record(ctx, g.Event{
    Tenant: "acme",
    Actor:  g.Actor{ID: "u123", Attributes: map[string]any{"type": "user", "email": "alice@example.com"}},
    Action: "login",
  })
  return err
}
```

Querying:

```go
until := time.Now().Add(1 * time.Hour)
list, err := rec.Query(ctx, g.Query{Tenant: "acme", Until: &until, Limit: 100})
```

## Simple API: EasyRecorder

```go
rec := g.NewRecorder(g.NewMemoryStorage())
ez := g.NewEasyRecorder(rec)
_, _ = ez.Record(ctx, "acme", map[string]any{"name": "Alice"}, "update", map[string]any{"model": "users"})
```

## HTTP Server

Run:

```bash
go run ./cmd/gauditor
# listens on :8091 (override with GAUDITOR_ADDR or -addr)
```

Ingest:

```bash
curl -s -X POST localhost:8091/v1/events \
  -H 'content-type: application/json' \
  -d '{
    "tenant":"acme",
    "actor":{"id":"uuid","attributes":{"type":"user","email":"alice@example.com"}},
    "action":"login",
    "data":{"method":"password"}
  }'
```

Query:

```bash
curl -s 'localhost:8091/v1/events?tenant=acme'
```

## Gin Integration (Middleware)

Example middleware (see `examples/gincrud`):

- Automatically records successful requests (2xx/3xx)
- Action = `METHOD + route` (e.g., `POST /users`)
- Target.ID from `:id` when present

Run the sample:

```bash
go run ./examples/gincrud
```

## Storage Backends

See `docs/Storage.md` for full details.

Quick picks:
- Memory: development/testing
- Redis: simple demo (append-only lists)
- SQL (Postgres/MySQL): share your app DB, with configurable table prefix/name
- S3: append-only JSON objects (archive/data lake)

Programmatic examples:

```go
// Redis
rdb := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})
store := redisstore.New(rdb, redisstore.WithKeyPrefix("gauditor:"))
rec := gauditor.NewRecorder(store)

// SQL (Postgres)
db, _ := sql.Open("postgres", dsn)
store := sqlstore.New(db).ApplyOptions(sqlstore.WithTablePrefix("app_"))
_ = store.EnsureSchema(ctx)
rec := gauditor.NewRecorder(store)

// S3
cfg, _ := config.LoadDefaultConfig(ctx)
cli := s3.NewFromConfig(cfg)
store := s3store.New(cli, "my-bucket", "gauditor")
rec := gauditor.NewRecorder(store)
```

## Env-based Setup (Rails-like)

Use `gauditorenv.NewRecorderFromEnv` to create a recorder from environment variables.

```go
rec, err := gauditorenv.NewRecorderFromEnv(ctx)
```

Variables:
- GAUDITOR_STORAGE: `memory` | `redis` | `sql` | `s3`
- GAUDITOR_ADDR: `:8091` (HTTP server)
- Redis: `REDIS_ADDR`, `REDIS_KEY_PREFIX`
- SQL: `SQL_DRIVER`, `SQL_DSN`, `GAUDITOR_SQL_ENSURE_SCHEMA`
- S3: `S3_BUCKET`, `S3_PREFIX`, plus `AWS_*`

## Examples

- Basic: `examples/basic`
- HTTP client: `examples/httpclient`
- Gin CRUD + middleware: `examples/gincrud`
- Redis storage: `examples/redis`
