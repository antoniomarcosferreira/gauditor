## Gauditor

The Go-powered audit trail solution: secure, simple, and scalable.

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
# server listening on :8080
```

Ingest an event:

```bash
curl -s -X POST localhost:8080/v1/events \
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
curl -s 'localhost:8080/v1/events?tenant=acme'
```

Run the example:

```bash
go run ./examples/basic
```

---

### Samples

- Go library sample: `examples/basic`
- HTTP client sample: `examples/httpclient`

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

### Library Usage

```go
import (
  "context"
  g "github.com/antoniomarcosferreira/gauditor/pkg/gauditor"
)

func recordLogin(ctx context.Context) error {
  rec := g.NewRecorder(g.NewMemoryStorage())
  _, err := rec.Record(ctx, g.Event{
    Tenant: "acme",
    Actor:  g.Actor{ID: "u123", Attributes: map[string]any{"type": "user", "email": "alice@example.com", "name": "Alice", "provider": "github"}},
    Action: "login",
    Data:   map[string]any{"method": "password"},
  })
  return err
}
```

---

### HTTP API

- `POST /v1/events` — ingest an event (JSON body)
- `GET  /v1/events` — query events with optional filters: `tenant`, `actorId`, `action`, `targetId`

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
# server on localhost:8080
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

PRs should include tests where applicable. See `CONTRIBUTING.md`.

---

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
