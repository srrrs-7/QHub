# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Go monorepo with functional programming patterns. Uses Go 1.26 with workspaces, Docker Compose for services, and PostgreSQL 18.

## Coding Guidelines

**CRITICAL: This repository requires Test-Driven Development (TDD) for all code changes.**

Detailed coding standards and patterns are documented in `.claude/rules/`:

- **architecture.md**: Layered architecture (routes → domain ← infra), clean architecture principles, API design, database migrations, error handling strategy
- **go-patterns.md**: Result monad usage (always use `Result[T, E]`, never `(value, error)`), domain model value objects, error handling with `AppError` interface, concurrency patterns with `parallel` package
- **testing.md**: Table-driven test pattern (mandatory), HTTP handler testing with `testutil`, database testing, test organization
- **tdd.md**: **TDD workflow (Red → Green → Refactor → Commit)**, required test categories (6 categories: 正常系, 異常系, 境界値, 特殊文字, 空文字, Null/Nil), coverage requirements (≥80% overall, ≥80% per function, 100% for critical paths)

**TDD is mandatory. Always follow this workflow:**

1. 🔴 **RED**: Write failing test first
2. 🟢 **GREEN**: Write minimal code to pass
3. 🔵 **REFACTOR**: Improve code quality
4. ✅ **COMMIT**: Commit after green with ≥80% coverage
5. ♻️ **REPEAT**: Next test case

**Never write production code without writing the test first.** Coverage below 80% blocks commits.

## Repository Structure

```
apps/
  api/          # Backend API (go-chi router, port 8080)
  pkgs/         # Shared packages (db, logger, env, types, parallel, testutil)
  web/          # Frontend (templ + HTMX, port 3000)
  iac/          # Terraform infrastructure (AWS)
```

Go workspace: `apps/go.work` manages `api`, `pkgs`, and `web` modules.

## Development Commands

```bash
# Local services (from repo root)
docker compose up -d          # Start all services (api, web, db)
docker compose up -d db       # Start only PostgreSQL

# Run servers locally (alternative to Docker Compose)
make run-api                  # Run API server on port 8080
make run-web                  # Run web server on port 3000 (requires make templ-gen first)
make run-all                  # Run migrations, then start API and web servers

# Tests (all modules)
make test                     # Run all tests (requires DB running)
make check                    # Run fmt, vet, lint, cspell, and test

# Run single test
cd apps/api && go test -run TestListHandler ./src/routes/tasks/

# Code quality
make fmt      # Format Go code
make vet      # Static analysis
make lint     # Run golangci-lint
make fix      # Run go fix
make tidy     # go mod tidy
make cspell   # Check spelling (misspell)
make tf-fmt   # Format Terraform files

# Database migrations (Atlas)
make atlas-diff NAME=<name>  # Generate migration from schema changes
make atlas-apply             # Apply pending migrations
make atlas-new NAME=<name>   # Create new migration file
make atlas-status            # Show migration status
# Use ATLAS_ENV=docker for Docker environment: make atlas-apply ATLAS_ENV=docker

# sqlc code generation
make sqlc-gen      # Generate Go code from SQL queries
make sqlc-compile  # Validate SQL queries

# templ template generation (web frontend)
make templ-gen     # Generate Go code from .templ templates
make templ-watch   # Watch and regenerate on template changes
make templ-fmt     # Format .templ files

# Git hooks (optional)
make hooks-install   # Install pre-commit (fmt, vet) and pre-push (test) hooks
make hooks-uninstall # Remove git hooks
```

## Architecture

### Functional Error Handling with Result Type

The codebase uses a `Result[T, E]` monad (`apps/pkgs/types/result.go`) for functional error handling instead of Go's traditional `(value, error)` pattern:

```go
// Pipeline example from handlers
res := types.Pipe2(
    newListRequest(r).validate(),
    func(req listRequest) types.Result[[]model.Task, model.AppError] {
        return task_repository.FindAllTasks()
    },
    func(tasks []model.Task) listResponse { ... },
)

res.Match(
    func(resp listResponse) { response.OK(w, resp) },
    func(e model.AppError) { response.HandleAppError(w, e) },
)
```

Key functions: `Ok()`, `Err()`, `Map()`, `FlatMap()`, `Pipe2-5()`, `Match()`, `Combine()`

### Domain Model Pattern

- Value objects with type safety: `TaskID`, `TaskTitle`, `TaskDescription`, `TaskCompleted`
- Domain errors implement `AppError` interface with `ErrorName()` and `DomainName()`
- Error types: `ValidationError`, `NotFoundError`, `DatabaseError`, etc.

### API Layer Structure (apps/api/src)

```
cmd/main.go              # Entry point, graceful shutdown
routes/
  routes.go              # Chi router setup, /api/v1 prefix
  response/response.go   # JSON response helpers, error mapping
  tasks/                 # Handler per endpoint (list.go, post.go, get.go, put.go)
domain/model/            # Domain types and errors
infra/rds/               # Repository implementations
```

Route pattern: `/api/v1/tasks`, `/api/v1/tasks/{id}`

### Web Frontend (apps/web)

Go server-side rendered frontend using templ + HTMX:
- `cmd/main.go` - Entry point (port 3000), graceful shutdown, API client initialization
- `templates/*.templ` - Type-safe Go templates compiled to `*_templ.go` files
- `handlers/` - HTTP handlers returning templ components
- `routes/` - Chi router, serves full pages and HTMX partials
- `client/` - API client for communicating with backend API

**templ workflow**:
1. Edit `.templ` files in `templates/`
2. Run `make templ-gen` to generate Go code
3. Templates are type-safe and compile-time checked

**HTMX pattern**: Handlers return full pages or partial HTML fragments for dynamic updates without JavaScript.

Dockerfile: Multi-stage build (Go + templ → nginx for serving)

### Database Layer (apps/pkgs/db)

- **Atlas**: Schema-first migrations in `migrations/`
- **sqlc**: Type-safe query generation from `queries/*.sql` → `db/*.go`

Configuration files:
- `atlas.hcl` - Migration environments (local, docker, ci)
- `sqlc.yaml` - Code generation config

### Shared Packages (apps/pkgs)

**Core packages**:
- `types/` - `Result[T, E]` monad for functional error handling
- `env/` - Environment variable utilities
- `logger/` - Structured logging
- `db/` - Database connection and sqlc-generated queries

**Utilities**:
- `parallel/` - Concurrent execution helpers (`Parallel2`-`Parallel5` for running functions concurrently, `KeyShard` for key-sharded worker pools)
- `testutil/` - Test helpers (HTTP testing, DB testing)

## Testing Patterns

**All code must follow TDD workflow** (see `.claude/rules/tdd.md` for details).

Tests use table-driven pattern with `args`/`expected` structs:

```go
tests := []struct {
    testName string
    args     args
    expected expected
}{...}

for _, tt := range tests {
    t.Run(tt.testName, func(t *testing.T) { ... })
}
```

**Required test coverage**:
- ≥80% overall package coverage
- ≥80% per function
- 100% for critical paths (value objects, error handling, repositories, handlers)

**Required test categories** (cover all 6 for every function):
1. ✅ 正常系 (Happy Path): Valid inputs
2. ❌ 異常系 (Error Cases): Invalid inputs
3. 📏 境界値 (Boundary Values): Min/max, zero, negative
4. 🔤 特殊文字 (Special Chars): Unicode, emoji, SQL injection attempts
5. 📭 空文字 (Empty String): Empty, whitespace-only
6. ⚠️ Null/Nil: Nil pointers, zero values, empty slices

HTTP handlers tested with `httptest.NewRequest` and `httptest.NewRecorder`.

## Infrastructure (apps/iac)

Terraform modules for AWS deployment:

```
apps/iac/
  environments/       # Environment configs (dev, stg, prd)
  modules/           # Reusable Terraform modules
    vpc/             # Multi-AZ VPC with public/private/database subnets
    ecs/             # ECS Fargate service
    ecr/             # Container registry
    alb/             # Application Load Balancer
    aurora/          # Aurora PostgreSQL Serverless v2
    cognito/         # Authentication (MFA-enabled)
    cloudfront/      # CDN (S3 + ALB origins)
    s3/              # Static assets
    waf/             # Web Application Firewall
    acm/             # SSL/TLS certificates
    route53/         # DNS records
    iam/             # GitHub Actions OIDC, ECS task roles
    security-groups/ # ALB, ECS, Aurora security groups
```

Naming convention: `${project}-${environment}-${resource}`

Terraform setup:
```bash
cd apps/iac/environments/dev
cp backend.hcl.example backend.hcl     # Configure S3 state backend
cp terraform.tfvars.example terraform.tfvars  # Set environment variables
terraform init -backend-config=backend.hcl
terraform plan && terraform apply
```

## Git Workflow

### Committing Changes

**Always follow when creating commits**:
- Use conventional commit format: `feat:`, `fix:`, `refactor:`, `test:`, `docs:`, `chore:`
- Stage specific files (avoid `git add -A`)
- Never commit: `.env`, `credentials.json`, API keys, secrets
- Include `Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>`
- If pre-commit hook fails: fix issue and create NEW commit (never `--amend`)

**Commit message format**:
```bash
git commit -m "$(cat <<'EOF'
feat: add email field to tasks table

Detailed description of changes and why they were made.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
EOF
)"
```

### Creating Pull Requests

**PR creation checklist**:
1. Analyze ALL commits (not just latest): `git log main..HEAD`
2. Review all changes: `git diff main...HEAD`
3. Title: < 70 chars, conventional format (e.g., `feat: add user auth`)
4. Body structure:
   ```markdown
   ## Summary
   - High-level overview

   ## Changes
   - Specific changes with file paths

   ## Testing
   - [ ] Unit tests added/updated
   - [ ] Integration tests pass
   - [ ] Manual testing performed

   🤖 Generated with [Claude Code](https://claude.com/claude-code)
   ```
5. Use `gh pr create` with full description

### Push Safety

- Never force push to `main`/`master`
- Use `--force-with-lease` instead of `--force`
- Show what will be pushed before pushing
- Respect pre-push hooks (tests must pass)

## CI/CD Pipeline

**CI**: GitHub Actions runs in devcontainer on push/PR to main: `make vet && make test`

**CD**: Triggered by push to `main` (→ dev) or manual workflow dispatch (→ dev/stg/prd)
- Flow: Database Migration → Build & Push to ECR → Deploy to ECS
- Environments configured in GitHub Settings → Environments with AWS OIDC credentials
