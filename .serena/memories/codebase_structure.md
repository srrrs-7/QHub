# Codebase Structure

```
/workspace/main/
├── CLAUDE.md                    # Project instructions for AI
├── Makefile                     # All dev commands
├── compose.yaml                 # Docker Compose (api, web, db)
├── .golangci.yml                # Linter config
├── .claude/rules/               # Detailed coding rules
│   ├── architecture.md
│   ├── go-patterns.md
│   ├── testing.md
│   └── tdd.md
├── .github/                     # CI/CD workflows
├── apps/
│   ├── go.work                  # Go workspace definition
│   ├── api/                     # Backend API (module: "api")
│   │   ├── go.mod
│   │   └── src/
│   │       ├── cmd/main.go      # Entry point (port 8080)
│   │       ├── domain/
│   │       │   ├── apperror/    # AppError interface & types
│   │       │   └── task/        # Task domain: models, value objects, repository interface
│   │       ├── infra/rds/
│   │       │   └── task_repository/  # PostgreSQL implementation
│   │       └── routes/
│   │           ├── routes.go         # Chi router, DI wiring
│   │           ├── middleware/       # Logger, BearerAuth
│   │           ├── response/         # JSON response helpers
│   │           └── tasks/            # CRUD handlers (handler.go, get/post/put/list.go)
│   ├── pkgs/                    # Shared packages (module: "utils")
│   │   ├── db/                  # DB connection, sqlc generated code, migrations, queries
│   │   ├── env/                 # Environment variable helpers
│   │   ├── logger/              # Structured logging (slog)
│   │   └── testutil/            # Test helpers (SetupTestTx, SetAuthHeader)
│   ├── web/                     # Frontend (templ + HTMX, port 3000)
│   │   └── src/
│   │       ├── cmd/main.go
│   │       ├── templates/       # .templ files + generated _templ.go
│   │       ├── client/          # API client
│   │       ├── handlers/        # HTTP handlers
│   │       └── routes/          # Chi router
│   ├── iac/                     # Terraform (AWS infrastructure)
│   │   ├── environments/        # dev, stg, prd
│   │   └── modules/             # vpc, ecs, ecr, alb, aurora, cognito, etc.
│   └── migrate/                 # Migration Docker image
```

## API Routes
- `GET /health` - Health check with DB ping
- `GET /api/v1/tasks` - List tasks
- `POST /api/v1/tasks` - Create task
- `GET /api/v1/tasks/{id}` - Get task by ID
- `PUT /api/v1/tasks/{id}` - Update task

All `/api/v1/*` routes require Bearer auth.
