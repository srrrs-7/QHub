# CLAUDE.md - Web Frontend Module

This file provides guidance to Claude Code when working with the `apps/web` module.

## Overview

Server-side rendered Go frontend using templ templates and HTMX for interactivity. Material Design 3 (M3) styling. Runs on port 3000, communicates with the API on port 8080.

## Structure

```
src/
  cmd/main.go         # Entry point, API client init, graceful shutdown (port 3000)
  client/             # HTTP client for backend API communication
  handlers/           # HTTP handlers returning templ components
  routes/             # Chi router, full page vs HTMX partial routing
  templates/          # .templ files (type-safe, compile to Go code)
    *.templ            # Template source files
    *_templ.go         # Generated Go files (do not edit)
```

## Key Patterns

### templ + HTMX

- `.templ` files are type-safe Go templates compiled to `*_templ.go`
- HTMX attributes (`hx-get`, `hx-post`, `hx-target`, etc.) enable dynamic updates
- No client-side JavaScript — all interactivity through HTMX
- M3 design system for consistent UI components

### Page vs Partial Rendering

Handlers detect HTMX requests (`HX-Request` header) and return:
- **Full page**: Complete HTML document with layout
- **Partial**: HTML fragment for HTMX swap

### API Client

`client/` package wraps HTTP calls to the backend API. Initialized in `cmd/main.go` with the API base URL.

## Commands

```bash
# Run web server locally
make run-web                      # Port 3000 (run make templ-gen first)

# Template generation
make templ-gen                    # Generate Go code from .templ files
make templ-watch                  # Watch mode for development
make templ-fmt                    # Format .templ files

# Build Docker image
docker compose up -d web
```

## Development Workflow

1. Edit `.templ` files in `src/templates/`
2. Run `make templ-gen` (or `make templ-watch` for auto-rebuild)
3. Generated `*_templ.go` files are created alongside `.templ` files
4. Never edit `*_templ.go` files directly — they are regenerated

## Dockerfile

Multi-stage build: Go compilation with templ generation, then nginx for serving static assets.
