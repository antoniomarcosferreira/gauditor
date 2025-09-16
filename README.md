## Gauditor

The Go-powered audit trail solution: secure, simple, and scalable.

[![CI](https://github.com/antoniomarcosferreira/gauditor/actions/workflows/ci.yml/badge.svg)](https://github.com/antoniomarcosferreira/gauditor/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/antoniomarcosferreira/gauditor/pkg/gauditor.svg)](https://pkg.go.dev/github.com/antoniomarcosferreira/gauditor/pkg/gauditor)
[![Go Report Card](https://goreportcard.com/badge/github.com/antoniomarcosferreira/gauditor)](https://goreportcard.com/report/github.com/antoniomarcosferreira/gauditor)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](/LICENSE)

[Coverage Dashboard](docs/coverage.html) — updated via `make coverage`.

---

### Overview

Gauditor is a modern audit trail system for Go projects. It records immutable events about user and system activity, offering simple integration, strong data hygiene, and deployment flexibility (local, Docker, or cloud).

- **Library-first**: import and record events from your Go apps
- **HTTP API**: ingest and query events via REST
- **Pluggable storage**: starts with in-memory, extensible to durable backends

See also: `docs/About.md` for a narrative overview.

---

### Features

- **Comprehensive activity logging**: login, logout, CRUD, views, and more
- **Simple query model**: filter by tenant, actor, action, target
- **Immutable facts**: append-only records for trustworthy trails
- **Observability-friendly**: OpenAPI, easy to integrate with your stack
- **Cloud-ready**: containerized; infra as code planned

---

### Quickstart

Prerequisites: Go 1.23+

Run the server locally:

```bash
go run ./cmd/gauditor
# server listening on :8091
```

Ingest an event:

```bash
curl -s -X POST localhost:8091/v1/events \
  -H 'content-type: application/json' \
  -d '{
    "tenant":"acme",
    "actor":{"id":"uuid", "attributes":{"type":"user", "email":"alice@example.com", "name":"Alice"}},
    "action":"login",
    "data":{"method":"password"}
  }'
```

Query events:

```bash
curl -s 'localhost:8091/v1/events?tenant=acme'
```

Run the example:

```bash
go run ./examples/basic
```

---

### Samples

- Go library sample: `examples/basic`
- HTTP client sample: `examples/httpclient`
- Gin CRUD sample (integrated with gauditor): `examples/gincrud`
- Redis-backed storage sample (custom Storage): `examples/redis`

#### Env-based setup

Prefer a Rails-like experience? Configure storage via environment variables and build a recorder from env.

```go
ctx := context.Background()
rec, err := gauditorenv.NewRecorderFromEnv(ctx)
if err != nil { panic(err) }
```

Env keys:
- GAUDITOR_STORAGE: memory | redis | sql | s3 (default memory)
- Redis: REDIS_ADDR (127.0.0.1:6379), REDIS_KEY_PREFIX (gauditor:)
- SQL: SQL_DRIVER (postgres|mysql), SQL_DSN, GAUDITOR_SQL_ENSURE_SCHEMA=1
- S3: S3_BUCKET, S3_PREFIX (gauditor) + AWS_* creds/region

See `docs/Storage.md` for full details and code snippets.

#### Using pluggable storages

Redis (development/demo):

```go
import (
  "github.com/redis/go-redis/v9"
  rs "github.com/antoniomarcosferreira/gauditor/pkg/gauditor/redisstore"
)

rdb := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})
store := rs.New(rdb, rs.WithKeyPrefix("gauditor:"))
rec := gauditor.NewRecorder(store)
```

Postgres/MySQL (via database/sql):

```go
import (
  "database/sql"
  ss "github.com/antoniomarcosferreira/gauditor/pkg/gauditor/sqlstore"
  _ "github.com/lib/pq" // or mysql driver
)

db, _ := sql.Open("postgres", dsn)
store := ss.New(db)
_ = store.EnsureSchema(ctx)
rec := gauditor.NewRecorder(store)
```

S3 (append-only JSON objects):

```go
import (
  "github.com/aws/aws-sdk-go-v2/config"
  "github.com/aws/aws-sdk-go-v2/service/s3"
  s3s "github.com/antoniomarcosferreira/gauditor/pkg/gauditor/s3store"
)

cfg, _ := config.LoadDefaultConfig(ctx)
cli := s3.NewFromConfig(cfg)
store := s3s.New(cli, "my-bucket", "gauditor")
rec := gauditor.NewRecorder(store)
```

Single end-to-end flow:

```bash
# 1) Start the server
GO111MODULE=on go run ./cmd/gauditor

# 2) In another terminal, run the HTTP client sample
GO111MODULE=on go run ./examples/httpclient

# 3) Output will show created event JSON and the query results list
```

---

### Installation

Use as a library:

```bash
go get github.com/antoniomarcosferreira/gauditor
```

Install the server binary:

```bash
go install github.com/antoniomarcosferreira/gauditor/cmd/gauditor@latest
```

---

### Documentation

- Getting started & examples: `docs/Usage.md`
- Storage backends (Memory, Redis, SQL, S3): `docs/Storage.md`
- Developers guide (contrib, release, CI): `docs/Developers.md`

---

### HTTP API

- `POST /v1/events` — ingest an event (JSON body)
- `GET  /v1/events` — query events with optional filters: `tenant`, `actorId`, `action`, `targetId`, `limit`

OpenAPI spec: `api/openapi.yaml`

Event shape (response example):

```json
{
  "id": "3b6c2...",
  "timestamp": "2025-09-12T12:34:56Z",
  "tenant": "acme",
  "actor": {"id": "uuid", "attributes": {"type": "user", "email": "alice@example.com", "name": "Alice"}},
  "action": "login",
  "target": {"id": "", "type": "", "name": ""},
  "data": {"method": "password"}
}
```

---

### Docker

Build and run with Compose:

```bash
docker compose up --build -d
# server on localhost:8091
```

Or build the image directly:

```bash
docker build -t gauditor:dev .
```

---

### Development

- Format & lint: `make format` and `make lint`
- Tidy modules: `make tidy`
- Build all: `make build`
- Test all: `make test`
- Coverage: `make coverage` → opens dashboard at `docs/coverage.html`

#### Makefile commands

```bash
# install/update modules
make tidy

# format and lint
make format
make lint

# run tests
make test

# build binaries/libraries
make build

# generate coverage report and open the dashboard
make coverage
open docs/coverage.html

# watch and regenerate coverage (requires fswatch)
make coverage-watch
```

PRs should include tests where applicable. See `CONTRIBUTING.md`.

---

### Releases

Gauditor uses a simple, repeatable release process with a single source of truth for the version.

- Version file: `VERSION` (e.g., `v0.1.0-alpha.1`)
- Changelog: `CHANGELOG.md` (Keep a Changelog style)
- CI release workflow: `.github/workflows/release.yml` (runs on tag push)

CLI version:

```bash
gauditor -version
```

Make targets:

```bash
make release-print     # print current VERSION
make release-tag       # create and push a tag from VERSION (triggers release)
make build-versioned   # local build with ldflags injecting version
```

How to cut a release:

1) Update `VERSION` and `CHANGELOG.md`.
2) Commit and push changes.
3) Tag and push the tag:

```bash
make release-tag
```

What CI does on tag push:

- Builds binaries for Linux, macOS (arm64) and Windows
- Injects version via `-ldflags -X main.version=$(cat VERSION)`
- Produces `SHA256SUMS`
- Creates a GitHub Release with artifacts and autogenerated notes

### Roadmap (initial)

- Pluggable durable storage (PostgreSQL, S3/Parquet, SQLite)
- Structured masking/tokenization for sensitive fields
- gRPC API and client SDKs
- OTel integration and Grafana dashboards
- Multi-tenant auth and API keys
- Terraform modules for cloud deployment

Have ideas? Please open an issue to discuss.

---

### Contributing & Community

- Guidelines: `CONTRIBUTING.md`
- Code of Conduct: `CODE_OF_CONDUCT.md`
- CI: `.github/workflows/ci.yml`

We welcome issues, feature requests, and PRs. Star the repo to stay updated.

---

### License

MIT License. See `LICENSE`.
