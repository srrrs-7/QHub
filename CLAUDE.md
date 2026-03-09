# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

QHub — a prompt/answer version management system with consulting, execution logging, and prompt intelligence features. Go 1.26 monorepo with workspaces, PostgreSQL 18, Redis, ElasticMQ, Docker Compose.

## Development Commands

```bash
# Services (devcontainer starts db, cache, queue, embedding automatically)
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

# Ollama (host machine)
make ollama-health            # Check connectivity
make ollama-models            # List models
make ollama-embed TEXT="hello" # Generate embedding

# Terraform
make tf-fmt                   # Format .tf files
```

## Architecture

### Layered Clean Architecture (apps/api)

```
routes/    → HTTP handlers (depend on domain interfaces only)
domain/    → Value objects, entities, repository interfaces (no external deps)
services/  → Cross-cutting business logic (diff, lint, RAG, embeddings) using db.Querier directly
infra/     → Repository implementations (sqlc + PostgreSQL)
```

Dependency direction: `routes → domain ← infra`, `routes → services → domain`

Middleware chain (in `routes.go`): `RequestID → RealIP → Recoverer → Logger → CORS → BearerAuth → RateLimit`

DI wiring in `cmd/main.go`:
```
db.Querier → New*Repository(q) → New*Handler(repo) → routes.Handlers → NewRouter(h)
db.Querier → New*Service(q)   → Handler (for diff/lint/RAG/embedding)
embedding.Client → EmbeddingService → PromptHandler, SearchHandler, RAGService
ollama.Client → RAGService → ConsultingHandler
```

### Domain Entities

| Entity | Domain Package | Repository Interfaces | Infra Package |
|--------|---------------|----------------------|---------------|
| Task | `domain/task/` | `TaskRepository` | `task_repository/` |
| Organization | `domain/organization/` | `OrganizationRepository` | `organization_repository/` |
| Project | `domain/project/` | `ProjectRepository` | `project_repository/` |
| Prompt | `domain/prompt/` | `PromptRepository`, `VersionRepository` | `prompt_repository/` |
| ExecutionLog | `domain/executionlog/` | `LogRepository`, `EvaluationRepository` | `executionlog_repository/` |
| Consulting | `domain/consulting/` | `SessionRepository`, `MessageRepository`, `IndustryConfigRepository` | `consulting_repository/` |
| Tag | `domain/tag/` | `TagRepository` | `tag_repository/` |
| Intelligence | `domain/intelligence/` | (structs only — `SemanticDiff`, `LintResult`) | — |

### Services Layer

Services handle complex business logic that doesn't fit in repositories:

- **`services/diffservice/`**: Semantic diff between prompt versions (length, variables, tone, specificity analysis + LCS-based text diff)
- **`services/lintservice/`**: Prompt linting (excessive-length, output-format, variable-check, vague-instructions; score 0-100)
- **`services/ragservice/`**: RAG pipeline — embed query, search similar prompt versions, build context, stream LLM response via Ollama
- **`services/embeddingservice/`**: Generate and store vector embeddings for prompt versions (TEI backend); enables semantic search
- **`services/contentutil/`**: Shared text extraction from JSONB content and `{{variable}}` placeholder detection

Services receive `db.Querier` or domain interfaces and are wired in `cmd/main.go`. RAG and embedding services are optional (enabled by `OLLAMA_URI` and `EMBEDDING_URL` env vars).

### API Routes (`/api/v1`, Bearer auth)

```
/health                                          GET (no auth)
/tasks                                           GET POST
/tasks/{id}                                      GET PUT
/organizations                                   POST
/organizations/{org_slug}                        GET PUT
/organizations/{org_id}/projects                 GET POST
/organizations/{org_id}/projects/{project_slug}  GET PUT DELETE
/projects/{project_id}/prompts                   GET POST
/projects/{project_id}/prompts/{prompt_slug}     GET PUT
/prompts/{prompt_id}/versions                    GET POST
/prompts/{prompt_id}/versions/{version}          GET
/prompts/{prompt_id}/versions/{version}/status   PUT
/prompts/{prompt_id}/versions/{v1}/{v2}/diff     GET
/prompts/{prompt_id}/versions/{version}/lint     GET
/prompts/{prompt_id}/tags                        GET POST DELETE
/logs                                            GET POST
/logs/{id}                                       GET
/logs/batch                                      POST
/logs/{log_id}/evaluations                       GET POST
/evaluations                                     GET
/evaluations/{id}                                GET
/consulting/sessions                             GET POST
/consulting/sessions/{session_id}                GET
/consulting/sessions/{session_id}/messages       GET POST
/consulting/sessions/{session_id}/stream         GET (SSE)
/search/semantic                                 POST
/search/embedding-status                         GET
/analytics/projects/{project_id}                 GET
/analytics/prompts/{prompt_id}                   GET
/analytics/prompts/{prompt_id}/versions/{v}      GET
/analytics/prompts/{prompt_id}/trend             GET
/organizations/{org_id}/api-keys                 GET POST
/organizations/{org_id}/api-keys/{id}            DELETE
/organizations/{org_id}/members                  GET POST
/organizations/{org_id}/members/{user_id}        PUT DELETE
/tags                                            GET POST DELETE
/industries                                      GET POST
/industries/{slug}                               GET PUT
/industries/{slug}/compliance                    POST
/industries/{slug}/benchmarks                    GET
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

Go workspace `apps/go.work` manages five Go modules, plus SDK packages:

| Directory | Module Name | Description |
|-----------|-------------|-------------|
| `apps/api` | `api` | Backend API (chi, port 8080) |
| `apps/pkgs` | `utils` | Shared: `db/`, `env/`, `logger/`, `testutil/`, `ollama/`, `embedding/` |
| `apps/web` | `web` | templ + HTMX frontend (M3 design, port 3000) |
| `apps/cli` | `cli` | `qhub` CLI (cobra, JSON/table output) |
| `apps/sdk` | `sdk` | Go SDK module |
| `apps/sdk-python` | `qhub-sdk` | Python SDK (httpx + Pydantic v2, pip installable) |
| `apps/sdk-typescript` | `@qhub/sdk` | TypeScript SDK (native fetch, zero runtime deps) |

`apps/iac/` — Terraform AWS infrastructure (VPC, ECS, Aurora, Cognito, CloudFront, WAF)

### Database

- **Atlas**: Schema-first migrations (`apps/pkgs/db/migrations/`)
- **sqlc**: Type-safe queries (`apps/pkgs/db/queries/*.sql` → `apps/pkgs/db/db/*.go`)
- Workflow: edit schema/queries → `make atlas-diff NAME=x` → review → `make atlas-apply` → `make sqlc-gen`

### Handler File Layout (per resource)

`handler.go` (struct + constructor), `get.go`, `post.go`, `put.go`, `list.go`, `request.go`, `response_types.go`, `*_test.go`

### Infrastructure

Devcontainer includes: PostgreSQL 18, Redis, ElasticMQ, Text Embeddings Inference (TEI with `BAAI/bge-m3`).
Host Ollama accessible via `host.docker.internal:11434` (`OLLAMA_URI` env var).

### CI/CD

GitHub Actions in Dev Container (push/PR to main): `make vet → make atlas-apply → make test`

CD: Manual dispatch workflows (`cd-api.yml`, `cd-web.yml`, `cd-migrate.yml`) deploy to AWS ECS via OIDC auth. Environments: `dev` / `stg` / `prd`.

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
