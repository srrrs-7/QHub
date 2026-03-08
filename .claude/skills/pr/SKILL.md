---
name: pr
description: Create a GitHub pull request following best practices
args: Optional title, --base [branch], --draft
---

# Pull Request Skill

Create a GitHub pull request following best practices.

## Usage

```bash
# Create PR with auto-generated title and description
/pr

# Create PR with custom title
/pr [title]

# Create PR to specific base branch
/pr --base [branch]

# Create draft PR
/pr --draft
```

## Implementation

When invoked:

1. **Verify Prerequisites**:
   ```bash
   # Check gh CLI is installed
   gh --version

   # Check current branch
   git branch --show-current

   # Verify not on main/master
   if [ "$(git branch --show-current)" = "main" ]; then
     echo "ERROR: Cannot create PR from main branch"
     exit 1
   fi
   ```

2. **Analyze Changes** (understand full context):
   ```bash
   # Current branch status
   git status  # Never use -uall

   # All changes since branching
   git diff main...HEAD

   # All commits in this PR
   git log main..HEAD --oneline
   git log main..HEAD --stat
   ```

   **CRITICAL**: Analyze **ALL commits** that will be included, not just the latest!

3. **Check Remote Status**:
   ```bash
   # Check if branch tracks remote
   git rev-parse --abbrev-ref --symbolic-full-name @{u}

   # Check if up to date with remote
   git status -sb
   ```

4. **Generate PR Content**:

   **Title** (< 70 characters):
   - Use conventional commit format: `feat:`, `fix:`, `refactor:`, etc.
   - Concise summary of the change
   - Example: `feat: add email field to tasks table`

   **Description/Body**:
   ```markdown
   ## Summary
   - High-level overview of changes
   - Why this change is needed
   - What problem it solves

   ## Changes
   - Bullet points of specific changes
   - One bullet per major change
   - Include file paths for context

   ## Testing
   - [ ] Unit tests added/updated
   - [ ] Integration tests pass
   - [ ] Manual testing performed
   - [ ] Database migrations tested

   ## Additional Notes
   - Any caveats or considerations
   - Deployment notes if needed
   - Related issues or PRs

   🤖 Generated with [Claude Code](https://claude.com/claude-code)
   ```

5. **Push if Needed**:
   ```bash
   # If branch not pushed yet
   git push -u origin $(git branch --show-current)

   # If branch exists but not up to date
   git push
   ```

6. **Create PR**:
   ```bash
   gh pr create \
     --title "[title]" \
     --body "$(cat <<'EOF'
   ## Summary
   [summary content]

   ## Changes
   [changes list]

   ## Testing
   - [ ] Unit tests added/updated
   - [ ] Integration tests pass
   - [ ] Manual testing performed

   🤖 Generated with [Claude Code](https://claude.com/claude-code)
   EOF
   )"
   ```

   **Options**:
   - `--base main` - Target branch (default: main)
   - `--draft` - Create as draft PR
   - `--assignee @me` - Assign to yourself
   - `--label enhancement` - Add labels

7. **Display PR URL**:
   ```bash
   # gh pr create returns URL
   echo "✓ Pull request created: [URL]"
   ```

## PR Content Best Practices

### Good PR Titles
✅ `feat: add user authentication with JWT`
✅ `fix: resolve race condition in task processing`
✅ `refactor: migrate to Result monad pattern`
✅ `test: add integration tests for task API`

❌ `Update code`
❌ `Fix bug`
❌ `WIP: lots of changes`

### Good PR Descriptions

**Example 1: Feature PR**
```markdown
## Summary
Adds user authentication using JWT tokens with refresh token rotation.
Implements secure session management with Redis backing store.

## Changes
- Added `auth` package with JWT token generation/validation
- Implemented middleware for protected routes
- Added Redis session store for refresh tokens
- Updated API routes to require authentication
- Added login/logout endpoints

## Testing
- [x] Unit tests for auth package (92% coverage)
- [x] Integration tests for login/logout flow
- [x] Manual testing with Postman
- [x] Redis connection tested in Docker environment

## Security Considerations
- Tokens expire after 15 minutes
- Refresh tokens rotate on use
- Passwords hashed with bcrypt (cost factor 12)

🤖 Generated with [Claude Code](https://claude.com/claude-code)
```

**Example 2: Bug Fix PR**
```markdown
## Summary
Fixes race condition in parallel task processing that caused duplicate
task execution under high load.

## Changes
- Added mutex to `KeyShard` worker pool
- Fixed task deduplication in `Process()` method
- Updated tests to verify concurrent access

## Root Cause
The task queue was not protected from concurrent access, allowing
multiple workers to dequeue the same task.

## Testing
- [x] Added race condition test
- [x] Verified with `go test -race`
- [x] Load tested with 1000 concurrent requests

🤖 Generated with [Claude Code](https://claude.com/claude-code)
```

## Example Workflows

### Standard Feature PR
```bash
/pr
```
Analyzes commits, generates title and description, creates PR.

### Draft PR for Early Review
```bash
/pr --draft
```
Creates draft PR for getting early feedback before completion.

### PR to Staging Branch
```bash
/pr --base staging
```
Creates PR targeting staging branch instead of main.

## Important Notes

- **Only create PR when user explicitly requests it**
- Analyze **ALL commits** in the PR, not just the latest
- Keep PR titles under 70 characters
- Use description for details, not title
- Include testing checklist
- Link related issues with `Fixes #123`
- Add appropriate labels if available
- Return PR URL for easy access
- Never create PR from main branch
- Verify branch is pushed before creating PR

## Integration with /commit and /push

Typical workflow:
```bash
# 1. Make changes and commit
/commit

# 2. Push to remote
/push -u

# 3. Create pull request
/pr
```

Or combined:
```bash
# Make changes, commit, push, and create PR
/commit && /push -u && /pr
```
