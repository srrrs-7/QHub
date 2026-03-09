---

## Commit `132b190` — Insights Report

### Summary

This refactor decomposed a large, monolithic `prompt_detail.templ` into reusable sub-components (`VersionDetail`, `VersionItem`, `VersionCompareCard`, `LintResultCard`, `TextDiffCard`, `SemanticDiffCard`, `PromptTagsSection`) and extracted four small helper templ components (`statusTransitionBtn`, `lintSeverityIcon`, `impactBadge`, `compareWinnerBadge`). CSS that had been duplicated across templates (version list, prompt content blocks) was consolidated into `layout.templ` under clearly-labelled compat sections. `projects.templ` similarly extracted `ProjectList`, `ProjectEditDialog`, and `ProjectDeleteDialog` into independent components for HTMX partial-swap targets.

---

### Code Quality

| Concern | Finding |
|---|---|
| Result monad / error returns | N/A — template layer; no error returns expected |
| Value objects | Templates correctly consume `client.*` types; no domain bypassing |
| Clean architecture | Templates only import `web/src/client` — dependency direction respected |
| Helper functions | `lintScoreClass`, `contentLines`, `availableTags` are pure and appropriately placed |

The extracted components are well-scoped and improve reusability without introducing cross-layer coupling.

---

### Potential Issues

**1. `PromptHeaderUpdated` partial produces an incomplete swap target (medium severity)**

`PromptDetailPage` wraps the entire sidebar header in `<div class="sidebar-header" id="prompt-header">`, which includes the name, edit button, type badge, status badge, and description. The `PromptHeaderUpdated` component (`prompt_detail.templ:377`) emits only the name+edit-button row as `id="prompt-header"`, losing the type/status badges after an edit. The hx-target `outerHTML` swap will silently strip those elements.

```
// prompt_detail.templ:378 — narrow render vs full sidebar-header at line 52
```

**2. `availableTags` called twice per render (low severity)**

At lines 354 and 365, `availableTags(allTags, promptTags)` is evaluated twice in the same template execution (once for the guard `if`, once in the `for` range). For large tag sets this is wasteful, though currently negligible.

**3. `hx-vals` JSON built by string concatenation (low severity)**

`statusTransitionBtn` at line 193 constructs JSON via string concatenation:
```go
hx-vals={ `{"status":"` + nextStatus + `"}` }
```
`nextStatus` is always a hardcoded literal at every call site, so there is no injection risk today. However, if this helper is ever reused with a runtime value, it would be unsafe. Consider using `templ.JSONString` or a format function.

**4. Compat CSS dead weight (low severity)**

`layout.templ` retains `/* ─── Version List (compat) ─── */` and `/* ─── Prompt Content (compat) ─── */` blocks (lines 536–552). If `prompt_detail.templ` no longer uses `.version-list`/`.version-item`/`.prompt-content` classes directly, these can be removed.

---

### Test Coverage Concerns

- **`lintScoreClass`, `contentLines`, `availableTags`** are new Go functions with no unit tests. They contain branching logic (score thresholds, set-difference) that should be covered per project TDD rules.
- New templ components (`LintResultCard`, `TextDiffCard`, `SemanticDiffCard`, `VersionCompareCard`) are not exercised by `handlers/pages_test.go` or `handlers/partials_test.go` unless the relevant partial handler tests were updated. Verify coverage for the lint, diff, and compare partial handlers.
- The 6 mandatory test categories (happy path, error, boundary, special chars, empty, nil) are likely **not met** for `availableTags` (no tests for nil slices, empty inputs) and `lintScoreClass` (no boundary tests at scores 79/80 and 59/60).

---

### Recommendations

1. **Fix `PromptHeaderUpdated` mismatch** — either render the full sidebar-header block (including badges) or use `hx-swap-oob` to independently update each sub-element; otherwise the edit flow silently removes type/status badges.

2. **Add unit tests for `lintScoreClass`, `contentLines`, `availableTags`** — all three have non-trivial branching that the mandatory 6-category test coverage requires.

3. **Add partial handler tests for lint/diff/compare results** — ensure `LintResultCard`, `TextDiffCard`, `SemanticDiffCard`, and `VersionCompareCard` are rendered by at least one test per partial endpoint.

4. **Remove unused compat CSS** — confirm `.version-list`, `.version-item`, `.prompt-content` classes are unused and delete the compat blocks to reduce payload size.

5. **Replace `hx-vals` string concatenation with safe encoding** — low urgency now, but establish the pattern before `statusTransitionBtn` is reused with non-literal arguments.
