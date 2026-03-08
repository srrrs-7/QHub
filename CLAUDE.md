# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Go monorepo using Go 1.25 with workspaces, Docker Compose for services, and PostgreSQL 18.

## Coding Guidelines

**CRITICAL: This repository requires Test-Driven Development (TDD) for all code changes.**

Detailed coding standards and patterns are documented in `.claude/rules/`:

- **architecture.md**: Layered architecture (routes → domain ← infra), clean architecture principles, API design, database migrations, error handling strategy
- **go-patterns.md**: Domain model value objects, error handling with `AppError` interface, concurrency patterns
- **testing.md**: Table-driven test pattern (mandatory), HTTP handler testing with `testutil`, database testing, test organization
- **tdd.md**: **TDD workflow (Red → Green → Refactor → Commit)**, required test categories (6 categories: 正常系, 異常系, 境界値, 特殊文字, 空文字, Null/Nil), coverage requirements (≥80% overall, ≥80% per function, 100% for critical paths)

**TDD is mandatory. Always follow this workflow:**

1. 🔴 **RED**: Write failing test first
2. 🟢 **GREEN**: Write minimal code to pass
3. 🔵 **REFACTOR**: Improve code quality
4. ✅ **COMMIT**: Commit after green with ≥80% coverage
5. ♻️ **REPEAT**: Next test case

**Never write production code without writing the test first.** Coverage below 80% blocks commits.

## Repository Structure

```
apps/
  api/          # Backend API (go-chi router, port 8080)
  pkgs/         # Shared packages (db, logger, env, testutil)
  web/          # Frontend (templ + HTMX, port 3000)
  migrate/      # Database migration app
  iac/          # Terraform infrastructure (AWS)
```

Go workspace: `apps/go.work` manages `api`, `pkgs`, and `web` modules.

## Development Commands

```bash
# Local services (from repo root)
docker compose up -d          # Start all services (api, web, db, cache, queue)
docker compose up -d db       # Start only PostgreSQL

# Run servers locally (alternative to Docker Compose)
make run-api                  # Run API server on port 8080
make run-web                  # Run web server on port 3000 (requires make templ-gen first)
make run-all                  # Run migrations, then start API and web servers

# Tests (all modules)
make test                     # Run all tests (requires DB running)
make check                    # Run fmt, vet, lint, cspell, and test

# Run single test
cd apps/api && go test -run TestListHandler ./src/routes/tasks/

# Code quality
make fmt      # Format Go code
make vet      # Static analysis
make lint     # Run golangci-lint
make fix      # Run go fix
make tidy     # go mod tidy
make cspell   # Check spelling (misspell)
make tf-fmt   # Format Terraform files

# Database migrations (Atlas)
make atlas-diff NAME=<name>  # Generate migration from schema changes
make atlas-apply             # Apply pending migrations
make atlas-new NAME=<name>   # Create new migration file
make atlas-status            # Show migration status
# Use ATLAS_ENV=docker for Docker environment: make atlas-apply ATLAS_ENV=docker

# sqlc code generation
make sqlc-gen      # Generate Go code from SQL queries
make sqlc-compile  # Validate SQL queries

# templ template generation (web frontend)
make templ-gen     # Generate Go code from .templ templates
make templ-watch   # Watch and regenerate on template changes
make templ-fmt     # Format .templ files

# Git hooks (optional)
make hooks-install   # Install pre-commit (fmt, vet) and pre-push (test) hooks
make hooks-uninstall # Remove git hooks
```

## Architecture

### Error Handling

The codebase uses standard Go `(value, error)` returns. Domain errors implement the `AppError` interface from `domain/apperror/`:

```go
type AppError interface {
    Error() string
    ErrorName() string
    DomainName() string
}
```

Error types: `ValidationError`, `NotFoundError`, `DatabaseError`, `ConflictError`.
The response layer maps `AppError` to HTTP status codes via `response.HandleAppError()`.

### Domain Model Pattern

- Value objects with type safety: `TaskID`, `TaskTitle`, `TaskDescription`, `TaskStatus`
- Validation in value object constructors (e.g., `NewTaskTitle` validates length 3-100)
- Repository interfaces defined in domain layer (dependency inversion)
- Domain layer has no external dependencies

```go
// domain/task/repository.go — interface in domain, implemented by infra
type TaskRepository interface {
    FindByID(ctx context.Context, id TaskID) (Task, error)
    FindAll(ctx context.Context) ([]Task, error)
    Create(ctx context.Context, cmd TaskCmd) (Task, error)
    Update(ctx context.Context, id TaskID, cmd TaskCmd) (Task, error)
}
```

### API Layer Structure (apps/api/src)

```
cmd/main.go              # Entry point, graceful shutdown
routes/
  routes.go              # Chi router setup, /api/v1 prefix, DI wiring
  middleware/             # Logger, Bearer auth middleware
  response/response.go   # JSON response helpers, error mapping
  tasks/                 # Handler per endpoint (list.go, post.go, get.go, put.go)
domain/
  apperror/              # AppError interface and error types
  task/                  # Value objects, aggregate, repository interface
infra/rds/
  task_repository/       # Repository implementation using sqlc
```

DI flow in `routes.go`: sqlc Querier → `TaskRepository` → `TaskHandler`

Route pattern: `/api/v1/tasks`, `/api/v1/tasks/{id}` (Bearer auth required)

### Web Frontend (apps/web)

Go server-side rendered frontend using templ + HTMX:
- `cmd/main.go` - Entry point (port 3000), graceful shutdown, API client initialization
- `templates/*.templ` - Type-safe Go templates compiled to `*_templ.go` files
- `handlers/` - HTTP handlers returning templ components
- `routes/` - Chi router, serves full pages and HTMX partials
- `client/` - API client for communicating with backend API

**templ workflow**: Edit `.templ` files → `make templ-gen` → compile-time checked Go code.

### Database Layer (apps/pkgs/db)

- **Atlas**: Schema-first migrations in `migrations/`
- **sqlc**: Type-safe query generation from `queries/*.sql` → `db/*.go`
- Config files: `atlas.hcl` (environments: local, docker, ci), `sqlc.yaml`
- Migration workflow: Update schema → `make atlas-diff` → review → `make atlas-apply` → `make sqlc-gen`

### Shared Packages (apps/pkgs)

- `db/` - Database connection (`connect.go`) and sqlc-generated queries
- `env/` - Environment variable utilities (`GetString`, `GetStringOrDefault`, `GetInt`, `GetBool`)
- `logger/` - Structured logging with slog (JSON output)
- `testutil/` - Test helpers: `SetupTestTx()` for transaction-wrapped DB tests, `SetAuthHeader()` for Bearer token in HTTP tests

### Docker Compose Services

- **api** (port 8080) - Backend, depends on db/cache/queue
- **web** (port 3000) - Frontend, depends on api
- **db** - PostgreSQL 18 (port 5432)
- **cache** - Redis (port 6379)
- **queue** - ElasticMQ (port 9324)
- **migrate** - Database migration runner

## Testing Patterns

**All code must follow TDD workflow** (see `.claude/rules/tdd.md` for details).

Tests use table-driven pattern with `args`/`expected` structs. HTTP handlers tested with `httptest.NewRequest` and `httptest.NewRecorder`.

**Required test coverage**: ≥80% overall, ≥80% per function, 100% for critical paths.

**Required test categories** (cover all 6 for every function):
1. ✅ 正常系 (Happy Path)
2. ❌ 異常系 (Error Cases)
3. 📏 境界値 (Boundary Values)
4. 🔤 特殊文字 (Special Chars)
5. 📭 空文字 (Empty String)
6. ⚠️ Null/Nil

## Infrastructure (apps/iac)

Terraform modules for AWS: VPC, ECS Fargate, ECR, ALB, Aurora PostgreSQL Serverless v2, Cognito, CloudFront, S3, WAF, ACM, Route53, IAM (OIDC), Security Groups.

Naming convention: `${project}-${environment}-${resource}`

## Git Workflow

- Conventional commits: `feat:`, `fix:`, `refactor:`, `test:`, `docs:`, `chore:`
- Include `Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>`
- Stage specific files (avoid `git add -A`)
- If pre-commit hook fails: fix and create NEW commit (never `--amend`)
- Never force push to `main`; use `--force-with-lease` if needed

## CI/CD Pipeline

**CI**: GitHub Actions runs in devcontainer on push/PR to main: `make vet && make atlas-apply && make test`

**CD**: Triggered by push to `main` (→ dev) or manual workflow dispatch (→ dev/stg/prd)
- Flow: Database Migration → Build & Push to ECR → Deploy to ECS
- Environments configured in GitHub Settings → Environments with AWS OIDC credentials
