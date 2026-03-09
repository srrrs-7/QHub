---

## Commit Insights Report — `1052dec`

**docs: add comprehensive SDK documentation for Go, Python, and TypeScript**

---

## Summary

This commit adds 1,081 lines of documentation across four README files: a root-level `README.md` update and new comprehensive READMEs for the Go, Python, and TypeScript SDKs. The root README gains an SDK comparison table and per-language quickstart snippets. Each SDK README covers installation, authentication (Bearer + API key), all available resource methods, advanced usage (diff, lint, search, analytics, consulting/RAG), error handling, custom HTTP client/fetch configuration, project structure, and testing patterns. No production code was changed.

---

## Code Quality

Documentation-only commit — clean architecture, value objects, and Result patterns are not applicable here. The code snippets embedded in the docs are consistent with the project's Go idioms (standard `(value, error)` returns, context propagation, functional options via `sdk.WithBaseURL`/`sdk.WithHTTPClient`).

**Minor inconsistency noted:**

- Go SDK README (`apps/sdk/README.md:7`) states `Go 1.25+` as the requirement, but `CLAUDE.md` and project tooling reference **Go 1.26**. This will cause user confusion.
- Python README (`apps/sdk-python/README.md:10`) claims "Zero external dependencies" in the Features list, yet the Requirements section correctly lists `httpx` and `pydantic` as dependencies. The feature bullet is misleading.

---

## Test Coverage Concerns

This commit contains no new functions or production code — no test coverage impact. However:

- The Python and TypeScript READMEs both claim "Fully tested" in the Features section. If the referenced test suites (`pytest`, `vitest`) don't yet exist or have gaps, these claims are unverifiable from this commit alone.
- The Go SDK README's testing example (`apps/sdk/README.md:360–374`) shows an `httptest.NewServer` pattern, which is the correct approach, but doesn't demonstrate the 6 required test categories mandated by TDD rules.

---

## Potential Issues

1. **API key misuse pattern** — All three SDKs document using an API key as the `bearer_token` constructor argument (e.g., Go: `sdk.NewClient(apiKey.Key)`). The constructor is named `NewClient(bearerToken, ...)`, implying Bearer semantics. Callers may not realize API keys are sent as a different auth header. If the SDK implementation sends API keys as `Authorization: Bearer <key>` rather than `X-API-Key: <key>` (or the server's expected format), this silently breaks log ingestion. Worth verifying the actual SDK implementation handles dual-auth correctly.

2. **Hardcoded localhost default** — All three SDKs default `base_url`/`baseUrl` to `http://localhost:8080`. This is undocumented as a default in the Go README's `Available Methods` section, but is mentioned only in the Quick Start. Users who forget to set this in production will silently hit localhost.

3. **Go module path** — The install instruction is `go get sdk` and import path is `"sdk"` (not a qualified module path like `github.com/org/qhub/sdk`). This is a non-standard, non-publishable module path. If the Go SDK is intended for external distribution, this needs a proper vanity or VCS-based import path.

---

## Recommendations

1. **Fix the Go version discrepancy** — Update `apps/sdk/README.md` line 7 from `Go 1.25+` to `Go 1.26+` to match the actual project requirement.

2. **Fix the Python "Zero external dependencies" claim** — Change the feature bullet to something accurate, e.g., "Minimal dependencies (httpx + Pydantic v2 only)", to avoid misleading users.

3. **Assign a real Go module path** — Replace the `sdk` bare import with a proper module path (e.g., `github.com/<org>/qhub/sdk`) in both `go.mod` and all README install/import examples, making the SDK actually `go get`-able from outside the monorepo.

4. **Clarify dual-auth behavior** — Add a note in all three SDK READMEs explaining that API keys are sent via a different mechanism than Bearer tokens (or confirm the SDK implementation transparently handles this), so users don't assume `NewClient(apiKey.Key)` works identically to `NewClient(bearerToken)`.

5. **Add a "Defaults & Configuration" section** — Document the `base_url`/`baseUrl` default (`http://localhost:8080`) prominently near the constructor docs in each README, with a warning that production use requires explicit configuration.
