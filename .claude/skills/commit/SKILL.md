---
name: commit
description: Create a Git commit following project conventions
args: Optional message or files to commit
---

# Commit Skill

Create a Git commit following project conventions.

## Usage

```bash
# Commit all changes with auto-generated message
/commit

# Commit with custom message
/commit [message]

# Commit specific files
/commit [files...] -m [message]
```

## Implementation

When invoked:

1. **Verify Git Status**:
   ```bash
   git status  # Never use -uall flag
   ```

2. **Review Changes**:
   ```bash
   git diff          # Staged changes
   git diff HEAD     # All changes
   ```

3. **Analyze Recent Commits** (for message style):
   ```bash
   git log --oneline -10
   ```

4. **Generate Commit Message**:
   - Analyze the nature of changes (feature, fix, refactor, test, docs, chore)
   - Use conventional commit format:
     - `feat:` - New feature
     - `fix:` - Bug fix
     - `refactor:` - Code refactoring
     - `test:` - Test changes
     - `docs:` - Documentation
     - `chore:` - Maintenance tasks
   - Keep first line under 72 characters
   - Add detailed description if needed
   - Include `Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>`

5. **Stage Files**:
   - **Prefer staging specific files** by name
   - Avoid `git add -A` or `git add .` (can include sensitive files)
   - Never commit:
     - `.env` files
     - `credentials.json`
     - API keys or secrets
     - Large binaries

6. **Create Commit**:
   ```bash
   git add [specific-files]
   git commit -m "$(cat <<'EOF'
   [commit message]

   [detailed description]

   Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
   EOF
   )"
   ```

7. **Verify Success**:
   ```bash
   git status
   git log -1 --stat
   ```

## Pre-commit Hook Handling

If pre-commit hook fails (fmt, vet):
- ✅ Fix the issue
- ✅ Re-stage files
- ✅ Create a **NEW commit** (do NOT use `--amend`)
- ❌ Never use `--no-verify` unless explicitly requested
- ❌ Never use `git commit --amend` after hook failure (would modify previous commit)

## Example Commits

```bash
# Feature
feat: add KeyShard concurrent worker pool

Implement key-sharded worker pool for distributing tasks across
goroutines based on key hashing. Provides guaranteed serial execution
per key while allowing parallel execution across different keys.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>

# Bug fix
fix: correct validation logic in TaskTitle.New()

Empty strings were not being rejected. Added explicit check for
empty title strings in value object constructor.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>

# Refactor
refactor: use Result monad in task handlers

Replace traditional error handling with Result[T, E] monad pattern
for better composability and error handling.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

## Important Notes

- **Only commit when user explicitly requests it**
- Review changes before committing
- Never skip hooks unless explicitly authorized
- Never force push to main/master
- Use meaningful, descriptive commit messages
- Focus on "why" not "what" (code shows what)
