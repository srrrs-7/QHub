# Insights Report: 459f347

## Summary

This commit adds automated codebase insights generation via a post-commit git hook. After each non-trivial commit (3+ changed files), the hook runs `claude -p` in the background to analyze the changed files and produce a markdown report saved to `.claude/reports/`. Supporting changes include a `.gitkeep` for the reports directory, `.gitignore` rules to exclude generated reports, and Makefile updates to ensure the post-commit hook gets proper permissions during `hooks-install`.

## Code Quality

No Go application code was changed — this commit is purely tooling/infrastructure (shell script, Makefile, gitignore). The Result monad, value object, and clean architecture patterns are not applicable here.

## Potential Issues

- **Command injection risk in the heredoc prompt** (line 44-81 of `.githooks/post-commit`): The variables `${COMMIT_MSG}`, `${DIFF_STAT}`, and `${CHANGED_FILE_LIST}` are interpolated directly into the prompt string without sanitization. A crafted commit message containing shell metacharacters or prompt injection content could alter the `claude -p` behavior. Using a non-interpolating heredoc (`<<'PROMPT'`) and passing variables separately would be safer, though the practical risk is low since commit messages come from the local user.
- **Silent failure**: If `claude -p` exits with an error, the report file may contain error output or be empty. The empty-file cleanup (line 85-87) handles zero-byte files, but partial/corrupt output would persist.
- **`hooks-uninstall` destroys the post-commit hook**: The `hooks-uninstall` target runs `rm -rf .githooks`, which would delete the version-controlled `post-commit` script. Since this file is tracked in git it can be recovered, but this is surprising behavior — uninstall should only unset `core.hooksPath`, not delete tracked files.

## Recommendations

1. **Fix `hooks-uninstall` to preserve tracked files** — Remove only the generated hook scripts (`pre-commit`, `pre-push`) or just unset `core.hooksPath` without deleting the directory. The current `rm -rf .githooks` will delete the committed `post-commit` script.
2. **Sanitize heredoc interpolation** — Switch to a non-interpolating heredoc and pass commit context via a temp file or environment variable to avoid potential shell escaping issues with unusual commit messages.
3. **Add error handling for `claude -p`** — Check the exit code and remove or flag reports generated from failed runs, not just empty ones.
4. **Consider a configurable skip mechanism** — Allow disabling insights via an environment variable (e.g., `SKIP_INSIGHTS=1`) for CI environments or batch operations where background `claude` invocations are undesirable.
