# CLAUDE.md - CLI Module

This file provides guidance to Claude Code when working with the `apps/cli` module.

## Overview

Cobra-based CLI tool (`qhub`) for interacting with the backend API from the command line. Table output by default, JSON via `-o json`.

## Structure

```
main.go              # Entry point, executes root command
cmd/
  root.go            # Root command, global flags, API client helpers
  output.go          # Output formatting: printResult(), printTable(), colored messages
  table.go           # Resource-specific table printers (org, project, prompt, etc.)
  health.go          # Health check (no auth)
  config.go          # Show current configuration
  admin.go           # Admin commands (batch aggregate)
  org.go             # Organization commands (get, create, update)
  project.go         # Project commands (list, get, create, update, delete)
  prompt.go          # Prompt commands (list, get, create, update)
  version.go         # Version commands (list, get, create, promote, status, diff, text-diff, lint, compare)
  log.go             # Execution log commands (list, get, create)
  eval.go            # Evaluation commands (list, get, create, update)
  analytics.go       # Analytics commands (prompt, version, project, trend)
  search.go          # Semantic search commands (semantic, status)
  tag.go             # Tag commands (list, create, delete, add, remove, list-by-prompt)
  member.go          # Member commands (list, add, update, remove)
  apikey.go          # API key commands (list, create, delete)
  consulting.go      # Consulting commands (sessions, session, create-session, close, messages, send)
  industry.go        # Industry commands (list, get, create, update, benchmarks, compliance-check)
```

## Global Flags

- `--api-url`: Backend API URL (default: `http://localhost:8080`, env: `QHUB_API_URL`)
- `--token`: Bearer auth token (default: `dev-token`, env: `QHUB_TOKEN`)
- `--output` / `-o`: Output format — `table` (default) or `json`
- `--version` / `-v`: Show version

## Command Hierarchy

```
qhub health
qhub config
qhub org get|create|update
qhub project --org <id> list|get|create|update|delete
qhub prompt --project <id> list|get|create|update
qhub version --prompt <id> list|get|create|promote|status|diff|text-diff|lint|compare
qhub log list|get|create
qhub eval list|get|create|update
qhub analytics prompt|version|project|trend
qhub search semantic|status
qhub tag list|create|delete|add|remove|list-by-prompt
qhub member --org <id> list|add|update|remove
qhub apikey --org <id> list|create|delete
qhub consulting sessions|session|create-session|close|messages|send
qhub industry list|get|create|update|benchmarks|compliance-check
qhub admin aggregate
```

## Key Patterns

### Output
- Default output is **table** format via `text/tabwriter`
- JSON available via `-o json` for scripting/piping
- Use `printResult(v)` for generic dispatch, or resource-specific printers (e.g., `printOrgTable()`)
- `printSuccess()`, `printInfo()`, `printWarning()` for colored stderr messages
- Single objects render as vertical key-value; arrays render as columnar tables

### API Communication
- `apiGet/apiPost/apiPut/apiDelete` helpers in root.go
- Bearer token auth on all requests, 30s timeout
- Health endpoint skips auth (direct HTTP call)

### Input
- File/stdin input support: `--file -` or `--content-file -`
- Path helper functions for nested resources (e.g., `memberPath()`, `apikeyPath()`)

## Commands

```bash
# Build
make build-cli                # Builds to bin/qhub

# Run directly
cd apps/cli && go run main.go --help

# Examples
qhub health
qhub config
qhub org create --name "My Org" --slug my-org
qhub org get my-org
qhub org update my-org --plan pro
qhub project --org <uuid> list
qhub prompt --project <uuid> create --name "My Prompt" --slug my-prompt
qhub version --prompt <uuid> create --content "You are a helpful assistant"
qhub version --prompt <uuid> lint 3
qhub version --prompt <uuid> compare 1 2
qhub search semantic "customer support" --org <uuid>
qhub analytics trend <prompt-id> --days 30
qhub consulting create-session --org <uuid> --title "Review my prompts"
qhub industry compliance-check healthcare --content "Your prompt text"
qhub admin aggregate
```
