---
name: test
description: Run tests with proper database setup
args: Optional package or test name
---

# Test Skill

Run tests with proper database setup.

## Prerequisites

- PostgreSQL must be running: `docker compose up -d db`
- Database migrations are applied automatically

## Usage

```bash
# Run all tests
/test

# Run specific package tests
/test [package]

# Run specific test
/test [test-name]
```

## Implementation

When invoked:

1. Verify database is running: `docker compose ps db`
2. Apply migrations: `make atlas-apply`
3. Run appropriate test command:
   - No args: `make test`
   - Package specified: `cd apps/[package] && go test -cover ./...`
   - Test name specified: `cd apps/api && go test -run [test-name] ./...`
4. Report results with coverage percentage

## Examples

```bash
# All tests
/test

# API tests only
/test api

# Specific handler test
/test TestListHandler
```
