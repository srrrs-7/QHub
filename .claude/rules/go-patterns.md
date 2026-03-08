# Go Functional Programming Patterns

## Result Type Monad

Always use `Result[T, E]` monad from `apps/pkgs/types/result.go` for error handling:

```go
// Use Pipe2-5 for sequential operations
res := types.Pipe2(
    validateInput(),
    func(input Input) types.Result[Output, model.AppError] {
        return processInput(input)
    },
    func(output Output) Response { return toResponse(output) },
)

// Use Match for handling success/error cases
res.Match(
    func(success Response) { /* handle success */ },
    func(err model.AppError) { /* handle error */ },
)
```

**Never use traditional Go (value, error) pattern in new code** - always return `Result[T, E]`.

## Domain Model Value Objects

- Create type-safe value objects for all domain entities
- Example: `TaskID`, `TaskTitle`, `TaskDescription`, `TaskCompleted`
- Validation logic belongs in value object constructors
- Value objects should be immutable

## Error Handling

All domain errors must implement `AppError` interface:

```go
type AppError interface {
    Error() string
    ErrorName() string
    DomainName() string
}
```

Use specific error types:
- `ValidationError` - Input validation failures
- `NotFoundError` - Resource not found
- `DatabaseError` - Database operation failures
- `ConflictError` - Business rule violations

## Concurrency with parallel Package

Use `parallel.Parallel2`-`Parallel5` for concurrent operations:

```go
result1, result2, err := parallel.Parallel2(ctx,
    func(ctx context.Context) (Type1, error) { /* operation 1 */ },
    func(ctx context.Context) (Type2, error) { /* operation 2 */ },
)
```

For key-sharded operations, use `parallel.KeyShard`:

```go
pool := parallel.NewKeyShard[Key, Result](workerCount, taskFunc)
defer pool.Close()
result := pool.Process(ctx, key)
```

## Code Organization

- **Handlers**: One file per endpoint (e.g., `list.go`, `post.go`, `get.go`, `put.go`)
- **Repository**: Use interface-based design in `infra/rds/`
- **Domain logic**: Keep in `domain/model/`
- **No business logic in handlers** - handlers orchestrate, domain logic validates and processes

## Naming Conventions

- Handlers: `{action}Handler` (e.g., `listHandler`, `createHandler`)
- Requests: `{action}Request` (e.g., `listRequest`, `createRequest`)
- Responses: `{action}Response` (e.g., `listResponse`, `createResponse`)
- Repositories: `{Entity}Repository` with methods like `Find{Entity}`, `Create{Entity}`
