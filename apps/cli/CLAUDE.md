# CLAUDE.md - CLI Module

This file provides guidance to Claude Code when working with the `apps/cli` module.

## Overview

Cobra-based CLI tool (`qhub`) for interacting with the backend API from the command line.

## Structure

```
main.go              # Entry point, executes root command
cmd/
  root.go            # Root command, global flags, API client helpers
  org.go             # Organization commands (get, create)
  project.go         # Project commands (list, get, create, delete)
  prompt.go          # Prompt commands (list, get, create)
  version.go         # Version commands (list, get, create, promote, status, diff, text-diff, lint)
  log.go             # Execution log commands (list, get, create)
  eval.go            # Evaluation commands (list, get, create)
  analytics.go       # Analytics commands (prompt, version, project, trend)
  search.go          # Semantic search commands (semantic, status)
  tag.go             # Tag commands (list, create, delete, add, remove, list-by-prompt)
  member.go          # Member commands (list, add, update, remove)
  apikey.go          # API key commands (list, create, delete)
  consulting.go      # Consulting commands (sessions, session, create-session, messages, send)
  industry.go        # Industry commands (list, get, create, update, benchmarks, compliance-check)
```

## Global Flags

- `--api-url`: Backend API URL (default: `http://localhost:8080`, env: `QHUB_API_URL`)
- `--token`: Bearer auth token (default: `dev-token`, env: `QHUB_TOKEN`)
- `--output` / `-o`: Output format (`json`, `table`)

## Command Hierarchy

```
qhub org get|create
qhub project --org <id> list|get|create|delete
qhub prompt --project <id> list|get|create
qhub version --prompt <id> list|get|create|promote|status|diff|text-diff|lint
qhub log list|get|create
qhub eval list|get|create
qhub analytics prompt|version|project|trend
qhub search semantic|status
qhub tag list|create|delete|add|remove|list-by-prompt
qhub member --org <id> list|add|update|remove
qhub apikey --org <id> list|create|delete
qhub consulting sessions|session|create-session|messages|send
qhub industry list|get|create|update|benchmarks|compliance-check
```

## Key Patterns

- Each command file defines subcommands using Cobra's `AddCommand()`
- API communication via `apiGet/apiPost/apiPut/apiDelete` helpers in root.go
- Bearer token auth on all requests, 30s timeout
- JSON pretty-print output via `printJSON()`
- File/stdin input support: `--file -` or `--content-file -`
- Path helper functions for nested resources (e.g., `memberPath()`, `apikeyPath()`)

## Commands

```bash
# Build and run
cd apps/cli && go run main.go --help

# Examples
qhub org get my-org
qhub project --org <uuid> list
qhub version --prompt <uuid> lint 3
qhub search semantic "customer support prompt" --org <uuid>
qhub analytics trend <prompt-id> --days 30
qhub industry compliance-check healthcare --content "Your prompt text"
```
