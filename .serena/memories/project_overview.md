# Project Overview - QHub

## Purpose
QHub is a system for creating new questions and version-managing answers to them.

## Tech Stack
- **Language**: Go 1.26 with workspaces
- **Router**: go-chi/chi
- **Database**: PostgreSQL 18 (Aurora Serverless v2 in prod)
- **Migrations**: Atlas (schema-first)
- **Query Generation**: sqlc (type-safe SQL → Go)
- **Frontend**: templ + HTMX (server-side rendered, no client JS)
- **Infrastructure**: Terraform (AWS: ECS Fargate, ALB, Aurora, Cognito, CloudFront, WAF)
- **Containerization**: Docker Compose for local dev

## Architecture
Clean Architecture with dependency inversion:
- `routes/` → HTTP layer (handlers depend on domain interfaces)
- `domain/` → Business logic, interfaces, value objects (no external deps)
- `infra/` → Repository implementations (implements domain interfaces)

Dependency flow: `routes → domain ← infra`

## Module Structure
Go workspace at `apps/go.work` manages:
- `apps/api` - Backend API (port 8080)
- `apps/pkgs` - Shared packages (module name: "utils")
- `apps/web` - Frontend (port 3000)
- `apps/iac` - Terraform infrastructure

## Key Patterns
- **Idiomatic Go**: Standard `(value, error)` returns with `if err != nil`
- **Struct-based DI**: Constructor injection via structs
- **Value Objects**: `TaskID`, `TaskTitle`, `TaskDescription`, `TaskStatus`
- **AppError interface**: Domain errors with `ErrorName()`, `DomainName()`, `Unwrap()`
- **Table-driven tests**: All tests use `testName/args/expected` pattern
- **TDD mandatory**: Red → Green → Refactor → Commit
