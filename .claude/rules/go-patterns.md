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

### Integer Path Parameter Parsing

`strconv.Atoi` and `strconv.ParseInt` errors must be wrapped in `BadRequestError`, not left as raw errors (which produce 500):

```go
// WRONG — non-numeric input returns 500 InternalServerError
n, err := strconv.Atoi(chi.URLParam(r, "version"))
if err != nil {
    return err
}

// CORRECT — non-numeric input returns 400 BadRequestError
n, err := strconv.Atoi(chi.URLParam(r, "version"))
if err != nil {
    return apperror.NewBadRequestError("version must be an integer", "version", err)
}
```

### Never Discard json.Marshal Errors

Even when marshalling a simple `map[string]string` cannot realistically fail, discarding the error deviates from the project's explicit error-handling style:

```go
// WRONG — silently discards error
data, _ := json.Marshal(payload)

// CORRECT — handle even theoretically-unreachable errors
data, err := json.Marshal(payload)
if err != nil {
    // handle or log
}
```

## URL Query Parameter Construction

Always use `url.Values` for building query strings. String concatenation produces malformed URLs on special characters (`&`, `=`, `#`, space):

```go
// WRONG — breaks on tag names containing special characters
url := "/api/v1/tags?name=" + name

// CORRECT
params := url.Values{"name": {name}}
url := "/api/v1/tags?" + params.Encode()
```

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

## Web Client: Typed Request Structs

`APIClient` methods must use typed request structs rather than `map[string]string` or `map[string]any`. Untyped maps bypass value object validation at the client layer:

```go
// WRONG — no validation, allows missing required fields silently
client.CreatePrompt(ctx, map[string]string{"name": name, "slug": slug})

// CORRECT — typed struct with validation tags
type CreatePromptRequest struct {
    Name        string `json:"name" validate:"required,min=1,max=100"`
    Slug        string `json:"slug" validate:"required"`
    PromptType  string `json:"prompt_type" validate:"required"`
}
client.CreatePrompt(ctx, CreatePromptRequest{Name: name, Slug: slug})
```

## Web Client: Typed API Errors

The BFF `client` package must use a typed `APIError` struct that captures both the HTTP status code and parsed error body. Generic `fmt.Errorf("API error %d", status)` prevents callers from distinguishing 401 from 403 from 404:

```go
type APIError struct {
    StatusCode int
    Message    string
    ErrorName  string
}

func (e *APIError) Error() string {
    return fmt.Sprintf("API error %d: %s", e.StatusCode, e.Message)
}
```

## Domain Model Value Objects

- Create type-safe value objects for all domain entities
- Example: `TaskID`, `TaskTitle`, `TaskDescription`, `TaskStatus`, `OrganizationSlug`, `ProjectSlug`
- Validation logic belongs in value object constructors returning `(T, error)`
- Value objects should be immutable
- Use `XxxFromYyy` constructors for trusted sources (e.g., `TaskIDFromUUID` from DB)

### Slug Validation Consistency

Backend domain slug constructors and frontend HTML `pattern` attributes must use identical regex. When the slug pattern changes, update both. The canonical pattern is `[a-z0-9]+(-[a-z0-9]+)*` (allows single-char slugs; no leading, trailing, or consecutive hyphens).

Extract repeated slug regex to a shared constant rather than duplicating across templates.

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

## SDK Module Paths

SDK Go modules must use fully-qualified import paths (e.g., `github.com/<org>/qhub/sdk`), not bare names like `sdk`. Bare paths are not `go get`-able from outside the monorepo.

SDK constructors named `NewClient(bearerToken, ...)` must clearly document that API keys require different handling than Bearer tokens (different header: `X-API-Key` vs `Authorization: Bearer`). Do not silently accept an API key as a bearer token argument.
