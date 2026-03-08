---
name: push
description: Push commits to remote repository
args: Optional -u for upstream, --force for force push
---

# Push Skill

Push commits to remote repository.

## Usage

```bash
# Push current branch to remote
/push

# Push and set upstream for new branch
/push -u

# Force push (use with extreme caution)
/push --force
```

## Implementation

When invoked:

1. **Check Current Branch**:
   ```bash
   git branch --show-current
   git status
   ```

2. **Verify Commits to Push**:
   ```bash
   git log origin/$(git branch --show-current)..HEAD --oneline
   ```
   - Show user what commits will be pushed
   - Confirm there are commits to push

3. **Pre-push Checks**:
   - Verify branch is not `main` or `master` if using `--force`
   - Check for uncommitted changes (warn if found)
   - Verify remote tracking branch exists

4. **Execute Push**:

   **Standard push**:
   ```bash
   git push
   ```

   **New branch** (set upstream):
   ```bash
   git push -u origin $(git branch --show-current)
   ```

   **Force push** (with safety):
   ```bash
   # ⚠️ REQUIRES EXPLICIT USER CONFIRMATION
   # ❌ NEVER force push to main/master
   git push --force-with-lease
   ```

5. **Verify Success**:
   ```bash
   git status
   git log -1
   ```

## Safety Checks

### Force Push Protection
- ❌ **Never** force push to `main` or `master`
- ⚠️ Warn user about destructive nature
- ✅ Use `--force-with-lease` instead of `--force` (safer)
- 🔒 Require explicit user confirmation

### Pre-push Hook
If pre-push hook fails (tests):
- Show test failures
- Ask user if they want to:
  1. Fix tests and try again
  2. Skip push and investigate
  3. Override hook (requires explicit approval with `--no-verify`)

## Example Workflows

### First push (new branch)
```bash
/push -u
```
Output:
```
Branch: feature/add-email-field
Commits to push:
  abc1234 feat: add email field to tasks table
  def5678 test: add tests for email validation

Pushing to origin/feature/add-email-field...
✓ Successfully pushed 2 commits
```

### Standard push
```bash
/push
```
Output:
```
Branch: feature/add-email-field
Commits to push:
  ghi9012 refactor: extract email validation logic

Pushing to origin/feature/add-email-field...
✓ Successfully pushed 1 commit
```

### Force push (requires confirmation)
```bash
/push --force
```
Output:
```
⚠️  WARNING: Force push is destructive and can overwrite remote changes
Branch: feature/add-email-field
Remote commits that will be overwritten:
  jkl3456 fix: typo in email field

Do you want to proceed with force push? (y/N)
```

## Important Notes

- **Only push when user explicitly requests it**
- Never push without showing what will be pushed
- Protect main/master branches from force push
- Use `--force-with-lease` instead of `--force`
- Respect pre-push hooks (tests must pass)
- Always verify push success
