# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

QHub — a system for creating questions (prompts) and version-managing answers. Go monorepo (Go 1.25, workspaces), PostgreSQL 18, Docker Compose.

## Development Commands

```bash
# Services
docker compose up -d          # Start all (api, web, db, cache, queue)
docker compose up -d db       # PostgreSQL only (required for tests)
make run-api                  # API on :8080
make run-web                  # Web on :3000 (run make templ-gen first)
make run-all                  # Migrate + API + Web
make build-cli                # Build qhub CLI to bin/qhub
make run-cli ARGS="prompt list --project <id>"  # Run CLI directly

# Quality & Tests
make check                    # fmt + vet + lint + cspell + test (full CI check)
make test                     # All tests (requires DB)
make fmt && make vet          # Format + static analysis
make lint                     # golangci-lint

# Single test
cd apps/api && go test -run TestGetHandler ./src/routes/tasks/

# Coverage
cd apps/api && go test -coverprofile=c.out ./... && go tool cover -func=c.out

# Database
make atlas-diff NAME=<name>   # Generate migration from schema diff
make atlas-apply              # Apply pending migrations
make sqlc-gen                 # Regenerate Go from SQL queries (after query changes)

# Frontend
make templ-gen                # Generate Go from .templ templates

# Terraform
make tf-fmt                   # Format .tf files
```

## Architecture

### Layered Clean Architecture

```
routes/  → HTTP handlers (depend on domain interfaces only)
domain/  → Value objects, entities, repository interfaces (no external deps)
infra/   → Repository implementations (sqlc + PostgreSQL)
```

Dependency direction: `routes → domain ← infra` (domain is independent).

DI wiring in `cmd/main.go`: `db.Querier → *Repository → *Handler → routes.Handlers → NewRouter`

### Domain Entities

| Entity | Domain Package | Repository Interface | Infra Package |
|--------|---------------|---------------------|---------------|
| Task | `domain/task/` | `TaskRepository` | `task_repository/` |
| Organization | `domain/organization/` | `OrganizationRepository`, `MemberRepository` | `organization_repository/` |
| Project | `domain/project/` | `ProjectRepository` | `project_repository/` |
| Prompt | `domain/prompt/` | `PromptRepository`, `VersionRepository` | `prompt_repository/` |

### API Routes (`/api/v1`, Bearer auth required)

```
/tasks                                    GET POST
/tasks/{id}                               GET PUT
/organizations                            POST
/organizations/{org_slug}                 GET PUT
/organizations/{org_id}/projects          GET POST
/organizations/{org_id}/projects/{slug}   GET PUT DELETE
/projects/{project_id}/prompts            GET POST
/projects/{project_id}/prompts/{slug}     GET PUT
/prompts/{prompt_id}/versions             GET POST
/prompts/{prompt_id}/versions/{version}   GET
/prompts/{prompt_id}/versions/{version}/status  PUT
/health                                   GET (no auth)
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

`response.HandleError(w, err)` maps `AppError` → HTTP status via `errors.As`. Repository errors use shared `repoerr.Handle()` for DB→domain error mapping.

### Struct-Based DI Pattern

```go
// Domain: interface defined here
type TaskRepository interface { FindByID(ctx, id) (Task, error) }

// Infra: struct implements interface
type TaskRepository struct { q db.Querier }
func NewTaskRepository(q db.Querier) *TaskRepository { ... }
var _ task.TaskRepository = (*TaskRepository)(nil)  // compile-time check

// Routes: handler depends on interface
type TaskHandler struct { repo task.TaskRepository }
func NewTaskHandler(repo task.TaskRepository) *TaskHandler { ... }
func (h *TaskHandler) Get() http.HandlerFunc { ... }
```

### Value Objects

Type-safe wrappers with validation in constructors: `TaskID`, `TaskTitle(3-100 chars)`, `TaskDescription(≤500)`, `TaskStatus(pending|completed)`, `OrganizationSlug`, `ProjectSlug`, `PromptSlug`, etc.

### Module Structure

Go workspace `apps/go.work` manages four modules:

- `apps/api` (module: `api`) — Backend API
- `apps/pkgs` (module: `utils`) — Shared: `db/`, `env/`, `logger/`, `testutil/`
- `apps/web` (module: `web`) — templ + HTMX frontend (M3 design system)
- `apps/cli` (module: `cli`) — `qhub` CLI (cobra-based, JSON output)

`apps/iac/` — Terraform AWS infrastructure (VPC, ECS, Aurora, Cognito, CloudFront, WAF, etc.)

### Database

- **Atlas**: Schema-first migrations (`apps/pkgs/db/migrations/`)
- **sqlc**: Type-safe queries (`apps/pkgs/db/queries/*.sql` → `apps/pkgs/db/db/*.go`)
- Workflow: edit schema/queries → `make atlas-diff NAME=x` → review → `make atlas-apply` → `make sqlc-gen`

## Coding Conventions

### TDD Mandatory (Red → Green → Refactor → Commit)

1. Write failing test first
2. Write minimal code to pass
3. Refactor
4. Coverage ≥80% overall, ≥80% per function, 100% critical paths

### Required Test Categories (all 6 for every function)

1. 正常系 (Happy Path)
2. 異常系 (Error Cases)
3. 境界値 (Boundary Values)
4. 特殊文字 (Special Chars: emoji, Japanese, SQL injection)
5. 空文字 (Empty/whitespace)
6. Null/Nil

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

Integration tests use `testutil.SetupTestTx(t)` for transaction-isolated DB. `testutil.SetAuthHeader(req)` for Bearer token.

### Handler File Layout (per resource)

`handler.go` (struct + constructor), `get.go`, `post.go`, `put.go`, `list.go`, `request.go`, `response_types.go`, `*_test.go`

### Git Commits

Conventional format: `feat:`, `fix:`, `refactor:`, `test:`, `docs:`, `chore:`
Include `Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>`

Detailed rules in `.claude/rules/`: `architecture.md`, `go-patterns.md`, `testing.md`, `tdd.md`
