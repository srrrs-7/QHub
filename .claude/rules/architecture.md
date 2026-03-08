# Architecture Guidelines

## Layered Architecture

```
routes/          → HTTP layer (handlers, request/response, middleware)
domain/          → Domain entities, value objects, business logic, repository interfaces
infra/rds/       → Data access, repository implementations (sqlc + PostgreSQL)
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
```

## API Design

- RESTful endpoints: `/api/v1/{resource}`, `/api/v1/{resource}/{id}`
- Nested resources: `/api/v1/organizations/{org_id}/projects`
- Use Chi router with middleware (Logger, BearerAuth, Recoverer)
- JSON request/response format
- Graceful shutdown with context cancellation
- Bearer auth on all `/api/v1/*` routes

## Frontend Architecture (templ + HTMX)

- **Full pages**: Render complete HTML from templ templates
- **Partials**: HTMX requests return HTML fragments
- **No client-side JavaScript**: Use HTMX for interactivity
- **Type-safe templates**: templ compiles to Go code

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

## Security

- Validate all inputs using domain value objects
- Use parameterized queries (sqlc handles this)
- Environment-based configuration (no hardcoded secrets)
- Bearer auth middleware for API endpoints

## Performance

- Connection pooling for database
- Context timeouts on all DB operations (5s default in repositories)
- Index database columns used in WHERE clauses
