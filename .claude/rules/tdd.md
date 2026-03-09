# Test-Driven Development (TDD) Rules

## TDD Cycle

**Red → Green → Refactor → Commit → Repeat**

1. 🔴 **RED**: Write failing test first
2. 🟢 **GREEN**: Write minimal code to pass
3. 🔵 **REFACTOR**: Improve code quality
4. ✅ **COMMIT**: Commit after green
5. ♻️ **REPEAT**: Next test case

## Required Test Cases

**Every function must test these 6 categories**:

1. ✅ **正常系 (Happy Path)**: Valid inputs that succeed
2. ❌ **異常系 (Error Cases)**: Invalid inputs that fail
3. 📏 **境界値 (Boundary Values)**: Min/max values, zero, negative
4. 🔤 **特殊文字 (Special Chars)**: Unicode, emoji, symbols, SQL injection
5. 📭 **空文字 (Empty String)**: Empty, whitespace-only
6. ⚠️ **Null/Nil**: Nil pointers, zero values, empty slices

## Coverage Requirements

**Mandatory minimums**:
- 📊 **Overall**: ≥ 80% package coverage
- 🎯 **Per Function**: ≥ 80% each function
- 🔍 **Critical Paths**: 100% (value objects, error handling, repositories, handlers)

**Measurement**:
```bash
# Check coverage
go test -cover ./...

# Find functions below 80%
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | awk -F'[ \t]+' '$NF+0 < 80.0 {print $0}'

# HTML view
go tool cover -html=coverage.out
```

**Coverage levels**:
```
❌ < 60%  - Critical: Block commit
⚠️ 60-79% - Warning: Add tests before commit
✅ 80-89% - Good: Meets requirement
🌟 90%+   - Excellent
```

## Best Practices

### DO ✅

- Write test before production code (always)
- Write minimal code to pass (don't over-engineer)
- Refactor after green (improve quality)
- Run all tests after refactoring
- Commit after each green phase
- Test behavior, not implementation
- Cover all 6 test categories
- Assert the **exact** expected HTTP status code (not just "not 200")
- Test new template helper functions (`lintScoreClass`, `availableTags`, etc.) in `helpers_test.go`

### DON'T ❌

- Write production code first
- Skip the failing test step
- Write multiple tests at once
- Ignore failing tests
- Test implementation details
- Skip edge cases (empty, nil, boundaries)
- Commit with < 80% coverage
- Use weak assertions like `if status != http.StatusOK` — always assert the exact expected code
- Mutate the loop variable `tt` inside table-driven test body (use local vars instead)
- Use `os.Setenv` with `t.Parallel()` — use `t.Setenv` which prevents parallel execution
- Commit `*.test` binaries or `coverage.out` profiles (build artifacts, not source)

## Table-Driven Test Pattern

```go
func TestNewTaskTitle(t *testing.T) {
    tests := []struct {
        testName string
        args     args
        expected expected
    }{
        // 正常系
        {testName: "valid input", args: args{title: "Valid"}, expected: expected{wantErr: false}},

        // 異常系
        {testName: "empty string", args: args{title: ""}, expected: expected{wantErr: true, errName: "ValidationError"}},

        // 境界値
        {testName: "max length", args: args{title: strings.Repeat("a", 100)}, expected: expected{wantErr: false}},
        {testName: "over max", args: args{title: strings.Repeat("a", 101)}, expected: expected{wantErr: true}},

        // 特殊文字
        {testName: "emoji", args: args{title: "Task 📋"}, expected: expected{wantErr: false}},
        {testName: "Japanese", args: args{title: "タスク"}, expected: expected{wantErr: false}},

        // 空文字
        {testName: "whitespace only", args: args{title: "   "}, expected: expected{wantErr: true}},

        // Nil
        {testName: "nil slice", args: args{items: nil}, expected: expected{wantErr: false}},
    }

    for _, tt := range tests {
        t.Run(tt.testName, func(t *testing.T) {
            got, err := task.NewTaskTitle(tt.args.title)

            if tt.expected.wantErr {
                if err == nil {
                    t.Fatal("expected error but got nil")
                }
                var appErr apperror.AppError
                if errors.As(err, &appErr) {
                    if diff := cmp.Diff(tt.expected.errName, appErr.ErrorName()); diff != "" {
                        t.Errorf("error name mismatch (-want +got):\n%s", diff)
                    }
                }
            } else {
                if err != nil {
                    t.Fatalf("unexpected error: %v", err)
                }
                _ = got
            }
        })
    }
}
```

## Coverage Gaps to Always Check

When writing tests for a new handler or route, verify the following are covered:

| Area | Common Gap |
|---|---|
| SSE `Stream()` handlers | session-not-found, flusher-not-supported, message-fetch-error, client-disconnect |
| List endpoints (e.g., `GetCompare`) | may exist in router but have no test file |
| Template helper functions | `lintScoreClass`, `contentLines`, `availableTags` — live in `.templ` files but are testable Go funcs |
| New templ components | `LintResultCard`, `TextDiffCard`, etc. — must be exercised by at least one partial handler test |
| `FindAll` repository methods | Must test both empty result (no rows) and populated result |

## Pre-commit Coverage Check

```bash
# .githooks/pre-commit
go test -cover ./... | tee /tmp/coverage.txt
if grep -E "coverage: [0-7][0-9]\.[0-9]%" /tmp/coverage.txt; then
    echo "❌ Coverage below 80% - commit blocked"
    exit 1
fi
```

## PR Review Checklist

- [ ] Overall coverage ≥ 80%
- [ ] No functions below 80%
- [ ] All 6 test categories covered
- [ ] All new functions have tests
- [ ] Coverage didn't decrease
- [ ] No weak status code assertions (`!= 200`) — exact codes asserted
- [ ] No loop variable mutation (`tt.field = value` inside loop body)
- [ ] No build artifacts committed (`*.test`, `coverage.out`)
- [ ] Template helper functions in new `.templ` files have corresponding `_test.go`

Remember: **Test First, Code Second, Refactor Third** 🔴🟢🔵
