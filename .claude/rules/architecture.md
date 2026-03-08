# Architecture Guidelines

## Layered Architecture

```
routes/          → HTTP layer (routing, request/response)
  handlers/      → Request handling, orchestration
domain/model/    → Domain entities, business logic, validation
infra/rds/       → Data access, repository implementations
```

**Dependency direction**: routes → domain ← infra (domain is independent)

## Clean Architecture Principles

1. **Domain layer is pure**: No external dependencies (HTTP, DB)
2. **Handlers orchestrate**: Validate request → call domain/repository → format response
3. **Repository pattern**: Abstract data access behind interfaces
4. **Value objects**: Encapsulate validation and domain rules

## API Design

- RESTful endpoints: `/api/v1/{resource}`, `/api/v1/{resource}/{id}`
- Use Chi router with middleware
- JSON request/response format
- Graceful shutdown with context cancellation

## Frontend Architecture (templ + HTMX)

- **Full pages**: Render complete HTML from templ templates
- **Partials**: HTMX requests return HTML fragments
- **No client-side JavaScript**: Use HTMX for interactivity
- **Type-safe templates**: templ compiles to Go code

## Database Migrations

- **Atlas**: Schema-first migrations
- **sqlc**: Generate type-safe query code
- Migration workflow:
  1. Update schema in `apps/pkgs/db/schema.sql`
  2. Run `make atlas-diff NAME=description`
  3. Review generated migration
  4. Run `make atlas-apply`
  5. Run `make sqlc-gen` if queries changed

## Error Handling Strategy

1. **Domain validation**: Return `ValidationError` with specific field errors
2. **Not found**: Return `NotFoundError` when resource doesn't exist
3. **Database errors**: Wrap with `DatabaseError` and log details
4. **HTTP mapping**: Use `response.HandleAppError()` to map domain errors to HTTP status codes

## Concurrency Patterns

- Use `parallel.Parallel2`-`Parallel5` for independent concurrent operations
- Use `parallel.KeyShard` for key-based work distribution
- Always pass `context.Context` for cancellation
- Handle graceful shutdown in main.go

## Security

- Validate all inputs using domain value objects
- Use parameterized queries (sqlc handles this)
- Environment-based configuration (no hardcoded secrets)
- CORS middleware for API endpoints

## Performance

- Use connection pooling for database
- Concurrent operations for independent tasks
- Index database columns used in WHERE clauses
- Use `LIMIT` and `OFFSET` for pagination
