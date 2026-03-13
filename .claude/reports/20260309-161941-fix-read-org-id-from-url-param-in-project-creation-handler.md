---

## Report: `fix: read org_id from URL param in project creation handler`

**Hash**: `3a421b9` | **Files**: `post.go`, `request.go`, `post_test.go`

---

### Summary

The handler previously read `organization_id` from the JSON request body, creating a mismatch with the RESTful route design (`/organizations/{org_id}/projects`) where the org is identified by the URL path. This fix moves `org_id` extraction to `requtil.ParseUUID(r, "org_id")` — the URL param — and removes the field from `postRequest`. The `request.go` comment now explicitly documents the intent. Tests were updated to inject `org_id` as a chi route context param and expanded to cover missing/invalid `org_id` URL param cases.

---

### Code Quality

- **Clean architecture**: Correct. Handler parses URL param and constructs value objects before passing to the repository. No domain logic leaks into the handler.
- **Value objects**: All fields (`name`, `slug`, `description`) go through domain constructors (`NewProjectName`, `NewProjectSlug`, `NewProjectDescription`). `orgID` is parsed via `requtil.ParseUUID` which returns a typed UUID.
- **Error handling**: All errors are propagated to `response.HandleError` — no errors discarded.
- **`postRequest` cleanup**: Removing `organization_id` from the body struct eliminates a silent footgun where callers could send a body `org_id` that was silently ignored (or worse, previously used instead of the URL param).

---

### Test Coverage Concerns

- **Missing test category — 特殊文字 for slug**: The happy-path special-char test uses a valid ASCII slug (`test-project`) alongside a Japanese name. There is no test for a slug containing characters that are invalid per `ProjectSlug` validation (e.g., `テスト`), which would exercise the 400 path from `NewProjectSlug`.
- **No nil/zero-value body test**: There is no test case sending an empty `{}` body or a `null` body (Null/Nil category for the request struct).
- **400 assertion style inconsistency**: The 400 sub-tests use `resp.StatusCode != http.StatusBadRequest` (weak inequality) rather than `cmp.Diff(http.StatusBadRequest, resp.StatusCode)` as required by project conventions (`testing.md`). The 201 sub-tests correctly use `cmp.Diff`.
- **`createTestOrg` uses `t.Context()`**: This is Go 1.21+ API. Confirm `go.mod` minimum version supports it (non-blocking, just worth noting).

---

### Potential Issues

1. **`409 Conflict` not tested**: If a project with the same slug already exists under the org, the repository should return a `ConflictError`. No test covers a duplicate-slug creation attempt.
2. **`createTestOrg` uses raw value objects without constructors**: `organization.OrganizationName("Test Org")` and `organization.OrganizationSlug("test-org")` are used as direct casts rather than through validated constructors. If the value object types are type aliases (not structs with constructors), this is fine — but worth confirming validation is not bypassed in test setup.

---

### Recommendations

1. **Fix 400 assertion style** — replace `resp.StatusCode != http.StatusBadRequest` with `cmp.Diff(http.StatusBadRequest, resp.StatusCode)` in the 400 and "Invalid JSON" sub-tests to match project conventions.
2. **Add duplicate-slug conflict test** — create a project, then attempt to create another with the same slug under the same org; assert `409 Conflict`.
3. **Add nil/empty body test** — send `{}` or `null` body; assert `400 Bad Request` to cover the Null/Nil TDD category.
4. **Add invalid-slug special-char test** — send a slug with non-ASCII characters or spaces; assert `400 Bad Request` to complete the 特殊文字 error-path coverage.
