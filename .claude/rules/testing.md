# Testing Conventions

## Table-Driven Tests

All tests must use table-driven pattern:

```go
tests := []struct {
    testName string
    args     args
    expected expected
}{
    {
        testName: "descriptive name of test case",
        args:     args{ /* test inputs */ },
        expected: expected{ /* expected outputs */ },
    },
}

for _, tt := range tests {
    t.Run(tt.testName, func(t *testing.T) {
        // Test implementation
    })
}
```

## HTTP Handler Testing

Use `testutil` package helpers:

```go
import "pkgs/testutil"

req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks", nil)
w := httptest.NewRecorder()
handler(w, req)

// Use testutil for assertions
testutil.AssertStatusCode(t, w, http.StatusOK)
testutil.AssertJSONResponse(t, w, expectedResponse)
```

## Database Testing

Use `testutil.SetupTestDB` for integration tests:

```go
db := testutil.SetupTestDB(t)
defer db.Close()

// Test repository methods with real database
```

## Test Organization

- Unit tests: Test individual functions with mocked dependencies
- Integration tests: Test with real database (requires `docker compose up -d db`)
- Table-driven tests: Group related test cases together
- Test naming: Use descriptive names that explain the scenario and expected behavior

## Running Tests

```bash
# All tests (requires DB)
make test

# Single package
cd apps/api && go test ./src/routes/tasks/

# Single test
cd apps/api && go test -run TestListHandler ./src/routes/tasks/

# With coverage
go test -cover ./...
```

## Test Database Migrations

Tests automatically run migrations using `make atlas-apply` before test execution.
