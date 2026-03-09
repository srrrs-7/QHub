## Commit Analysis Report: `fix: enable BFF→API communication in dev via DEV_BYPASS_RBAC`

---

## Summary

This commit adds a development bypass for RBAC enforcement (`DEV_BYPASS_RBAC=true`) in the API middleware, solving the practical problem of the BFF (web frontend) being unable to communicate with the API locally without a JWT/Cognito identity. When the env var is set, the middleware short-circuits all role/membership checks and injects a synthetic `owner` role into the request context. Alongside this, the web client (`APIClient`) now reads its auth token from `API_AUTH_TOKEN` with a `dev-token` fallback. Both changes include new or expanded unit tests. The devcontainer override example documents the recommended local dev environment variables.

---

## Code Quality

**Clean Architecture adherence**: Good. The bypass flag is confined to the middleware layer and does not leak into domain logic. The `RequireRole` function correctly maintains its single-responsibility (auth enforcement) while the bypass is a clearly-marked operational escape hatch.

**Go idioms**: Standard `(value, error)` returns throughout. No Result monads introduced. The `do()` helper in `api.go` wraps errors with `fmt.Errorf("...: %w", err)` consistently.

**Minor concern — `map[string]string` / `map[string]any` request bodies**: Several `APIClient` methods (e.g., `CreatePrompt`, `CreateVersion`, `CreateOrganization`) accept untyped `map[string]string` or `map[string]any` instead of typed request structs. This predates this commit but remains a code quality gap — no value object validation at the client layer.

---

## Test Coverage

**Newly covered**: `DEV_BYPASS_RBAC` bypass is tested as a first-class case in `TestRequireRole`. All 6 test categories are explicitly labeled and populated in `rbac_test.go`:
- 正常系, 異常系, 境界値, 特殊文字, 空文字, Null/Nil — all present and well-structured.

**`TestRoleLevel` and `TestGetMemberRole` / `TestGetUserID`**: All pre-existing accessor functions now have standalone tests with full category coverage.

**`TestNewAPIClient_AuthHeader`**: Covers happy path, empty-token fallback, and a special-character token. Missing coverage:
- Boundary: whitespace-only `API_AUTH_TOKEN` (currently falls through to the `if token == ""` guard since it's non-empty — would silently send `Bearer   ` as the header)
- Nil/cancel context behavior on the `do()` method
- HTTP 4xx/5xx response handling in `do()` beyond the basic `>= 400` check

**`GetEmbeddingStatus`**, **`GetSemanticDiff`**, **`GetTextDiff`** and other `APIClient` methods added in previous commits have no tests.

---

## Potential Issues

1. **`DEV_BYPASS_RBAC` env var checked at request-time via `os.Getenv`**: This is intentional for flexibility, but means the bypass can be toggled without restart. In a production image where the var is accidentally set, every request would gain `owner` privileges silently. A startup-time warning log when this flag is active would provide a clear signal in logs.

2. **Whitespace `API_AUTH_TOKEN`**: `if token == ""` does not handle `"   "` — a whitespace-only env value would be forwarded as the Bearer token verbatim. Low-risk in practice but a silent misconfiguration vector.

3. **`DEV_BYPASS_RBAC` injects `RoleOwner` but skips setting `userIDKey`**: Downstream handlers calling `GetUserID(ctx)` will receive `uuid.UUID{}, false`. If any member-scoped handler relies on a valid user ID (e.g., audit logging, filtering by current user), it will receive the zero UUID silently. This is a latent bug that could produce incorrect data in dev.

4. **`DeleteTag` constructs URL with unsanitized query param** (`"/api/v1/tags?name=" + name`): predates this commit, but if `name` contains `&`, `=`, or `#`, the query string will be malformed. Should use `url.Values`.

5. **Generic `fmt.Errorf("API error %d", resp.StatusCode)`** in `do()` discards the response body. The API returns structured JSON errors, so callers cannot distinguish 401 from 403 from 404 programmatically without parsing the status code themselves — brittle for the BFF layer.

---

## Recommendations

1. **(High)** Add a startup-time `logger.Warn("DEV_BYPASS_RBAC is enabled — all RBAC checks are skipped")` log in `RequireRole` (or `cmd/main.go`) so the bypass is visible in container logs and impossible to overlook in staging/prod incidents.

2. **(High)** When `DEV_BYPASS_RBAC` bypasses role checks, also inject a fixed synthetic `userID` into context (e.g., a well-known dev UUID constant) so `GetUserID` returns a usable value and downstream code behaves correctly end-to-end in dev.

3. **(Medium)** Add a `TestNewAPIClient_AuthHeader` case for whitespace-only token, and consider trimming `token` before the empty check in `NewAPIClient`.

4. **(Medium)** Replace `"/api/v1/tags?name=" + name` with `url.Values{"name": {name}}.Encode()` to prevent malformed URLs on special-character tag names.

5. **(Low)** Introduce a typed `APIError` struct in the web `client` package that captures both the HTTP status code and parsed body, allowing the BFF to surface domain error messages (e.g., "not a member of this organization") rather than generic `"API error 403"` strings.
