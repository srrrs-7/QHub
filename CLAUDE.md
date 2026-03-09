# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

QHub — a prompt/answer version management system with consulting, execution logging, and prompt intelligence features. Go 1.26 monorepo with workspaces, PostgreSQL 18, Redis, ElasticMQ, Docker Compose.

## Development Commands

```bash
# Services (devcontainer starts db, cache, queue automatically)
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
routes/    → HTTP handlers (depend on domain interfaces only)
domain/    → Value objects, entities, repository interfaces (no external deps)
services/  → Cross-cutting business logic (diff, lint) using db.Querier directly
infra/     → Repository implementations (sqlc + PostgreSQL)
```

Dependency direction: `routes → domain ← infra`, `routes → services → domain`

Middleware chain (in `routes.go`): `RequestID → RealIP → Recoverer → Logger → CORS → BearerAuth → RateLimit`

DI wiring in `cmd/main.go`:
```
db.Querier → New*Repository(q) → New*Handler(repo) → routes.Handlers → NewRouter(h)
db.Querier → New*Service(q)   → Handler (for diff/lint)
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

- **`services/diffservice/`**: Semantic diff between prompt versions (length, variables, tone, specificity analysis + LCS-based text diff). Optional Redis caching (24h TTL).
- **`services/lintservice/`**: Prompt linting (excessive-length, output-format, variable-check, vague-instructions, missing-constraints, prompt-injection-risk; score 0-100). Supports custom regex rules via `LintWithCustomRules`.
- **`services/intentservice/`**: Rule-based intent classifier for consulting chat (EN/JP). 7 intent types: improve, compare, explain, create, compliance, best_practice, general.
- **`services/actionservice/`**: Extract and execute actions from consulting chat responses (e.g., create prompt versions from code blocks).
- **`services/statsservice/`**: Welch's t-test for A/B comparing prompt version metrics (latency, tokens, scores). Pure Go implementation (no external stats deps).
- **`services/batchservice/`**: Monthly metric aggregation across organizations for admin dashboard.
- **`services/contentutil/`**: Shared text extraction from JSONB content and `{{variable}}` placeholder detection.

Services receive `db.Querier` or domain interfaces and are wired in `cmd/main.go`. Diff service optionally uses Redis cache (`cache.Client`, nil-safe — nil client is no-op).

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
/prompts/{prompt_id}/versions/{v1}/{v2}/compare  GET (Welch's t-test)
/prompts/{prompt_id}/versions/{version}/lint     GET
/prompts/{prompt_id}/tags                        GET POST DELETE
/logs                                            GET POST
/logs/{id}                                       GET
/logs/batch                                      POST
/logs/{log_id}/evaluations                       GET POST
/evaluations                                     GET
/evaluations/{id}                                GET PUT
/consulting/sessions                             GET POST
/consulting/sessions/{session_id}                GET
/consulting/sessions/{session_id}/messages       GET POST
/consulting/sessions/{session_id}/stream         GET (SSE)
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
/admin/batch/aggregate                           POST
```

Log ingestion routes (`POST /logs`, `POST /logs/batch`) support combined auth: Bearer token or API key (`middleware/combined_auth.go`).

**Auth distinction**: API keys (`X-API-Key` header) and Bearer tokens (`Authorization: Bearer`) are different auth mechanisms. SDK constructors like `NewClient(bearerToken, ...)` do not transparently accept API keys — callers must use the correct auth path.

### Structured Logging (5W1H Schema)

Every log entry follows a unified 5W1H schema implemented in `apps/pkgs/logger/` and `middleware/logger.go`.

**JSON shape** (one entry per HTTP request, msg = `"http.request"`):
```json
{
  "time": "2026-03-09T15:56:51Z", "level": "INFO", "msg": "http.request",
  "who":   { "user_id": "uuid", "org_id": "uuid", "auth": "bearer|apikey|bypass", "ip": "…" },
  "what":  { "action": "POST /organizations/{org_id}/projects" },
  "when":  { "duration_ms": 42 },
  "where": { "layer": "http", "component": "Logger" },
  "why":   { "outcome": "success|client_error|server_error", "status": 201 },
  "how":   { "method": "POST", "path": "/api/v1/…", "route": "/api/v1/…/{org_id}/…",
             "request_id": "…", "user_agent": "…", "query": "" }
}
```

**`what.action`** uses the chi route *pattern* (e.g. `/organizations/{org_id}/projects`), not the actual URL, to keep cardinality safe for dashboards and alert rules.

**Log level** is determined by HTTP status: 2xx/3xx → `INFO`, 4xx → `WARN`, 5xx → `ERROR`. This lets alerting pipelines filter on `level` without parsing `why.outcome`.

**`RequestLog` pointer pattern** — the mechanism that populates WHO fields across the middleware chain without restructuring it:

```
Logger middleware
  └─ creates *RequestLog, stores in context
      └─ BearerAuth  → sets rl.AuthMethod = "bearer"
      └─ ApiKeyAuth  → sets rl.AuthMethod = "apikey", rl.OrgID = "…"
      └─ RequireRole → sets rl.OrgID (on org_id parse), rl.UserID (on userID parse),
                        even when the request is ultimately rejected (401/403)
  └─ after next.ServeHTTP returns, reads all WHO fields from the same pointer
```

The key property: `*RequestLog` is a **mutable pointer** stored in context. Every downstream `context.WithValue` call creates a new `*http.Request` but shares the same pointer, so mutations by any middleware are visible to the Logger after `next.ServeHTTP` returns. No mutex needed — one pointer per request.

**Non-HTTP log calls** (errors, startup, lifecycle) use `slog.Group` with the same 5W1H field names:
```go
// msg format: <domain>.<event>
logger.Error("rbac.membership_lookup_failed",
    slog.Group("who",   slog.String("user_id", …), slog.String("org_id", …)),
    slog.Group("where", slog.String("layer", "middleware"), slog.String("component", "RequireRole")),
    slog.Group("why",   slog.String("outcome", "error"), slog.String("error", err.Error())),
    slog.Group("how",   slog.String("request_id", rl.RequestID)),
)
```

**Event naming convention** — dot-separated `<domain>.<event>`:

| Domain | Events |
|--------|--------|
| `http` | `http.request` |
| `server` | `server.start`, `server.shutdown`, `server.exit`, `server.listen_failed`, `server.shutdown_failed` |
| `db` | `db.connect_failed`, `db.close_failed`, `db.config_missing` |
| `cache` | `cache.enabled`, `cache.disabled` |
| `auth` | `auth.token_validation_failed`, `auth.apikey_lookup_failed`, `auth.apikey_last_used_update_failed` |
| `rbac` | `rbac.membership_lookup_failed`, `rbac.bypass_enabled` |
| `health` | `health.check_failed` |

**Rules when adding new log calls:**
- Always use `slog.Group(dimension, slog.String(key, val), …)` — never bare key-value pairs
- Include `where.layer` + `where.component` so the log is traceable without reading code
- Include `how.request_id` on every in-request error log (read from `logger.RequestLogFrom(r.Context()).RequestID`)
- Use `logger.Warn` (not `Error`) for best-effort background tasks (e.g. `UpdateApiKeyLastUsed`)

### Known Pitfalls (from post-commit analysis)

These patterns have caused real bugs; avoid them:

| Pattern | Problem | Correct Approach |
|---------|---------|-----------------|
| `"/api/v1/tags?name=" + name` | Breaks on `&`, `=`, `#` in values | `url.Values{"name": {name}}.Encode()` |
| `strconv.Atoi` error not wrapped | Returns 500 on bad path param | Wrap in `apperror.NewBadRequestError(...)` → 400 |
| `data, _ := json.Marshal(x)` | Discards error, violates project convention | Always handle `json.Marshal` errors |
| `tt.field = value` inside test loop | Races under `t.Parallel()` | Use local variable, don't mutate `tt` |
| `if status != http.StatusOK` | Accepts wrong 5xx errors silently | Assert exact code: `cmp.Diff(http.StatusBadRequest, status)` |
| SSE headers before data fetch | Status always 200 even on errors | Fetch data first, then write headers |
| `os.Setenv` + `t.Parallel()` | Data race on global state | Use `t.Setenv` (no-parallel safe) |
| Committing `*.test` / `coverage.out` | Bloats repo, not source files | Add to `.gitignore` |

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
| `apps/pkgs` | `utils` | Shared: `db/`, `env/`, `logger/`, `testutil/`, `ollama/`, `cache/` |
| `apps/web` | `web` | templ + HTMX frontend (GitHub Primer design, port 3000) |
| `apps/cli` | `cli` | `qhub` CLI (cobra, JSON/table output) |
| `apps/sdk` | `sdk` | Go SDK module |
| `apps/sdk-python` | `qhub-sdk` | Python SDK (httpx + Pydantic v2, pip installable) |
| `apps/sdk-typescript` | `@qhub/sdk` | TypeScript SDK (native fetch, zero runtime deps) |

**SDK notes**: Go SDK requires a fully-qualified module path (e.g., `github.com/<org>/qhub/sdk`) to be `go get`-able externally. Python SDK dependencies are httpx + Pydantic v2 (not zero-dependency despite older docs).

`apps/iac/` — Terraform AWS infrastructure (VPC, ECS, Aurora, Cognito, CloudFront, WAF)

### Database

- **Atlas**: Schema-first migrations (`apps/pkgs/db/migrations/`)
- **sqlc**: Type-safe queries (`apps/pkgs/db/queries/*.sql` → `apps/pkgs/db/db/*.go`)
- Workflow: edit schema/queries → `make atlas-diff NAME=x` → review → `make atlas-apply` → `make sqlc-gen`

### Handler File Layout (per resource)

`handler.go` (struct + constructor), `get.go`, `post.go`, `put.go`, `list.go`, `request.go`, `response_types.go`, `*_test.go`

### Infrastructure

Devcontainer includes: PostgreSQL 18, Redis, ElasticMQ.

**Local dev environment variables** — automatically injected by the Makefile targets:

| Variable | Value | Where set |
|----------|-------|-----------|
| `DEV_BYPASS_RBAC=true` | Skips JWT/Cognito identity checks in `RequireRole` | `make run-api`, `make run-all` |
| `API_AUTH_TOKEN=dev-token` | Bearer token the web BFF sends to the API | `make run-web`, `make run-all` |

These are also set in `compose.override.yaml` for Docker-based dev. You do **not** need to set them manually when using `make run-*`.

**`DEV_BYPASS_RBAC` requirements** — when this flag is active, the middleware must:
1. Log a startup warning: `"DEV_BYPASS_RBAC is enabled — all RBAC checks are skipped"`
2. Inject **both** a synthetic role (`RoleOwner`) AND a fixed synthetic `userID` (`devBypassUserID`) into context — omitting `userID` causes zero-UUID in downstream handlers that call `GetUserID(ctx)`

### Web Frontend Architecture (apps/web)

templ + HTMX server-rendered frontend. Two handler types:
- **PageHandler**: Full page renders (18 handlers). Returns complete HTML with layout.
- **PartialHandler**: HTMX partial responses (33 handlers). Returns HTML fragments for dynamic updates.

Both depend on `client.Client` interface (not concrete `*APIClient`), enabling mock-based E2E testing.

**Critical HTMX swap rules**:
- Partial response `id` must cover the exact same DOM subtree as the full-page element with that `id`. Returning a narrower subtree silently drops elements (e.g., status badges lost after prompt edit).
- `hx-vals` JSON must use `templ.JSONString(...)`, never string concatenation — injection risk if values are ever non-literal.
- CDN scripts (`unpkg.com`) must include SRI hashes or be vendored into `static/`.

**Template helper functions** (e.g., `lintScoreClass`, `contentLines`, `availableTags` in `.templ` files) compile to regular Go and must have `_test.go` coverage with all 6 TDD categories.

**Testing pattern** (no DB required):
```go
mock := &client.MockClient{
    GetOrganizationFn: func(ctx context.Context, slug string) (*client.Organization, error) {
        return &client.Organization{ID: "org-1", Name: "Test", Slug: slug}, nil
    },
}
router := routes.NewRouter(mock)
w := httptest.NewRecorder()
router.ServeHTTP(w, httptest.NewRequest("GET", "/orgs/test/projects", nil))
// Assert: w.Code, w.Body.String() contains expected HTML
```

`MockClient` has sensible defaults for all methods. Override individual `Fn` fields. Use `client.NewMockClientWithError(err)` for error scenarios.

### CI/CD

GitHub Actions in Dev Container (push/PR to main): `make vet → make atlas-apply → make test`

CD: Manual dispatch workflows (`cd-api.yml`, `cd-web.yml`, `cd-migrate.yml`) deploy to AWS ECS via OIDC auth. Environments: `dev` / `stg` / `prd`.

## Coding Conventions

Detailed standards in `.claude/rules/`: `architecture.md`, `go-patterns.md`, `testing.md`, `tdd.md`

### TDD Mandatory (Red → Green → Refactor → Commit)

Coverage: ≥80% overall, ≥80% per function, 100% critical paths. Never write production code without writing the test first.

**All list (`FindAll`) repository methods require pagination parameters** before being wired to production routers. Unbounded queries are a performance time-bomb.

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
