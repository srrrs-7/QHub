# Go Idiomatic Patterns

## Error Handling

Standard Go `(value, error)` returns with `if err != nil` pattern. Never use Result monads.

All domain errors implement `AppError` interface:

```go
type AppError interface {
    error
    ErrorName() string
    DomainName() string
    Unwrap() error
}
```

Use specific error types:
- `ValidationError` - Input validation failures
- `NotFoundError` - Resource not found
- `DatabaseError` - Database operation failures
- `ConflictError` - Business rule violations
- `BadRequestError` - Malformed requests
- `UnauthorizedError` - Authentication failures
- `ForbiddenError` - Authorization failures
- `InternalServerError` - Unexpected errors

Use `errors.As` for type checking:
```go
var appErr apperror.AppError
if errors.As(err, &appErr) {
    // handle AppError
}
```

Repository error mapping uses shared `repoerr.Handle(err, repoName, entity)`.

## Struct-Based Dependency Injection

Dependencies are injected via constructor functions. Interfaces defined in domain layer.

```go
// Repository struct with db.Querier
type TaskRepository struct { q db.Querier }
func NewTaskRepository(q db.Querier) *TaskRepository { return &TaskRepository{q: q} }
var _ task.TaskRepository = (*TaskRepository)(nil)  // compile-time check

// Handler struct with domain interface
type TaskHandler struct { repo task.TaskRepository }
func NewTaskHandler(repo task.TaskRepository) *TaskHandler { return &TaskHandler{repo: repo} }
```

## Domain Model Value Objects

- Create type-safe value objects for all domain entities
- Example: `TaskID`, `TaskTitle`, `TaskDescription`, `TaskStatus`, `OrganizationSlug`, `ProjectSlug`
- Validation logic belongs in value object constructors returning `(T, error)`
- Value objects should be immutable
- Use `XxxFromYyy` constructors for trusted sources (e.g., `TaskIDFromUUID` from DB)

## Code Organization

- **Handlers**: One file per endpoint (e.g., `list.go`, `post.go`, `get.go`, `put.go`)
- **Handler struct**: Defined in `handler.go` with constructor
- **Request types**: Defined in `request.go` with validation
- **Response types**: Defined in `response_types.go` with conversion functions
- **Repository**: Struct in `{entity}_repository.go`, reads in `read.go`, writes in `write.go`
- **Domain logic**: Keep in `domain/{entity}/` with repository interface in `repository.go`
- **No business logic in handlers** - handlers orchestrate, domain logic validates and processes

## Naming Conventions

- Handler struct: `{Entity}Handler` with methods `Get()`, `Post()`, `Put()`, `List()`, `Delete()`
- Requests: `{action}Request` (e.g., `getRequest`, `postRequest`)
- Responses: `{entity}Response` (e.g., `taskResponse`) with `to{Entity}Response()` converter
- Repositories: `{Entity}Repository` with methods like `FindByID`, `FindAll`, `Create`, `Update`
- Constructors: `New{Type}(deps) *{Type}`
