# Suggested Commands

## Local Services
```bash
docker compose up -d          # Start all services (api, web, db)
docker compose up -d db       # Start only PostgreSQL
make run-api                  # Run API server on port 8080
make run-web                  # Run web server on port 3000
make run-all                  # Run migrations + start API and web
```

## Testing
```bash
make test                     # Run all tests (requires DB running)
make check                    # fmt + vet + lint + cspell + test
cd apps/api && go test -run TestFoo ./src/routes/tasks/  # Single test
go test -cover ./...          # Coverage report
```

## Code Quality
```bash
make fmt                      # Format Go code
make vet                      # Static analysis
make lint                     # golangci-lint
make fix                      # go fix
make tidy                     # go mod tidy
make cspell                   # Spell check
```

## Database
```bash
make atlas-diff NAME=<name>   # Generate migration from schema changes
make atlas-apply              # Apply pending migrations
make atlas-new NAME=<name>    # Create new migration file
make atlas-status             # Show migration status
make sqlc-gen                 # Generate Go code from SQL queries
make sqlc-compile             # Validate SQL queries
```

## Frontend (templ)
```bash
make templ-gen                # Generate Go code from .templ templates
make templ-watch              # Watch and regenerate
make templ-fmt                # Format .templ files
```

## Terraform
```bash
make tf-fmt                   # Format Terraform files
cd apps/iac/environments/dev && terraform plan
```

## Git
```bash
git status                    # Check changes
git diff                      # View unstaged changes
git log --oneline -10         # Recent history
```

## Building
```bash
cd apps && go build ./api/... ./pkgs/...   # Build api and pkgs
cd apps && go vet ./api/... ./pkgs/...     # Vet all
```
