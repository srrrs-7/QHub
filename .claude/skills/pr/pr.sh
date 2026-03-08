#!/bin/bash
# GitHub PR creation helper script
# Usage: ./scripts/pr.sh [--draft]

set -euo pipefail

# Check if gh CLI is installed
if ! command -v gh &> /dev/null; then
    echo "❌ Error: gh CLI is not installed"
    echo "Install with: brew install gh (macOS) or apt install gh (Linux)"
    exit 1
fi

# Check not on main/master
CURRENT_BRANCH=$(git branch --show-current)
if [[ "$CURRENT_BRANCH" == "main" ]] || [[ "$CURRENT_BRANCH" == "master" ]]; then
    echo "❌ Error: Cannot create PR from main/master branch"
    exit 1
fi

echo "🔍 Analyzing changes for PR..."
echo ""
echo "Branch: $CURRENT_BRANCH"
echo ""

# Show commits that will be in PR
echo "📝 Commits in this PR:"
git log main..HEAD --oneline
echo ""

# Show diff stats
echo "📊 Changes:"
git diff main...HEAD --stat
echo ""

# Check if branch is pushed
if ! git rev-parse --abbrev-ref --symbolic-full-name @{u} &> /dev/null; then
    echo "⚠️  Branch not pushed to remote. Pushing now..."
    git push -u origin "$CURRENT_BRANCH"
    echo ""
fi

# Parse arguments
DRAFT_FLAG=""
if [[ "${1:-}" == "--draft" ]]; then
    DRAFT_FLAG="--draft"
    echo "📝 Creating draft PR..."
else
    echo "📝 Creating PR..."
fi

# Let gh CLI handle the PR creation with editor
# This opens an editor for title and description
gh pr create $DRAFT_FLAG

echo ""
echo "✅ Pull request created successfully!"
