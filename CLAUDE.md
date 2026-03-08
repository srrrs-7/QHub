# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

QHub — a prompt/answer version management system. Go 1.26 monorepo with workspaces, PostgreSQL 18, Redis, ElasticMQ, Docker Compose.

## Development Commands

```bash
# Services
docker compose up -d          # Start all (api, web, db, cache, queue)
docker compose up -d db       # PostgreSQL only (required for tests)
make run-api                  # API on :8080
make run-web                  # Web on :3000 (run make templ-gen first)
make run-all                  # Migrate + API + Web
make build-cli                # Build qhub CLI to bin/qhub
make run-cli ARGS="prompt list --project <id>"

# Quality & Tests
make check                    # fmt + vet + lint + cspell + test (full CI check)
make test                     # All tests across all modules (requires DB, runs atlas-apply first)
make fmt && make vet          # Format + static analysis
make lint                     # golangci-lint

# Single test
cd apps/api && go test -run TestGetHandler ./src/routes/tasks/

# Coverage
cd apps/api && go test -coverprofile=c.out ./... && go tool cover -func=c.out

# Database
make atlas-diff NAME=<name>   # Generate migration from schema diff
make atlas-apply              # Apply pending migrations (ATLAS_ENV=docker for Docker)
make atlas-status             # Show migration status
make sqlc-gen                 # Regenerate Go from SQL queries

# Frontend
make templ-gen                # Generate Go from .templ templates
make templ-watch              # Watch mode for development

# Terraform
make tf-fmt                   # Format .tf files
```

## Architecture

### Layered Clean Architecture (apps/api)

```
routes/  → HTTP handlers (depend on domain interfaces only)
domain/  → Value objects, entities, repository interfaces (no external deps)
infra/   → Repository implementations (sqlc + PostgreSQL)
```

Dependency direction: `routes → domain ← infra` (domain is independent).

DI wiring in `cmd/main.go`:
```
db.Querier → New*Repository(q) → New*Handler(repo) → routes.Handlers → NewRouter(h)
```

### Domain Entities

| Entity | Domain Package | Repository Interface | Infra Package |
|--------|---------------|---------------------|---------------|
| Task | `domain/task/` | `TaskRepository` | `task_repository/` |
| Organization | `domain/organization/` | `OrganizationRepository` | `organization_repository/` |
| Project | `domain/project/` | `ProjectRepository` | `project_repository/` |
| Prompt | `domain/prompt/` | `PromptRepository`, `VersionRepository` | `prompt_repository/` |

### API Routes (`/api/v1`, Bearer auth)

```
/health                                   GET (no auth)
/tasks                                    GET POST
/tasks/{id}                               GET PUT
/organizations                            POST
/organizations/{org_slug}                 GET PUT
/organizations/{org_id}/projects          GET POST
/organizations/{org_id}/projects/{project_slug}  GET PUT DELETE
/projects/{project_id}/prompts            GET POST
/projects/{project_id}/prompts/{prompt_slug}     GET PUT
/prompts/{prompt_id}/versions             GET POST
/prompts/{prompt_id}/versions/{version}   GET
/prompts/{prompt_id}/versions/{version}/status   PUT
```

### Error Handling

Standard Go `(value, error)` returns. Domain errors implement `AppError` interface (`domain/apperror/`):

```go
type AppError interface {
    error
    ErrorName() string   // "ValidationError", "NotFoundError", etc.
    DomainName() string  // "Task", "Organization", etc.
    Unwrap() error
}
```

Error types: `ValidationError`, `NotFoundError`, `DatabaseError`, `InternalServerError`, `BadRequestError`, `ConflictError`, `UnauthorizedError`, `ForbiddenError`.

HTTP mapping: `response.HandleError(w, err)` uses `errors.As` to map `AppError` → HTTP status. Repository errors use `repoerr.Handle(err, repoName, entity)` for DB→domain error mapping.

### Struct-Based DI Pattern

```go
// Domain: interface
type TaskRepository interface { FindByID(ctx, id) (Task, error) }

// Infra: implementation
type TaskRepository struct { q db.Querier }
var _ task.TaskRepository = (*TaskRepository)(nil)  // compile-time check

// Routes: handler depends on interface
type TaskHandler struct { repo task.TaskRepository }
func (h *TaskHandler) Get() http.HandlerFunc { ... }
```

### Request Handling (requtil package)

- `requtil.Decode[T](r, sanitizeFn)` — JSON decode + bluemonday sanitization + go-playground/validator
- `requtil.ParseUUID(r, "id")` — extract chi URL param as UUID
- `requtil.ValidateParams[T](v)` — validate a struct of URL params
- `requtil.MergeField(existing, raw, constructor)` — partial update helper (keep existing if raw is empty)

### Value Objects

Type-safe wrappers with validation in constructors returning `(T, error)`:
`TaskTitle(3-100 chars)`, `TaskDescription(≤500)`, `TaskStatus(pending|completed)`, `OrganizationSlug`, `ProjectSlug`, `PromptSlug`, etc.

`XxxFromYyy` constructors for trusted sources (e.g., `TaskIDFromUUID` from DB reads).

### Module Structure

Go workspace `apps/go.work` manages four modules:

| Directory | Module Name | Description |
|-----------|-------------|-------------|
| `apps/api` | `api` | Backend API (chi, port 8080) |
| `apps/pkgs` | `utils` | Shared: `db/`, `env/`, `logger/`, `testutil/` |
| `apps/web` | `web` | templ + HTMX frontend (M3 design, port 3000) |
| `apps/cli` | `cli` | `qhub` CLI (cobra, JSON/table output) |

`apps/iac/` — Terraform AWS infrastructure (VPC, ECS, Aurora, Cognito, CloudFront, WAF)

### Database

- **Atlas**: Schema-first migrations (`apps/pkgs/db/migrations/`)
- **sqlc**: Type-safe queries (`apps/pkgs/db/queries/*.sql` → `apps/pkgs/db/db/*.go`)
- Workflow: edit schema/queries → `make atlas-diff NAME=x` → review → `make atlas-apply` → `make sqlc-gen`

### Handler File Layout (per resource)

`handler.go` (struct + constructor), `get.go`, `post.go`, `put.go`, `list.go`, `request.go`, `response_types.go`, `*_test.go`

## Coding Conventions

Detailed standards in `.claude/rules/`: `architecture.md`, `go-patterns.md`, `testing.md`, `tdd.md`

### TDD Mandatory (Red → Green → Refactor → Commit)

Coverage: ≥80% overall, ≥80% per function, 100% critical paths. Never write production code without writing the test first.

### Required Test Categories (all 6 for every function)

1. 正常系 (Happy Path) — valid inputs
2. 異常系 (Error Cases) — invalid inputs
3. 境界値 (Boundary Values) — min/max, zero
4. 特殊文字 (Special Chars) — emoji, Japanese, SQL injection
5. 空文字 (Empty/whitespace)
6. Null/Nil — nil pointers, zero values

### Table-Driven Tests

```go
tests := []struct {
    testName string
    args     args
    expected expected
}{...}
for _, tt := range tests {
    t.Run(tt.testName, func(t *testing.T) { ... })
}
```

Integration tests: `testutil.SetupTestTx(t)` for transaction-isolated DB (auto-rollback). `testutil.SetAuthHeader(req)` for Bearer token.

### Git Commits

Conventional format: `feat:`, `fix:`, `refactor:`, `test:`, `docs:`, `chore:`
Include `Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>`
