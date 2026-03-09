# CLAUDE.md - Shared Packages Module

This file provides guidance to Claude Code when working with the `apps/pkgs` module.

## Overview

Shared Go packages used by api, web, and cli modules. Contains database layer, external service clients, utilities, and test helpers.

## Packages

### db/ - Database Layer

```
db/
  connect.go          # NewConnection(connStr) (*sql.DB, error) with connection pooling
  db/                 # sqlc-generated code (models.go, querier.go, db.go, *.sql.go)
  queries/            # SQL query files for sqlc generation
  migrations/         # Atlas schema-first migration files
  atlas.hcl           # Atlas config with environments: local, docker, ci
  sqlc.yaml           # sqlc code generation config
```

- **sqlc** generates type-safe Go code from SQL queries into `db/db/`
- **Atlas** manages schema migrations in `db/migrations/`
- `db.Querier` interface is the main dependency injected into repositories
- Connection pooling configured in `connect.go`

### env/ - Environment Variables

Type-safe environment variable loading with `GetEnv(key, defaultValue)` and typed variants.

### logger/ - Structured Logging

`slog`-based structured logger initialization.

### testutil/ - Test Helpers

- `SetupTestTx(t)` - Creates transaction-wrapped DB connection that auto-rolls back on test cleanup
- `SetAuthHeader(req)` - Sets Bearer auth header for handler tests
- Requires running PostgreSQL (`docker compose up -d db`)

### ollama/ - Ollama LLM Client

HTTP client for the Ollama inference API (`ollama.Client`). Supports streaming (`Chat`) and synchronous (`ChatSync`) chat completions, plus health checks. Configured via `OLLAMA_URI` env var. Reports `Available() == false` when unconfigured.

### cache/ - Redis Cache Client

Wraps `go-redis/v9`. Nil-safe design: `nil` client is a no-op (cache is fully optional). `New(url)` returns `*Client`; methods: `Get()`, `Set()`, `Delete()`, `Available()`. Used by DiffService for caching diff results (24h TTL).

## Commands

```bash
# Database migrations
make atlas-diff NAME=description   # Generate migration from schema changes
make atlas-apply                   # Apply pending migrations
make atlas-status                  # Show migration status

# sqlc
make sqlc-gen                      # Generate Go code from SQL queries
make sqlc-compile                  # Validate SQL queries

# Tests
cd apps/pkgs && go test ./...
```

## Adding a New Query

1. Write SQL in `db/queries/{entity}.sql` with sqlc annotations
2. Run `make sqlc-gen` to generate Go code
3. Generated code appears in `db/db/` with type-safe methods on `Querier`

## Adding a Migration

1. Update schema in migration files or `schema.sql`
2. Run `make atlas-diff NAME=add_column_to_tasks`
3. Review generated migration in `db/migrations/`
4. Run `make atlas-apply`
