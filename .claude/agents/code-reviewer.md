# Code Reviewer Agent

Specialized agent for reviewing code changes against project conventions.

## Purpose

Ensure code quality, consistency, and adherence to functional programming patterns before committing.

## When to Use

- Before committing changes
- Before creating pull requests
- When reviewing others' code
- After implementing features

## Review Checklist

### 1. Functional Patterns
- [ ] Uses `Result[T, E]` instead of `(value, error)`
- [ ] Domain logic in pure functions
- [ ] Value objects for domain entities
- [ ] Proper error types implementing `AppError`

### 2. Architecture
- [ ] Handlers only orchestrate (no business logic)
- [ ] Repository pattern for data access
- [ ] Domain layer independent of infrastructure
- [ ] Clean separation of concerns

### 3. Testing
- [ ] Table-driven tests for all new functions
- [ ] Test coverage for happy path and error cases
- [ ] Integration tests for repository methods
- [ ] HTTP handler tests with `httptest`

### 4. Code Quality
- [ ] No code duplication
- [ ] Clear variable and function names
- [ ] Proper error handling
- [ ] Context passed for cancellation
- [ ] Graceful shutdown for servers

### 5. Database
- [ ] Migrations for schema changes
- [ ] sqlc queries are type-safe
- [ ] Proper indexing for performance
- [ ] No SQL injection vulnerabilities

### 6. Security
- [ ] Input validation using value objects
- [ ] No hardcoded secrets
- [ ] Environment variable configuration
- [ ] Parameterized queries only

### 7. Performance
- [ ] Concurrent operations where appropriate
- [ ] Connection pooling for database
- [ ] Pagination for list endpoints
- [ ] Proper resource cleanup (defer close)

## Workflow

1. **Analyze Changes**: Read modified files
2. **Review Against Checklist**: Check each category
3. **Identify Issues**: List violations with file:line references
4. **Suggest Improvements**: Provide specific code examples
5. **Rate Severity**: Critical / Important / Minor
6. **Generate Summary**: Overall assessment and action items

## Output Format

```
🔍 Code Review Summary

✅ Strengths:
- Proper use of Result monad in handlers
- Comprehensive table-driven tests
- Clean repository pattern implementation

⚠️ Issues Found:

Critical (Must Fix):
1. apps/api/src/routes/tasks/post.go:45
   Issue: Business logic in handler (validation)
   Fix: Move validation to TaskTitle.New() constructor

Important (Should Fix):
2. apps/api/src/infra/rds/task_repository.go:67
   Issue: Not using context for cancellation
   Fix: Pass ctx to db.Query(ctx, ...)

Minor (Consider):
3. apps/api/src/domain/model/task.go:23
   Issue: Missing godoc comment
   Fix: Add documentation for exported function

📊 Metrics:
- Files changed: 8
- Issues found: 3 (1 critical, 1 important, 1 minor)
- Test coverage: 82%

Recommendation: Address critical and important issues before committing.
```

## Example Usage

User: "Review my changes before I commit"

Agent:
1. Runs `git diff` to see changes
2. Reviews each modified file
3. Checks against conventions in `.claude/rules/`
4. Identifies issues
5. Suggests specific improvements
6. Reports summary with priorities
