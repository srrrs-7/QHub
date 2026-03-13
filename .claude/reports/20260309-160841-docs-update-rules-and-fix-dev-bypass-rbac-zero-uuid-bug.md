The sandbox is blocking writes to `.claude/reports/`. Here is the full report:

---

# Commit Analysis: docs: update rules and fix DEV_BYPASS_RBAC zero-UUID bug

**Hash**: efbb7ec | **Date**: 2026-03-09

---

## Summary

This commit closes a real bug where `DEV_BYPASS_RBAC=true` skipped JWT/Cognito auth but failed to inject a synthetic `userID` into the request context. Downstream handlers calling `GetUserID(ctx)` received the zero UUID, causing silent data corruption in DB writes. The fix adds `devBypassUserID = uuid.MustParse("00000000-0000-0000-0000-000000000001")` and injects it alongside `RoleOwner` in `RequireRole`. A startup `Warn` log in `cmd/main.go` makes the bypass visible in container logs. The commit also codifies lessons from the past eight sessions into five `.claude/rules/` files and adds 8 prior-commit analysis reports.

---

## Code Quality

| Aspect | Finding |
|---|---|
| Clean architecture | `rbac.go` depends only on `db.Querier` + `logger` — no domain imports, no circular deps. ✅ |
| Type safety | Synthetic userID is `uuid.UUID` (typed); `GetUserID` returns `(uuid.UUID, bool)` — no silent zero-value fallback. ✅ |
| Error handling | `writeRBACError` discards `json.Encode` error with `//nolint:errcheck`. Legitimate (headers committed), but project style prefers `_ = json.NewEncoder(w).Encode(...)` with a comment over a lint suppression. Minor. |
| Context key safety | `memberRoleKey = 20`, `userIDKey = 21` as typed `contextKey` integers. Safe if `type contextKey int` is a distinct named type — verify no collision with other middleware keys. |

---

## Test Coverage

`rbac_test.go` covers all 6 mandatory TDD categories with 18 table-driven cases. All assertions use `cmp.Diff` — no weak `!= 200` patterns. A dedicated `TestDevBypassInjectsSyntheticUserID` test specifically guards the bugfix regression. `TestRoleLevel`, `TestGetMemberRole`, and `TestGetUserID` cover all helpers.

**One gap**: `GetApiKeyOrgID` is called in `RequireRole` but not tested in this file — confirm coverage exists in `combined_auth_test.go` or equivalent.

---

## Potential Issues

1. **`os.Getenv` called per request** — `DEV_BYPASS_RBAC` is read on every HTTP request in `RequireRole`. Should be read once at startup and closed over as a `bool`.

2. **`devBypassUserID` shares zero-UUID prefix** — `00000000-…-0001` shares the `00000000` prefix with `uuid.Nil`. A UUID in a clearly non-zero range (e.g., `ffffffff-…-0001`) would be harder to confuse with genuine zero-UUID bugs in DB dumps/logs.

3. **No startup log when bypass is OFF** — `rbac.bypass_enabled` is only logged when `true`. Production starts don't confirm RBAC is active; an `Info` log for the disabled case would aid security audits.

4. **`//nolint:errcheck` deviates from project convention** — should be `_ = json.NewEncoder(w).Encode(...)` with an inline comment.

---

## Recommendations (prioritized)

1. Read `DEV_BYPASS_RBAC` once at startup; close over the `bool` in `RequireRole` to eliminate per-request `os.Getenv` calls.
2. Verify `GetApiKeyOrgID` test coverage in an existing test file.
3. Add `Info` startup log confirming RBAC is active when bypass is disabled.
4. Replace `//nolint:errcheck` with explicit `_ = ...` discard + comment.
5. Use a visually distinctive dev bypass UUID outside the zero-prefix range.
