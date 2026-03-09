Here is the insights report for commit `3ff347c`:

---

## Commit Insights: `fix: remove stale embSvc argument from PromptHandler test calls`

### Summary

This commit is a pure test housekeeping fix following the prior refactor that removed the embedding/RAG stack (`43dc8f3`). The `embSvc` argument was removed from `NewPromptHandler(...)` in the production code, but 11 test files still passed it as the second argument. All 36 changes are identical mechanical substitutions — e.g., `NewPromptHandler(pRepo, vRepo, embSvc, diffSvc, nil)` → `NewPromptHandler(pRepo, vRepo, diffSvc, nil)` — restoring compile-time correctness across the entire `prompts` test suite. No logic changed.

---

### Code Quality

- **Architecture**: Tests correctly depend only on domain interfaces (`prompt.PromptRepository`, `prompt.VersionRepository`) and service constructors. Dependency direction is clean.
- **Value objects**: Proper use throughout — `prompt.PromptIDFromUUID`, `prompt.PromptName`, `prompt.PromptSlug`, `prompt.VersionCmd`, etc. No raw primitive strings bypassing domain types in production-path code.
- **No Result monad concern**: Project uses standard Go `(value, error)` — correct per convention.

---

### Test Coverage

All 6 required test categories are well-represented across the suite:

| Category | Evidence |
|---|---|
| 正常系 (happy path) | All handlers have 200/201 success cases |
| 異常系 (error) | 404 for non-existent resources, 400 for invalid inputs |
| 境界値 (boundary) | `version: "0"`, `slug: "a"` (too short), min-length name/slug |
| 特殊文字 (special chars) | Japanese names, emoji in descriptions, Japanese prompt content |
| 空文字 (empty) | Empty slug, empty prompt_id, empty status field |
| Null/Nil | Missing optional fields (variables, change_description), empty body |

`TestPutVersionStatusHandler` stands out with particularly thorough state-machine coverage: all valid transitions, all invalid transitions, and the production-replacement side effect (v1 auto-archives when v2 goes to production).

---

### Potential Issues

**Minor: weak assertion in `get_diff_test.go` 400 block**
```go
// get_diff_test.go:163
if resp.StatusCode == http.StatusOK {
    t.Errorf("expected error status, got %d", resp.StatusCode)
}
```
The non-numeric version test cases (`v1: "abc"`, `v2: "xyz"`) are noted as returning 500 (not 400) due to `strconv.Atoi` errors not being `AppError`. The assertion accepts any non-200 status, which silently permits 500s on what should arguably be client errors. The same loose assertion pattern appears in `get_text_diff_test.go:204` and `get_version_test.go:264`.

**Minor: `put_test.go` mutates shared test case struct at runtime**
```go
// put_test.go:183-187
if tt.testName != "invalid project_id" {
    projID := createTestProject(t, q)
    _ = createTestPromptRecord(t, q, projID)
    tt.projectID = projID  // mutates loop variable
}
```
Mutating the loop variable `tt` is harmless here since tests don't run in parallel, but it's fragile style — if `t.Parallel()` is ever added, this will race.

---

### Recommendations

1. **Fix `strconv.Atoi` error mapping** — wrap integer parse failures in `BadRequestError` so non-numeric version params return 400, not 500. This is a real UX issue, not just a test concern.
2. **Tighten 400 assertions** — replace `!= http.StatusOK` guards with explicit `!= http.StatusBadRequest` checks in the diff/text-diff/version 400 test groups, so accidental 500s are caught.
3. **Extract loop variable before mutation** — in `put_test.go` 400 block, use a local copy (`pid := createTestProject(...)` then assign to a local `projectID`) rather than mutating `tt.projectID` in place.
4. **Add `GetCompare` (Welch's t-test) to the test suite** — the `GET /prompts/{id}/versions/{v1}/{v2}/compare` endpoint exists in the router but no test file appears in this package. Verify it has coverage elsewhere or add it.
