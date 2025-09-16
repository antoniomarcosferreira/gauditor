# Guia do Desenvolvedor (Colaboradores)

Este documento é destinado a quem contribui com o projeto. O README principal foca no uso da biblioteca. Aqui você encontra arquitetura, como desenvolver, testar, versionar e publicar.

## Visão geral de arquitetura

- Pacote principal: `pkg/gauditor`
  - Tipos: `Event`, `Actor`, `Target`, `Query`
  - Núcleo: `Recorder` (valida, atribui defaults e persiste)
  - `Storage` (interface) + `MemoryStorage` (desenvolvimento)
- Armazenamentos pluggáveis:
  - `pkg/gauditor/redisstore`: Redis (lista por tenant; simples, ideal para demos)
  - `pkg/gauditor/sqlstore`: `database/sql` (Postgres/MySQL) com prefixo de tabela configurável
  - `pkg/gauditor/s3store`: S3 (objetos JSON append-only)
- Bootstrap por ambiente: `pkg/gauditorenv` (constrói `Recorder` via variáveis de ambiente)
- Servidor HTTP de exemplo: `cmd/gauditor` (ingestão e consulta REST)
- Exemplos: `examples/basic`, `examples/httpclient`, `examples/gincrud` (com middleware), `examples/redis`

## Fluxo de desenvolvimento

Pré-requisitos: Go 1.23+

```bash
make tidy          # atualizar módulos
make build         # compilar tudo
make test          # rodar testes
make coverage      # gerar docs/coverage.html
```

Lint e formatação:

```bash
make format
make lint
```

Cobertura (dashboard em `docs/coverage.html`):

```bash
make coverage
open docs/coverage.html
```

## Variáveis de ambiente (execução e configuração)

- Servidor HTTP: `GAUDITOR_ADDR` (ex.: `:8091`)
- Seleção de storage: `GAUDITOR_STORAGE` = `memory` | `redis` | `sql` | `s3`
- Redis: `REDIS_ADDR` (default `127.0.0.1:6379`), `REDIS_KEY_PREFIX` (default `gauditor:`)
- SQL: `SQL_DRIVER` (`postgres`|`mysql`), `SQL_DSN`, `GAUDITOR_SQL_ENSURE_SCHEMA=1`
- S3: `S3_BUCKET`, `S3_PREFIX` (default `gauditor`) + `AWS_*` (credenciais/região)

Para criar um `Recorder` a partir do ambiente:

```go
rec, err := gauditorenv.NewRecorderFromEnv(context.Background())
```

## Convenções de código

- Go idiomático, claro e testável
- `context.Context` para operações request-scoped
- Nomeação explícita e significativa (evitar abreviações obscuras)
- Evitar capturar e ignorar erros sem tratamento
- Expor opções via padrões `Option` quando fizer sentido (ex.: `WithClock`, `WithIDGenerator`)

## Armazenamentos

- `MemoryStorage`: seguro para dev/testes, com cancelamento de contexto
- `redisstore`: simples; filtra em aplicação; bom para demos
- `sqlstore`:
  - Tabela padrão: `gauditor_events`
  - Suporte a prefixo/nome de tabela: `WithTablePrefix("app_")` ou `WithTableName("minha_tabela")`
  - `EnsureSchema(ctx)` cria tabela e índice `idx_<tabela>_tenant_ts`
- `s3store`: grava um objeto JSON por evento; `Query` lista e filtra cliente; ordena por timestamp

## Middleware para Gin (exemplo)

Veja `examples/gincrud`. O middleware registra automaticamente requisições bem-sucedidas (2xx/3xx), derivando a ação de `METHOD + rota` e `Target.ID` de `:id` quando existir.

## Segurança

- CI executa `govulncheck` em cada push/PR para detectar vulnerabilidades conhecidas
- Handlers HTTP usam `DisallowUnknownFields`, limite de corpo (1MB) e timeouts conservadores

## Publicação de docs (pkg.go.dev)

- O projeto possui `doc.go` no pacote principal e subpacotes para documentação
- Para indexar: torne o repositório público, crie uma tag e acesse os caminhos no `pkg.go.dev`

## Versionamento e releases

- Versão única em `VERSION` (ex.: `v0.0.1`)
- Notas de versão em `CHANGELOG.md` (Keep a Changelog)
- Flag do CLI para versão: `gauditor -version`

Automação:

- Workflow de release: `.github/workflows/release.yml` (dispara ao criar tag `v*`)
- Artefatos: binários Linux/macOS/Windows + `SHA256SUMS`
- Alvo do Makefile:

```bash
make release-print     # mostra a versão atual
make release-tag       # cria e push da tag (dispara o release)
make build-versioned   # build local com ldflags injetando a versão
```

Como cortar um release:

1) Atualize `VERSION` e `CHANGELOG.md`
2) Commit e push
3) Rode `make release-tag`

## CI/CD

- `ci.yml`: testes com cobertura, `golangci-lint`, `govulncheck`
- `pages.yml`: publica `docs/` (coverage dashboard) no GitHub Pages
- `release.yml`: builds e publicação de release sob tags `v*`

## Contribuindo

- Leia `CONTRIBUTING.md` e o Código de Conduta
- Abra issues para discutir novas features/bugs
- Pull Requests com testes e docs são bem-vindos
