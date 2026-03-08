# CLAUDE.md - API Module

This file provides guidance to Claude Code when working with the `apps/api` module.

## Overview

Go backend API using chi router, serving RESTful JSON endpoints on port 8080. Clean architecture with layered structure.

## Structure

```
cmd/main.go           # Entry point, DI wiring, graceful shutdown
routes/
  routes.go           # Chi router setup, Handlers struct, /api/v1 prefix
  middleware/          # Bearer auth, logging
  requtil/            # Generic request decoding + validation (go-playground/validator, bluemonday sanitization)
  response/           # JSON response helpers, AppError → HTTP status mapping
  tasks/              # TaskHandler: list.go, get.go, post.go, put.go, handler.go, response_types.go
  organizations/      # OrganizationHandler
  projects/           # ProjectHandler (nested under organizations)
  prompts/            # PromptHandler + version management
domain/
  apperror/           # AppError interface, error types (Validation, NotFound, Database, etc.)
  task/               # Task entity, value objects (TaskID, TaskTitle, TaskDescription, TaskCompleted), repository interface
  organization/       # Organization entity, OrganizationSlug, repository interface
  project/            # Project entity, ProjectSlug, repository interface
  prompt/             # Prompt entity, PromptVersion, repository interfaces (PromptRepository, VersionRepository)
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
```

All wiring happens in `initHandlers()`. Handlers receive domain interfaces, not concrete types.

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
/api/v1/tasks
/api/v1/tasks/{id}
/api/v1/organizations
/api/v1/organizations/{org_slug}
/api/v1/organizations/{org_slug}/projects
/api/v1/organizations/{org_slug}/projects/{project_slug}
/api/v1/organizations/{org_slug}/projects/{project_slug}/prompts
/api/v1/organizations/{org_slug}/projects/{project_slug}/prompts/{prompt_id}
/api/v1/organizations/{org_slug}/projects/{project_slug}/prompts/{prompt_id}/versions
```

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
