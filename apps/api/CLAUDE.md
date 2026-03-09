# CLAUDE.md - API Module

This file provides guidance to Claude Code when working with the `apps/api` module.

## Overview

Go backend API using chi router, serving RESTful JSON endpoints on port 8080. Clean architecture with layered structure.

## Structure

```
cmd/main.go           # Entry point, DI wiring, graceful shutdown
routes/
  routes.go           # Chi router setup, Handlers struct, /api/v1 prefix
  middleware/          # Bearer auth, logging, CORS, RBAC, rate limiting
  requtil/            # Generic request decoding + validation (go-playground/validator, bluemonday sanitization)
  response/           # JSON response helpers, AppError → HTTP status mapping
  tasks/              # TaskHandler: list.go, get.go, post.go, put.go, handler.go, response_types.go
  organizations/      # OrganizationHandler
  projects/           # ProjectHandler (nested under organizations)
  prompts/            # PromptHandler + version management
  consulting/         # ConsultingHandler + SSE streaming (sse.go, stream.go)
  analytics/          # AnalyticsHandler: project, prompt, version, trend endpoints
  search/             # SearchHandler: semantic search + embedding status
  apikeys/            # ApiKeyHandler: org-scoped API key management (SHA-256 hashing)
  members/            # MemberHandler: org membership CRUD + role management
domain/
  apperror/           # AppError interface, error types (Validation, NotFound, Database, etc.)
  task/               # Task entity, value objects (TaskID, TaskTitle, TaskDescription, TaskCompleted), repository interface
  organization/       # Organization entity, OrganizationSlug, repository interface
  project/            # Project entity, ProjectSlug, repository interface
  prompt/             # Prompt entity, PromptVersion, repository interfaces (PromptRepository, VersionRepository)
services/
  diffservice/        # Semantic diff between prompt versions (optional Redis cache)
  lintservice/        # Prompt quality linting (score 0-100, custom rules)
  ragservice/         # RAG pipeline: embed → search → context → Ollama stream + citations
  embeddingservice/   # Vector embedding generation + storage (TEI backend)
  intentservice/      # Rule-based intent classification for consulting chat (EN/JP)
  actionservice/      # Extract/execute actions from chat responses (create versions)
  statsservice/       # Welch's t-test for A/B version comparison
  batchservice/       # Monthly metric aggregation across organizations
  contentutil/        # Shared text extraction + {{variable}} detection
infra/rds/
  repoerr/            # Shared DB→domain error mapping: Handle(err, repoName, entity)
  task_repository/    # TaskRepository implementation (read.go, write.go)
  organization_repository/
  project_repository/
  prompt_repository/  # PromptRepository + VersionRepository implementations
```

## DI Wiring (cmd/main.go)

```
db.Querier → NewXxxRepository(q) → NewXxxHandler(repo) → routes.Handlers → NewRouter(h)
embedding.Client → EmbeddingService → PromptHandler, SearchHandler, RAGService
ollama.Client → RAGService → ConsultingHandler
db.Querier → NewAnalyticsHandler(q), NewApiKeyHandler(q), NewMemberHandler(q)
```

All wiring happens in `initHandlers()`. Handlers receive domain interfaces, not concrete types. RAG and embedding services are optional (enabled by `EMBEDDING_URL` and `OLLAMA_URI` env vars).

## Middleware Chain

Applied in order in `routes.go`: `RequestID → RealIP → Recoverer → Logger → CORS → BearerAuth → RateLimit`

- **CORS** (`middleware/cors.go`): Configurable via `CORS_ORIGINS` env var (comma-separated). Defaults to `*`.
- **RBAC** (`middleware/rbac.go`): Role-based access control with `RequireRole(q, minRole)`. Roles: owner > admin > member > viewer. Currently wired as TODO (awaiting JWT/Cognito).
- **Rate Limiting** (`middleware/ratelimit.go`): Token bucket per client (60 req/min, burst 10). Client identified by X-API-Key > Bearer token > IP.

## Key Patterns

### Handler Pattern

Each handler is a struct with a method per HTTP verb returning `http.HandlerFunc`:

```go
type TaskHandler struct { repo task.TaskRepository }
func NewTaskHandler(repo task.TaskRepository) *TaskHandler { ... }
func (h *TaskHandler) List() http.HandlerFunc { ... }
func (h *TaskHandler) Get() http.HandlerFunc { ... }
```

### Request Handling

Use `requtil.Decode[T](r)` for JSON body decoding with validation and sanitization. URL params via `chi.URLParam(r, "id")`.

### Response Handling

- `response.OK(w, body)`, `response.Created(w, body)`, `response.NoContent(w)`
- `response.HandleError(w, err)` maps AppError to HTTP status codes
- `response.MapSlice()` for converting domain slices to response types

### Repository Pattern

- Interfaces defined in `domain/{entity}/repository.go`
- Implementations in `infra/rds/{entity}_repository/`
- Compile-time check: `var _ task.TaskRepository = (*TaskRepository)(nil)`
- 5-second context timeout on all DB operations
- Use `repoerr.Handle()` for consistent error mapping

### Route Nesting

```
/api/v1/tasks, /api/v1/tasks/{id}
/api/v1/organizations, /api/v1/organizations/{org_slug}
/api/v1/organizations/{org_id}/projects, .../projects/{project_slug}
/api/v1/organizations/{org_id}/api-keys, .../api-keys/{id}
/api/v1/organizations/{org_id}/members, .../members/{user_id}
/api/v1/projects/{project_id}/prompts, .../prompts/{prompt_slug}
/api/v1/prompts/{prompt_id}/versions, .../versions/{version}
/api/v1/consulting/sessions/{session_id}/stream  (SSE)
/api/v1/search/semantic, /api/v1/search/embedding-status
/api/v1/projects/{project_id}/analytics
/api/v1/prompts/{prompt_id}/analytics, .../trend
```

### SSE Streaming (consulting)

`stream.go` streams session messages via Server-Sent Events. When RAG is enabled (`?rag=true&query=...&org_id=...`), generates AI responses using the RAG pipeline and streams chunks as `event: chunk` SSE events. Uses `sse.go` SSEWriter utility for event formatting and keepalive pings.

## Commands

```bash
# Run API server locally
make run-api                    # Port 8080

# Run tests (requires DB)
cd apps/api && go test ./...
cd apps/api && go test -run TestGetHandler ./src/routes/tasks/

# Build Docker image
docker compose up -d api
```

## Testing

- Handler tests use `testutil.SetupTestTx(t)` for transaction-isolated DB
- `testutil.SetAuthHeader(req)` sets Bearer token
- Chi URL params: create `chi.RouteContext()` and inject via `context.WithValue`
- Repository tests also use `testutil.SetupTestTx(t)`
- See `.claude/rules/testing.md` for patterns
