# CLAUDE.md - CLI Module

This file provides guidance to Claude Code when working with the `apps/cli` module.

## Overview

Cobra-based CLI tool (`qhub`) for interacting with the backend API from the command line.

## Structure

```
main.go              # Entry point, executes root command
cmd/
  root.go            # Root command with global flags (--api-url, --token, --output)
  org.go             # Organization commands (list, create, get, update, delete)
  project.go         # Project commands (list, create, get, update, delete)
  prompt.go          # Prompt commands (list, create, get, update, delete)
  version.go         # Version commands (list, create, get, promote, archive)
```

## Global Flags

- `--api-url` / `-u`: Backend API URL (default: `http://localhost:8080`)
- `--token` / `-t`: Bearer auth token
- `--output` / `-o`: Output format (`json`, `table`)

## Command Hierarchy

```
qhub org list|create|get|update|delete
qhub project list|create|get|update|delete
qhub prompt list|create|get|update|delete
qhub version list|create|get|promote|archive
```

## Key Patterns

- Each command file defines subcommands using Cobra's `AddCommand()`
- API communication via HTTP client with Bearer auth
- JSON and table output formats
- Version commands support file input for prompt content
- Promote/archive workflows for version lifecycle

## Commands

```bash
# Build CLI
make build-cli

# Run CLI
make run-cli

# Or directly
cd apps/cli && go run main.go org list --api-url http://localhost:8080 --token <token>
```
