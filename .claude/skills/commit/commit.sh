#!/bin/bash
# Git commit helper script
# Usage: ./scripts/commit.sh [message]

set -euo pipefail

echo "📋 Checking git status..."
git status

echo ""
echo "📝 Reviewing changes..."
git diff --stat

echo ""
echo "📚 Recent commit style..."
git log --oneline -5

echo ""
read -p "Continue with commit? (y/N) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Commit cancelled."
    exit 0
fi

# Generate commit message based on changes
# This is a simple implementation - Claude Code can do this better
if [ $# -eq 0 ]; then
    echo "Please provide commit message:"
    echo "Format: <type>: <description>"
    echo "Types: feat, fix, refactor, test, docs, chore"
    read -r MESSAGE
else
    MESSAGE="$*"
fi

# Stage files (list them for user to confirm)
echo ""
echo "Modified files:"
git status --short
echo ""
read -p "Stage all modified files? (y/N) " -n 1 -r
echo

if [[ $REPLY =~ ^[Yy]$ ]]; then
    # Add modified and deleted files, but not untracked
    git add -u
else
    echo "Please stage files manually with: git add <files>"
    exit 1
fi

# Create commit with Co-Authored-By
git commit -m "$(cat <<EOF
${MESSAGE}

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
EOF
)"

echo ""
echo "✅ Commit created successfully!"
git log -1 --stat
