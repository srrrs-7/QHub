Here is the insights report for commit `7bafb66`:

---

## Summary

This commit replaces the Material Design 3 (M3) component library with a hand-rolled GitHub Primer-inspired design system. The entire visual language — color tokens, typography scale, component CSS, and layout patterns — has been migrated to a flat, GitHub-style aesthetic (`--canvas`, `--fg`, `--accent`, `--header-bg`). M3 CSS variable names are retained as aliases for backward compatibility. The most significant structural change is in `prompt_detail.templ`, which gains a commit-timeline sidebar, a GitHub-style code viewer with line numbers, and inline lint/diff/compare tooling. Four new pure Go helper functions are introduced in the template layer. All `*_templ.go` files were regenerated.

---

## Code Quality

**Not applicable (Result[T, E] monad pattern):** This is a frontend template layer — the project uses standard Go `(value, error)` returns. The new helper functions (`maxInt`, `lintScoreClass`, `contentLines`, `availableTags`) follow that convention correctly.

**Value object usage:** Template functions accept `*client.Prompt`, `[]client.PromptVersion`, etc. — client-layer structs rather than domain value objects, which is the correct pattern for the web module's API client boundary.

**Clean architecture:** The web module correctly depends only on `client.Client` (the API abstraction), not on any `api` domain packages. No layering violations introduced.

---

## Test Coverage Concerns

Four new Go helper functions in `prompt_detail.templ` have no corresponding tests:

| Function | Risk |
|---|---|
| `availableTags(all, current []client.Tag) []client.Tag` | Set-subtraction logic; missing boundary cases (nil slices, duplicate IDs) |
| `lintScoreClass(score int) string` | Score thresholds (80, 60) untested at boundaries |
| `contentLines(content string) []string` | Edge cases: empty string, Windows `\r\n` line endings |
| `maxInt(a, b int) int` | Trivial but coverage-policy requires tests |

None of the 6 mandatory test categories (happy path, error, boundary, special chars, empty, nil) are covered for these functions. They live in a `.templ` file where the generated code can be tested via `go test` on the `templates` package.

---

## Potential Issues

**1. HTMX swap regression in `PromptHeaderUpdated` (bug)**

The edit-prompt form targets `#prompt-header` with `hx-swap="outerHTML"`. In the full page, `#prompt-header` is assigned to the entire `.sidebar-header` div which contains three children: the name/edit-button row, the type/status badge row, and the description. `PromptHeaderUpdated` returns a replacement `<div id="prompt-header">` that contains only the name/edit-button row — the type badge and status badge are lost after a successful edit.

Relevant lines: `prompt_detail.templ:58` (original) vs `prompt_detail.templ:398-410` (partial).

**2. No SRI hashes on CDN resources**

`layout.templ:18-19` loads HTMX and htmx-ext-sse from `unpkg.com` without [Subresource Integrity](https://developer.mozilla.org/en-US/docs/Web/Security/Subresource_Integrity) hashes. A compromised CDN could inject arbitrary JavaScript. Same applies to Google Fonts (lower risk, CSS only).

**3. All CSS inlined — no browser caching**

The entire stylesheet (~800 lines) is embedded in every HTML response. Browsers cannot cache it independently. This inflates every page response by ~20 KB and negates one of the primary benefits of a shared CSS file.

**4. `gap-6` utility class missing**

`prompt_detail.templ:379` uses `class="flex gap-6 items-center"` but the utility stylesheet in `layout.templ` only defines `gap-8`, `gap-16`, `gap-24`. The tag-add form will render with no gap.

---

## Recommendations

1. **Fix the `PromptHeaderUpdated` swap regression** — either widen the returned fragment to include the badge row, or change the HTMX target to `#prompt-header-inner` scoped only to the name/button row so badges are not overwritten.

2. **Add unit tests for the four helper functions** in a new `templates/helpers_test.go` file covering all 6 required categories, especially boundary values for `lintScoreClass` (79 vs 80, 59 vs 60) and nil-slice handling in `availableTags`.

3. **Add the `gap-6` utility class** to the stylesheet or replace the class with `gap-8` at `prompt_detail.templ:379`.

4. **Add SRI hashes** to the HTMX `<script>` tags, or vendor the JS files into `static/` to eliminate the CDN dependency.

5. **Extract CSS to a static file** (e.g. `static/primer.css`) served with cache headers, reducing per-response payload and enabling browser caching across page navigations.
