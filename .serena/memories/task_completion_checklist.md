# Task Completion Checklist

When a task is completed, ensure the following:

## Code Quality
1. `make fmt` - Format all Go code
2. `make vet` - Run static analysis
3. `make lint` - Run golangci-lint
4. `go build ./api/... ./pkgs/...` - Verify compilation

## Testing
1. All new code must have tests written FIRST (TDD)
2. Cover all 6 test categories (正常系, 異常系, 境界値, 特殊文字, 空文字, Null/Nil)
3. `make test` - Run all tests (requires DB: `docker compose up -d db`)
4. Coverage must be ≥80% overall and per function

## Before Commit
1. No secrets in staged files (.env, credentials, API keys)
2. Stage specific files (avoid `git add -A`)
3. Use conventional commit format
4. Run `make check` for full validation (fmt + vet + lint + cspell + test)

## If Editing Database
1. Update schema → `make atlas-diff NAME=description`
2. Review migration → `make atlas-apply`
3. Update queries → `make sqlc-gen`

## If Editing Templates
1. Edit `.templ` files → `make templ-gen`
