# Architecture Guidelines

## Layered Architecture

```
routes/          → HTTP layer (handlers, request/response, middleware)
domain/          → Domain entities, value objects, business logic, repository interfaces
infra/rds/       → Data access, repository implementations (sqlc + PostgreSQL)
services/        → Cross-cutting business logic (diff, lint, stats, intent, action)
```

**Dependency direction**: routes → domain ← infra (domain is independent)

## Clean Architecture Principles

1. **Domain layer is pure**: No external dependencies (HTTP, DB)
2. **Handlers orchestrate**: Parse request → call repository → format response
3. **Repository pattern**: Interfaces in domain, implementations in infra
4. **Value objects**: Encapsulate validation and domain rules
5. **Dependency inversion**: Handlers depend on domain interfaces, not infra structs

## DI Wiring

All DI is wired in `cmd/main.go`:
```
db.Querier → NewXxxRepository(q) → NewXxxHandler(repo) → routes.Handlers → NewRouter(h)
db.Querier → NewXxxService(q)   → Handler (for diff/lint/stats)
```

## API Design

- RESTful endpoints: `/api/v1/{resource}`, `/api/v1/{resource}/{id}`
- Nested resources: `/api/v1/organizations/{org_id}/projects`
- Use Chi router with middleware (Logger, BearerAuth, Recoverer)
- JSON request/response format
- Graceful shutdown with context cancellation
- Bearer auth on all `/api/v1/*` routes

## Pagination

**All `FindAll`-style repository methods must accept pagination parameters** before being wired to production routers. Unbounded queries cause performance issues as datasets grow.

```go
// Required pattern for list endpoints
type ListParams struct {
    Limit  int32
    Offset int32
}

func (r *OrgRepository) FindAll(ctx context.Context, params ListParams) ([]organization.Organization, error)
```

## SSE Error Handling Pattern

SSE responses commit headers immediately (`Content-Type: text/event-stream`), so HTTP status is always 200 regardless of subsequent errors. **Set headers only after all pre-flight checks pass**:

```go
// CORRECT: fetch data before committing headers
func (h *StreamHandler) Stream() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        flusher, ok := w.(http.Flusher)
        if !ok {
            http.Error(w, "streaming not supported", http.StatusInternalServerError)
            return
        }

        // Pre-flight data fetch BEFORE writing headers
        messages, err := h.repo.FindAllBySession(r.Context(), sessionID)
        if err != nil {
            response.HandleError(w, err)  // can still set proper HTTP status
            return
        }

        // Only now commit SSE headers
        w.Header().Set("Content-Type", "text/event-stream")
        w.Header().Set("Cache-Control", "no-cache")
        w.Header().Set("Connection", "keep-alive")

        // Stream messages; errors after this point are SSE error events
        // Document clearly: HTTP status is 200 even on downstream errors
    }
}
```

If pre-flight restructuring is not feasible, add a comment: `// HTTP status is always 200 after this point; errors are sent as SSE error events`.

## DEV_BYPASS_RBAC Pattern

When `DEV_BYPASS_RBAC=true` is active:
1. **Must log a startup warning** so the bypass is visible in container logs
2. **Must inject a synthetic fixed userID** alongside the synthetic role, so `GetUserID(ctx)` returns a usable value — the zero UUID silently breaks downstream handlers that filter by user

```go
// In RequireRole middleware or cmd/main.go startup
if os.Getenv("DEV_BYPASS_RBAC") == "true" {
    logger.Warn("DEV_BYPASS_RBAC is enabled — all RBAC checks are skipped")
}

// When bypassing, inject both role AND userID
ctx = context.WithValue(ctx, roleKey, RoleOwner)
ctx = context.WithValue(ctx, userIDKey, devUserID)  // fixed well-known UUID
```

## Frontend Architecture (templ + HTMX)

- **Full pages**: Render complete HTML from templ templates
- **Partials**: HTMX requests return HTML fragments
- **No client-side JavaScript**: Use HTMX for interactivity
- **Type-safe templates**: templ compiles to Go code

### HTMX Swap Target Alignment

The partial response structure **must exactly match** the full-page `id` target. A common bug: the full page wraps multiple elements in `<div id="foo">` but the partial returns only a subset under the same `id`, silently dropping elements on swap.

```
// WRONG: full page has <div id="prompt-header"> containing name + badges + description
// but PromptHeaderUpdated returns only the name row with id="prompt-header"
// → badges and description are lost after HTMX swap

// CORRECT: partial must render the complete subtree that id="prompt-header" represents,
// or use hx-swap-oob to independently update each sub-element
```

### Safe hx-vals JSON

Never build `hx-vals` JSON by string concatenation. Use `templ.JSONString` or a format function:

```go
// WRONG — injection risk if nextStatus is ever a runtime value
hx-vals={ `{"status":"` + nextStatus + `"}` }

// CORRECT
hx-vals={ templ.JSONString(map[string]string{"status": nextStatus}) }
```

### CSS Organization

Inline CSS in every HTML response prevents browser caching and adds ~20KB per page. Extract stylesheets to `static/` and serve with cache headers:

```
// Preferred: <link rel="stylesheet" href="/static/primer.css">
// Avoid: <style> ... 800 lines ... </style> in layout.templ
```

### CDN Resource Integrity

External scripts (e.g., HTMX from unpkg.com) must include SRI hashes or be vendored into `static/`:

```html
<!-- Required: SRI hash prevents CDN compromise -->
<script src="https://unpkg.com/htmx.org@2.0.4"
        integrity="sha384-..."
        crossorigin="anonymous"></script>
```

## Database Migrations

- **Atlas**: Schema-first migrations in `apps/pkgs/db/migrations/`
- **sqlc**: Generate type-safe query code from `apps/pkgs/db/queries/*.sql`
- Migration workflow:
  1. Update schema or queries
  2. Run `make atlas-diff NAME=description`
  3. Review generated migration
  4. Run `make atlas-apply`
  5. Run `make sqlc-gen` if queries changed

## Error Handling Strategy

1. **Domain validation**: Return `ValidationError` via value object constructors
2. **Not found**: Return `NotFoundError` when resource doesn't exist
3. **Database errors**: Use `repoerr.Handle()` for consistent DB→domain error mapping
4. **HTTP mapping**: `response.HandleError(w, err)` maps `AppError` to HTTP status codes via `errors.As`
5. **Integer path params**: Wrap `strconv.Atoi`/`strconv.ParseInt` failures in `BadRequestError` so clients receive 400 not 500

## Security

- Validate all inputs using domain value objects
- Use parameterized queries (sqlc handles this)
- Environment-based configuration (no hardcoded secrets)
- Bearer auth middleware for API endpoints
- **Always use `url.Values` for query string construction** — never string concatenation (prevents broken URLs and injection on special characters)
- Add SRI hashes to any CDN resources in web templates

## Performance

- Connection pooling for database
- Context timeouts on all DB operations (5s default in repositories)
- Index database columns used in WHERE clauses
- All list endpoints must support pagination before production wiring
