# 詳細メトリクス分析

## 📊 Coverage Heatmap

```
█████████████████████████████████████████████████████ 99.3% utils/parallel ✅
█████████████████████████████████████████████████      91.0% routes/tasks ✅
█████████████████████                                  42.4% routes/response ⚠️
█████                                                  11.1% domain/task ❌
                                                        0.0% types/result 🔥
                                                        0.0% apperror 🔥
                                                        0.0% task_repository 🔥
                                                        0.0% env ❌
                                                        0.0% logger ❌
                                                        0.0% testutil ❌
                                                        0.0% db ❌
                                                        0.0% web/* ❌

Legend: ✅ Pass (≥80%) | ⚠️ Warning (60-79%) | ❌ Fail (40-59%) | 🔥 Critical (<40%)
```

---

## 🎯 Function-Level Analysis

### types/result.go (0% - 16 functions untested)

| Function | Lines | Complexity | Priority | Test Cases Needed |
|----------|-------|------------|----------|-------------------|
| `Ok()` | 2 | Low | 🔥 CRITICAL | 6 |
| `Err()` | 2 | Low | 🔥 CRITICAL | 6 |
| `FromPair()` | 5 | Medium | 🔥 CRITICAL | 12 |
| `IsOk()` | 2 | Low | High | 4 |
| `IsErr()` | 2 | Low | High | 4 |
| `Map()` | 5 | Medium | 🔥 CRITICAL | 12 |
| `MapErr()` | 5 | Medium | High | 12 |
| `FlatMap()` | 5 | Medium | 🔥 CRITICAL | 12 |
| `AndThen()` | 2 | Low | High | 6 |
| `Match()` | 6 | Medium | 🔥 CRITICAL | 10 |
| `Combine()` | 9 | High | High | 15 |
| `Pipe2()` | 2 | Medium | 🔥 CRITICAL | 10 |
| `Pipe3()` | 2 | Medium | 🔥 CRITICAL | 10 |
| `Pipe4()` | 2 | Medium | High | 8 |
| `Pipe5()` | 2 | Medium | High | 8 |
| `UnwrapOr()` | 5 | Medium | High | 8 |

**Total**: 135+ test cases needed

---

### apperror/apperror.go (0% - 12 functions untested)

| Function | Lines | Complexity | Priority | Test Cases Needed |
|----------|-------|------------|----------|-------------------|
| `NewNotFoundError()` | 7 | Low | 🔥 CRITICAL | 6 |
| `NewValidationError()` | 7 | Low | 🔥 CRITICAL | 6 |
| `NewDatabaseError()` | 7 | Low | 🔥 CRITICAL | 6 |
| `NewUnauthorizedError()` | 7 | Low | High | 6 |
| `NewInternalServerError()` | 7 | Low | High | 6 |
| `NewBadRequestError()` | 7 | Low | High | 6 |
| `NewConflictError()` | 7 | Low | Medium | 6 |
| `NewForbiddenError()` | 7 | Low | Medium | 6 |
| `ErrorName()` | 2 | Low | High | 4 |
| `DomainName()` | 2 | Low | High | 4 |
| `Error()` | 5 | Medium | 🔥 CRITICAL | 8 |
| `Unwrap()` | 2 | Low | High | 4 |

**Total**: 68+ test cases needed

---

### domain/task/task.go (0% - 8 functions untested)

| Function | Lines | Complexity | Priority | Test Cases Needed |
|----------|-------|------------|----------|-------------------|
| `NewTaskID()` | 2 | Low | 🔥 CRITICAL | 8 |
| `TaskID.String()` | 2 | Low | High | 4 |
| `TaskTitle.String()` | 2 | Low | Medium | 4 |
| `TaskDescription.String()` | 2 | Low | Medium | 4 |
| `TaskStatus.String()` | 2 | Low | Medium | 4 |
| `NewTask()` | 7 | Low | High | 6 |
| `IsCompleted()` | 2 | Low | High | 6 |
| `IsPending()` | 2 | Low | High | 6 |

**Total**: 42+ test cases needed

**Critical Issue**: `NewTaskID()` uses `uuid.MustParse()` which **panics on invalid input**. No recovery mechanism tested.

---

### task_repository (0% - 4 functions untested)

| Function | Lines | Complexity | Priority | Test Cases Needed |
|----------|-------|------------|----------|-------------------|
| `FindTaskByID()` | 26 | High | 🔥 CRITICAL | 12 |
| `FindAllTasks()` | 27 | High | 🔥 CRITICAL | 10 |
| `CreateTask()` | ~30 | High | High | 12 |
| `UpdateTask()` | ~40 | High | High | 12 |

**Total**: 46+ test cases needed

**Critical Paths**:
- SQL error handling (3 types: `sql.ErrNoRows`, `context.DeadlineExceeded`, generic)
- Data transformation (DB model → Domain model)
- Transaction management

---

## 💰 Cost/Effort Estimation

### Development Time (Engineer-Days)

| Phase | Package | Functions | Test Cases | Days | Priority |
|-------|---------|-----------|------------|------|----------|
| **1** | types/result | 16 | 135+ | 3-4 | 🔥 CRITICAL |
| **1** | apperror | 12 | 68+ | 2-3 | 🔥 CRITICAL |
| **2** | domain/task | 8 | 42+ | 2 | 🔥 CRITICAL |
| **2** | task_repository | 4 | 46+ | 3-4 | 🔥 CRITICAL |
| **3** | routes/response | 6 | 30+ | 2 | High |
| **3** | env/logger/testutil | 10+ | 35+ | 2 | Medium |
| **4** | web/* | 15+ | 50+ | 4-5 | Medium |
| - | **Refactoring** | - | - | 2-3 | - |
| - | **Code Review** | - | - | 2 | - |

**Total Estimated**: 22-30 engineer-days (4-6 weeks with 1 engineer)

### Risk-Adjusted Timeline

- **Optimistic**: 3 weeks (1 engineer, no blockers)
- **Realistic**: 5 weeks (1 engineer, normal blockers)
- **Pessimistic**: 8 weeks (1 engineer, significant rework)

**Recommendation**: 2 engineers for 3 weeks (Phase 1-2 parallel)

---

## 🔍 Specific Test Examples

### Example 1: Result.Map() Test Suite

```go
func TestMap(t *testing.T) {
    type args struct {
        result Result[int, error]
        fn     func(int) string
    }
    type expected struct {
        isOk  bool
        value string
        err   error
    }

    tests := []struct {
        testName string
        args     args
        expected expected
    }{
        // ✅ 正常系
        {
            testName: "map Ok value",
            args: args{
                result: Ok[int, error](42),
                fn:     func(i int) string { return strconv.Itoa(i) },
            },
            expected: expected{isOk: true, value: "42"},
        },
        {
            testName: "map Ok value with complex transformation",
            args: args{
                result: Ok[int, error](100),
                fn:     func(i int) string { return fmt.Sprintf("Value: %d", i) },
            },
            expected: expected{isOk: true, value: "Value: 100"},
        },

        // ❌ 異常系
        {
            testName: "map Err preserves error",
            args: args{
                result: Err[int, error](errors.New("error")),
                fn:     func(i int) string { return strconv.Itoa(i) },
            },
            expected: expected{isOk: false, err: errors.New("error")},
        },

        // 📏 境界値
        {
            testName: "map with zero value",
            args: args{
                result: Ok[int, error](0),
                fn:     func(i int) string { return strconv.Itoa(i) },
            },
            expected: expected{isOk: true, value: "0"},
        },
        {
            testName: "map with negative value",
            args: args{
                result: Ok[int, error](-42),
                fn:     func(i int) string { return strconv.Itoa(i) },
            },
            expected: expected{isOk: true, value: "-42"},
        },
        {
            testName: "map with max int",
            args: args{
                result: Ok[int, error](math.MaxInt64),
                fn:     func(i int) string { return strconv.Itoa(i) },
            },
            expected: expected{isOk: true, value: strconv.Itoa(math.MaxInt64)},
        },

        // 🔤 特殊文字
        {
            testName: "map to string with emoji",
            args: args{
                result: Ok[int, error](1),
                fn:     func(i int) string { return fmt.Sprintf("🎉 %d", i) },
            },
            expected: expected{isOk: true, value: "🎉 1"},
        },
        {
            testName: "map to string with Japanese",
            args: args{
                result: Ok[int, error](42),
                fn:     func(i int) string { return fmt.Sprintf("値: %d", i) },
            },
            expected: expected{isOk: true, value: "値: 42"},
        },

        // 📭 空文字
        {
            testName: "map to empty string",
            args: args{
                result: Ok[int, error](0),
                fn:     func(i int) string { return "" },
            },
            expected: expected{isOk: true, value: ""},
        },

        // ⚠️ Nil
        {
            testName: "map with nil error in Err",
            args: args{
                result: Err[int, error](nil),
                fn:     func(i int) string { return strconv.Itoa(i) },
            },
            expected: expected{isOk: false, err: nil},
        },
    }

    for _, tt := range tests {
        t.Run(tt.testName, func(t *testing.T) {
            result := Map(tt.args.result, tt.args.fn)

            if tt.expected.isOk {
                assert.True(t, result.IsOk())
                assert.Equal(t, tt.expected.value, *result.value)
            } else {
                assert.True(t, result.IsErr())
                if tt.expected.err != nil {
                    assert.EqualError(t, *result.err, tt.expected.err.Error())
                }
            }
        })
    }
}
```

**Coverage achieved**: ~90% for `Map()` function

---

### Example 2: NewTaskID() Panic Test

```go
func TestNewTaskID(t *testing.T) {
    type args struct {
        id string
    }
    type expected struct {
        shouldPanic bool
        value       string
    }

    tests := []struct {
        testName string
        args     args
        expected expected
    }{
        // ✅ 正常系
        {
            testName: "valid UUID",
            args:     args{id: "123e4567-e89b-12d3-a456-426614174000"},
            expected: expected{shouldPanic: false, value: "123e4567-e89b-12d3-a456-426614174000"},
        },
        {
            testName: "valid UUID uppercase",
            args:     args{id: "123E4567-E89B-12D3-A456-426614174000"},
            expected: expected{shouldPanic: false, value: "123e4567-e89b-12d3-a456-426614174000"},
        },

        // ❌ 異常系 - CRITICAL: これらはPANICする！
        {
            testName: "invalid UUID - too short",
            args:     args{id: "123"},
            expected: expected{shouldPanic: true},
        },
        {
            testName: "invalid UUID - malformed",
            args:     args{id: "not-a-uuid"},
            expected: expected{shouldPanic: true},
        },
        {
            testName: "invalid UUID - wrong format",
            args:     args{id: "123e4567e89b12d3a456426614174000"},
            expected: expected{shouldPanic: true},
        },

        // 📭 空文字 - CRITICAL: PANIC
        {
            testName: "empty string",
            args:     args{id: ""},
            expected: expected{shouldPanic: true},
        },

        // 🔤 特殊文字
        {
            testName: "SQL injection attempt",
            args:     args{id: "'; DROP TABLE tasks; --"},
            expected: expected{shouldPanic: true},
        },
    }

    for _, tt := range tests {
        t.Run(tt.testName, func(t *testing.T) {
            if tt.expected.shouldPanic {
                assert.Panics(t, func() {
                    _ = NewTaskID(tt.args.id)
                })
            } else {
                assert.NotPanics(t, func() {
                    result := NewTaskID(tt.args.id)
                    assert.Equal(t, tt.expected.value, result.String())
                })
            }
        })
    }
}
```

**Critical Finding**: `NewTaskID()` has **no error handling** - panics on invalid input!

**Recommendation**: Change signature to:
```go
func NewTaskID(id string) Result[TaskID, AppError]
```

---

### Example 3: Error.Error() Format Test

```go
func TestBaseErr_Error(t *testing.T) {
    type args struct {
        errName    string
        domainName string
        underlying error
    }
    type expected struct {
        message string
    }

    tests := []struct {
        testName string
        args     args
        expected expected
    }{
        // ✅ 正常系
        {
            testName: "error with underlying",
            args: args{
                errName:    "ValidationError",
                domainName: "Task",
                underlying: errors.New("title is required"),
            },
            expected: expected{
                message: "ValidationError [Task]: title is required",
            },
        },
        {
            testName: "error without underlying",
            args: args{
                errName:    "NotFoundError",
                domainName: "Task",
                underlying: nil,
            },
            expected: expected{
                message: "NotFoundError [Task]",
            },
        },

        // 🔤 特殊文字
        {
            testName: "error with emoji in domain",
            args: args{
                errName:    "ValidationError",
                domainName: "Task📋",
                underlying: errors.New("test"),
            },
            expected: expected{
                message: "ValidationError [Task📋]: test",
            },
        },
        {
            testName: "error with Japanese in message",
            args: args{
                errName:    "ValidationError",
                domainName: "タスク",
                underlying: errors.New("タイトルが必要です"),
            },
            expected: expected{
                message: "ValidationError [タスク]: タイトルが必要です",
            },
        },

        // 📭 空文字
        {
            testName: "empty domain name",
            args: args{
                errName:    "ValidationError",
                domainName: "",
                underlying: errors.New("test"),
            },
            expected: expected{
                message: "ValidationError []: test",
            },
        },

        // 境界値
        {
            testName: "very long error message",
            args: args{
                errName:    "DatabaseError",
                domainName: "Repository",
                underlying: errors.New(strings.Repeat("a", 1000)),
            },
            expected: expected{
                message: fmt.Sprintf("DatabaseError [Repository]: %s", strings.Repeat("a", 1000)),
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.testName, func(t *testing.T) {
            err := baseErr{
                errName:    tt.args.errName,
                domainName: tt.args.domainName,
                err:        tt.args.underlying,
            }

            assert.Equal(t, tt.expected.message, err.Error())
        })
    }
}
```

---

## 🎓 TDD Workflow Example

### Implementing `Result.Map()` with TDD

#### Step 1: 🔴 RED - Write Failing Test

```go
func TestMap_Basic(t *testing.T) {
    result := types.Ok[int, error](42)
    mapped := types.Map(result, func(i int) string { return strconv.Itoa(i) })
    
    assert.True(t, mapped.IsOk())
    assert.Equal(t, "42", mapped.UnwrapOr(""))
}
```

**Result**: ❌ Test fails (function doesn't exist yet)

#### Step 2: 🟢 GREEN - Write Minimal Code

```go
func Map[T, U, E any](r Result[T, E], fn func(T) U) Result[U, E] {
    if r.err != nil {
        return Err[U](*r.err)
    }
    return Ok[U, E](fn(*r.value))
}
```

**Result**: ✅ Test passes

#### Step 3: 🔵 REFACTOR - Add More Tests

Add tests for: Err case, edge cases, special characters, etc.

#### Step 4: ✅ COMMIT

```bash
git add apps/pkgs/types/result_test.go
git commit -m "test: add Map() tests with 6 categories"
```

**Coverage**: 90%+ ✅

#### Step 5: ♻️ REPEAT

Move to next function (`MapErr`, `FlatMap`, etc.)

---

## 📊 Progress Tracking Template

```markdown
# Test Coverage Progress

## Week 1: Phase 1 - Critical Core

- [ ] types/result.go (0% → 90%)
  - [x] Ok, Err, FromPair (Day 1)
  - [x] IsOk, IsErr, UnwrapOr (Day 1)
  - [ ] Map, MapErr, FlatMap (Day 2)
  - [ ] Match, Combine (Day 2)
  - [ ] Pipe2, Pipe3, Pipe4, Pipe5 (Day 3)
  - [ ] Edge cases & refactoring (Day 4)

- [ ] apperror/apperror.go (0% → 90%)
  - [ ] baseErr methods (Day 1)
  - [ ] All error constructors (Day 2)
  - [ ] Error chain & format tests (Day 3)

**Coverage**: 0% → ~60% (Week 1 end)

## Week 2-3: Phase 2 - Domain & Repository

- [ ] domain/task/task.go (11% → 85%)
  - [ ] Value Objects (Day 1)
  - [ ] NewTask, IsCompleted, IsPending (Day 2)
  - [ ] Fix NewTaskID panic issue (Day 2)

- [ ] task_repository (0% → 85%)
  - [ ] FindTaskByID with DB mocking (Day 3-4)
  - [ ] FindAllTasks (Day 4)
  - [ ] CreateTask, UpdateTask (Day 5)

**Coverage**: 60% → ~75% (Week 3 end)

## Week 4: Phase 3 - Utilities

- [ ] routes/response (42% → 80%)
- [ ] env, logger, testutil (0% → 80%)

**Coverage**: 75% → ~82% (Week 4 end) ✅ TARGET ACHIEVED
```

---

## 🚀 Quick Win Opportunities

### Easiest to Fix (Low Effort, High Impact)

1. **utils/env** (3 functions, ~1 hour)
   - Simple string/int/bool conversion tests
   - No dependencies

2. **baseErr methods** (4 methods, ~2 hours)
   - ErrorName, DomainName, Error, Unwrap
   - Pure functions, easy to test

3. **Task Value Objects** (5 String() methods, ~1 hour)
   - Trivial conversions
   - Quick confidence boost

**Total Quick Wins**: 4 hours → +15% coverage

---

