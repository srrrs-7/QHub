# Code Style & Conventions

## Go Patterns
- **Error handling**: Standard `(value, error)` returns, `if err != nil` pattern
- **No Result monad**: Previously used, now removed in favor of idiomatic Go
- **Struct-based DI**: Dependencies injected via constructor functions (`NewXxx(deps) *Xxx`)
- **Interface in domain**: Repository interfaces defined in domain layer, implemented in infra
- **Compile-time checks**: `var _ Interface = (*Impl)(nil)` for interface satisfaction

## Naming
- Handlers: `func (h *TaskHandler) Get() http.HandlerFunc`
- Requests: `type getRequest struct` / `newGetRequest(r)`
- Responses: `type taskResponse struct` / `toTaskResponse(t)`
- Repositories: `type TaskRepository struct` with methods `FindByID`, `FindAll`, `Create`, `Update`
- Constructors: `NewTaskHandler(repo)`, `NewTaskRepository(q)`

## Domain
- Value objects: `TaskID`, `TaskTitle`, `TaskDescription`, `TaskStatus` (type aliases with validation)
- Domain errors: Implement `AppError` interface (`ErrorName()`, `DomainName()`, `Unwrap()`)
- Error types: `ValidationError`, `NotFoundError`, `DatabaseError`, `InternalServerError`

## File Organization
- One handler per file: `get.go`, `post.go`, `put.go`, `list.go`
- Handler struct in `handler.go`, response types in `response_types.go`
- Request types and validation in `request.go`
- Repository struct in `task_repository.go`, reads in `read.go`, writes in `write.go`

## Testing
- **TDD mandatory**: Red → Green → Refactor → Commit
- **Table-driven**: All tests use `testName/args/expected` structs
- **6 categories required**: 正常系, 異常系, 境界値, 特殊文字, 空文字, Null/Nil
- **Coverage**: ≥80% overall, ≥80% per function, 100% critical paths
- Integration tests use `testutil.SetupTestTx(t)` for DB transactions

## Commit Messages
- Conventional commits: `feat:`, `fix:`, `refactor:`, `test:`, `docs:`, `chore:`
- Include `Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>`
