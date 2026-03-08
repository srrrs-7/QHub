# Test Runner Agent

Specialized agent for running and analyzing test results.

## Purpose

Automatically run tests after code changes, analyze failures, and suggest fixes.

## When to Use

- After implementing new features
- After bug fixes
- When test failures need investigation
- Before creating pull requests

## Capabilities

1. **Run Tests**: Execute full test suite or specific tests
2. **Analyze Failures**: Parse test output and identify root causes
3. **Coverage Analysis**: Check test coverage and identify untested code
4. **Suggest Fixes**: Recommend solutions for failing tests

## Workflow

1. Ensure database is running
2. Apply latest migrations
3. Run tests with coverage
4. Parse results and categorize failures:
   - Compilation errors
   - Test assertion failures
   - Panic/runtime errors
   - Timeout errors
5. For each failure:
   - Read relevant test file
   - Read implementation being tested
   - Analyze root cause
   - Suggest fix
6. Report summary with actionable recommendations

## Output Format

```
✓ Passed: 45 tests
✗ Failed: 3 tests
📊 Coverage: 78.5%

Failures:
1. TestCreateHandler (apps/api/src/routes/tasks/post_test.go:42)
   Issue: ValidationError not returned for empty title
   Cause: Validation logic missing in TaskTitle.New()
   Fix: Add empty string check in apps/api/src/domain/model/task.go:23

2. TestListHandler (apps/api/src/routes/tasks/list_test.go:58)
   Issue: Expected 200 OK, got 500 Internal Server Error
   Cause: Database query syntax error
   Fix: Correct SQL in apps/pkgs/db/queries/tasks.sql:15
```

## Example Usage

User: "Run tests and fix any failures"

Agent:
1. Runs `make test`
2. Identifies 3 failing tests
3. Analyzes each failure
4. Suggests specific code changes
5. Optionally applies fixes with user approval
