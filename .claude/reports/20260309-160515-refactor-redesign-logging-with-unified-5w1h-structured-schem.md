Here is the report for commit `340b8e7`:

---

# Insights Report: `refactor: redesign logging with unified 5W1H structured schema`

## Summary

This commit replaces ad-hoc slog calls scattered across middleware with a cohesive 5W1H (Who-What-When-Where-Why-How) structured logging schema. A new `RequestLog` pointer-in-context pattern (`NewRequestLog`/`RequestLogFrom` in `apps/pkgs/logger`) allows each middleware layer (BearerAuth, ApiKeyAuth, RequireRole) to write identity fields (UserID, OrgID, AuthMethod) that the HTTP Logger middleware reads after `next.ServeHTTP()` returns — achieving a single, rich log entry per request without channels or mutexes. `cmd/main.go` startup logs were also migrated to the new grouped schema.

---

## Code Quality

**Architecture compliance:** Clean. The `RequestLog` type lives in `utils/logger` (shared pkgs), middleware reads/writes it, and no domain layer is touched. Dependency direction is respected.

**Value objects:** N/A for this change — logging is infrastructure, not domain.

**Mutable shared struct (intentional design note):** `RequestLog` is a plain mutable struct stored as a pointer in context. The code is correct for the single-goroutine request path, but the comment in logger.go says "without any mutex, because each request has its own pointer." This assumption holds only if no goroutine spawned during the request writes to `rl` concurrently. The background goroutine in `ApiKeyAuth` (updating `last_used_at`) does not write to `rl`, so it is safe today — but this is a fragile invariant worth documenting more explicitly.

**`DEV_BYPASS_RBAC` partial compliance:** The bypass injects `RoleOwner` into context ✅ and sets `rl.AuthMethod = "bypass"` ✅, but does **not** inject a synthetic `userID` into context (the CLAUDE.md/architecture.md rule requires both). `GetUserID(ctx)` will return `(zero UUID, false)` for bypassed requests, which silently breaks any downstream handler that calls `GetUserID`.

---

## Test Coverage Concerns

Coverage is strong overall. All 6 test categories are present and correctly labeled across `logger_test.go`, `logger_test.go` (middleware), and `rbac_test.go`.

**Gap — `ApiKeyAuth` has no test file.** The middleware was extended with `RequestLog` writes and structured error logging, but there is no `apikey_test.go`. The following paths are untested:
- Missing `X-API-Key` header → 401
- Invalid key hash (DB miss) → 401
- Revoked key → 401
- Expired key → 401
- `GetApiKeyOrgID` context helper
- `rl.OrgID` field population

**Gap — `BearerAuth` has no test file.** The `RequestLog` AuthMethod annotation was added but no `bearer_auth_test.go` exists. Untested: valid token, invalid token, empty header, `rl.AuthMethod = "bearer"` propagation.

**`TestLogger_RequestLogPropagation`** (logger_test.go:204): The test verifies the request completes with 200 but never actually asserts that `rl.UserID/OrgID/AuthMethod` values were read by the Logger. The `capturedUserID/OrgID/Auth` variables are declared but immediately assigned `_ =`. The test documents the design intent but provides no behavioral assertion about propagation.

---

## Potential Issues

**1. `DEV_BYPASS_RBAC` missing synthetic userID injection** (security/correctness)
Per `architecture.md`, the bypass must inject both role and userID. Only role is injected (`rbac.go:71-72`). Any handler calling `GetUserID(ctx)` under bypass will receive a zero UUID, potentially writing zero-UUID records to the database or returning unexpected 401s from handlers that require a valid user.

**2. Raw query string logged without redaction** (privacy/security)
`logger.go:78,113` logs `r.URL.RawQuery` verbatim in `how.query`. If any endpoint ever accepts sensitive values in query parameters (tokens, API keys, PII), they will appear in structured logs. Consider a redact-list or truncation for production.

**3. `r.RemoteAddr` used for IP logging** (accuracy)
`logger.go:86` logs `r.RemoteAddr` as `who.ip`. The middleware chain applies `RealIP` before `Logger`, which rewrites `r.RemoteAddr` from `X-Real-IP`/`X-Forwarded-For` — so this is actually correct as long as middleware order is preserved. However, it is not obvious from reading logger.go alone and is order-dependent.

**4. `json.NewEncoder(w).Encode` errors silently discarded in `writeRBACError`**
`rbac.go:173` uses `//nolint:errcheck`. While the pattern is pragmatically fine (the response is already committed), it is inconsistent with the project's explicit convention of never silently discarding `json` errors. A `logger.Warn(...)` on encode failure would align with the new logging schema.

---

## Recommendations

1. **Add `apikey_test.go` and `bearer_auth_test.go`** — both middlewares were modified but have no tests. All 6 test categories are required by project convention; coverage is currently 0% for these files.

2. **Inject synthetic `userID` in `DEV_BYPASS_RBAC` path** (`rbac.go:71-74`) — per `architecture.md` requirement. Add `ctx = context.WithValue(ctx, userIDKey, devUserID)` with a fixed well-known UUID constant to prevent zero-UUID records in dev databases.

3. **Strengthen `TestLogger_RequestLogPropagation`** — the test currently asserts nothing about the propagated fields. Either capture slog output (e.g., via a custom `slog.Handler`) and assert field values, or remove the dead `captured*` variables and add a comment explaining what is actually being verified.

4. **Redact or truncate `how.query` in production logs** — add a `sanitizeQuery(raw string) string` helper that strips known sensitive parameter names (`token`, `key`, `secret`, `password`) before logging.

5. **Document the "no concurrent writes to RequestLog" invariant** — add a comment in `RequestLog` explicitly listing which middleware layers write to it and asserting that background goroutines must not write. This makes the lock-free design safe to maintain as the codebase grows.
