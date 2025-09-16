# Gauditor Storage Backends

This document explains how to configure and use storage backends for Gauditor. You can embed storage programmatically or via environment variables for a Rails-like developer experience.

## Quick choice

- Development/simple: Memory (default)
- Local Redis: Redis (append-only list per tenant)
- Shared app DB: SQL (Postgres/MySQL) with table prefix to avoid collisions
- Data lake/archive: S3 (append-only JSON objects)

## Env-based configuration

Set `GAUDITOR_STORAGE` and related env vars. If unset or unknown, Memory is used.

- GAUDITOR_STORAGE: `memory` | `redis` | `sql` | `s3`
- GAUDITOR_ADDR: server address (e.g., `:8091`) if using the HTTP server

Redis (when `GAUDITOR_STORAGE=redis`):
- REDIS_ADDR: host:port (default `127.0.0.1:6379`)
- REDIS_KEY_PREFIX: key prefix (default `gauditor:`)

SQL (when `GAUDITOR_STORAGE=sql`):
- SQL_DRIVER: `postgres` or `mysql`
- SQL_DSN: driver-specific DSN
- GAUDITOR_SQL_ENSURE_SCHEMA: `1` (default) to create table/index if missing
- Optional table prefix/name via code (see below)

S3 (when `GAUDITOR_STORAGE=s3`):
- S3_BUCKET: bucket name (required)
- S3_PREFIX: key prefix (default `gauditor`)
- AWS credentials/region via standard `AWS_*` env vars

To construct a recorder from env in apps:

```go
ctx := context.Background()
rec, err := gauditorenv.NewRecorderFromEnv(ctx)
if err != nil { panic(err) }
```

## Programmatic configuration

### Memory
```go
rec := gauditor.NewRecorder(gauditor.NewMemoryStorage())
```

### Redis
```go
rdb := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})
store := redisstore.New(rdb, redisstore.WithKeyPrefix("gauditor:"))
rec := gauditor.NewRecorder(store)
```

### SQL (Postgres/MySQL)
```go
// import your driver: _ "github.com/lib/pq" or _ "github.com/go-sql-driver/mysql"
db, _ := sql.Open("postgres", dsn)
store := sqlstore.New(db).ApplyOptions(sqlstore.WithTablePrefix("app_")) // app_gauditor_events
_ = store.EnsureSchema(ctx)
rec := gauditor.NewRecorder(store)
```

### S3
```go
cfg, _ := config.LoadDefaultConfig(ctx)
cli := s3.NewFromConfig(cfg)
store := s3store.New(cli, "my-bucket", "gauditor")
rec := gauditor.NewRecorder(store)
```

## Notes and trade-offs

- Redis and S3 examples are optimized for simplicity, not massive queries.
- SQL backend supports configurable table name/prefix to share the same database as your app safely.
- All backends honor `context.Context` cancellation.
- Use `WithIDGenerator` and `WithClock` to ensure deterministic IDs/timestamps in tests.

## Samples

- `examples/redis`: Redis-backed storage usage
- `examples/gincrud`: Gin CRUD app with automatic auditing middleware
- HTTP server: `cmd/gauditor` (REST ingestion/query)
