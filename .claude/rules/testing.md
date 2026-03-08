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

Integration tests use real DB with transaction isolation:

```go
func TestGetHandler(t *testing.T) {
    q := testutil.SetupTestTx(t)  // transaction-isolated DB
    tt.setup(t, q)                // seed test data

    repo := task_repository.NewTaskRepository(q)
    handler := NewTaskHandler(repo).Get()

    req := httptest.NewRequest(http.MethodGet, "/tasks/"+id, nil)
    testutil.SetAuthHeader(req)   // set Bearer token
    w := httptest.NewRecorder()

    handler.ServeHTTP(w, req)

    if diff := cmp.Diff(http.StatusOK, w.Result().StatusCode); diff != "" {
        t.Errorf("status code mismatch (-want +got):\n%s", diff)
    }
}
```

For chi URL params in tests:
```go
rctx := chi.NewRouteContext()
rctx.URLParams.Add("id", taskID)
req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
```

## Database Testing

Use `testutil.SetupTestTx(t)` for transaction-wrapped tests (auto-rollback on cleanup):

```go
q := testutil.SetupTestTx(t)
repo := NewTaskRepository(q)

// Test repository methods with real database
got, err := repo.FindByID(context.Background(), id)
```

## Test Organization

- Unit tests: Test value objects and domain logic
- Integration tests: Test repositories and handlers with real DB (requires `docker compose up -d db`)
- Table-driven tests: Group related test cases together
- Test naming: Use descriptive names that explain the scenario

## Running Tests

```bash
# All tests (requires DB)
make test

# Single package
cd apps/api && go test ./src/routes/tasks/

# Single test
cd apps/api && go test -run TestListHandler ./src/routes/tasks/

# With coverage
cd apps/api && go test -cover ./...
```

## Test Database

`make test` automatically runs `make atlas-apply` before test execution.
