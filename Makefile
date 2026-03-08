SHELL := /bin/bash

PHONY: help
help:
	cat ./Makefile

#############
# Variables #
#############
API_MOD = ./apps/api
PKGS_MOD = ./apps/pkgs
WEB_MOD = ./apps/web
DB_MOD = ./apps/pkgs/db
IAC_DIR = ./apps/iac
MODS = $(API_MOD) $(PKGS_MOD) $(WEB_MOD)
ATLAS_ENV ?= local
ATLAS_DIR = $(DB_MOD)

#################
# Dev Container #
#################
.PHONY: dev-cp run-api run-web run-migrate run-all init-firewall

dev-cp:
	cp .devcontainer/compose.override.yaml.example .devcontainer/compose.override.yaml

# Initialize firewall with whitelisted domains
firewall:
	@echo "Initializing firewall..."
	sudo .devcontainer/firewall.sh

# Run API server (port 8080)
run-api:
	cd ${API_MOD}/src && go run ./cmd

# Run Web server (port 3000)
# Requires: make templ-gen (run once after templ file changes)
run-web:
	cd ${WEB_MOD}/src && go run ./cmd

# Run database migrations (Atlas)
run-migrate:
	@echo "Running database migrations..."
	cd ${DB_MOD} && atlas migrate apply --env local

# Run all services (migrate -> api -> web in background)
run-all: run-migrate
	@echo "Starting API server..."
	cd ${API_MOD}/src && go run ./cmd &
	@echo "Starting Web server..."
	cd ${WEB_MOD}/src && go run ./cmd

###############
# Claude Code #
###############
.PHONY: claude

claude:
	claude --dangerously-skip-permissions

#############
# Container #
#############
.PHONY: cp check fmt fix vet lint cspell test tidy graph env

cp:
	cp compose.override.yaml.example compose.override.yaml

check: fmt vet lint cspell test

# Run tests (requires DB: docker compose up -d db)
test: atlas-apply
	for mod in $(MODS); do \
		(cd $$mod && go test -cover ./...); \
	done

tidy:
	for mod in $(MODS); do \
		(cd $$mod && go mod tidy); \
	done

fmt:
	for mod in $(MODS); do \
		(cd $$mod && go fmt ./...); \
	done

fix:
	for mod in $(MODS); do \
		(cd $$mod && go fix ./...); \
	done

vet:
	for mod in $(MODS); do \
		(cd $$mod && go vet ./...); \
	done

lint:
	for mod in $(MODS); do \
		(cd $$mod && golangci-lint run ./...); \
	done

cspell:
	misspell -error -locale US .

graph:
	for mod in $(MODS); do \
		(cd $$mod && go mod graph); \
	done

env:
	for mod in $(MODS); do \
		(cd $$mod && go env); \
	done

#########
# templ #
#########
.PHONY: templ-gen templ-watch templ-fmt

templ-gen:
	cd ${WEB_MOD}/src && templ generate

templ-watch:
	cd ${WEB_MOD}/src && templ generate --watch

templ-fmt:
	cd ${WEB_MOD}/src && templ fmt .

########
# wasm #
########
# NOTE: WASM targets are deprecated. Use templ for web frontend.
.PHONY: wasm wasm-clean

wasm:
	@echo "DEPRECATED: Use 'make templ-gen' instead"
	@echo "Building WebAssembly..."
	cd ${WEB_MOD} && GOOS=js GOARCH=wasm go build -o main.wasm .
	@echo "Done: ${WEB_MOD}/main.wasm"

wasm-clean:
	rm -f ${WEB_MOD}/main.wasm

########
# sqlc #
########
.PHONY: sqlc-gen sqlc-gen sqlc-compile sqlc-verify sqlc-help

sqlc-gen:
	cd ${DB_MOD} && sqlc generate

sqlc-compile:
	cd ${DB_MOD} && sqlc compile

sqlc-verify:
	cd ${DB_MOD} && sqlc verify

###################
# atlas migration #
###################
.PHONY: atlas-new atlas-diff atlas-apply atlas-status atlas-hash atlas-lint atlas-inspect atlas-clean atlas-validate

# Create a new migration file with a given name
# Usage: make atlas-new NAME=create_users_table
atlas-new:
	@if [ -z "$(NAME)" ]; then \
		echo "Error: NAME is required. Usage: make atlas-new NAME=create_users_table"; \
		exit 1; \
	fi
	cd ${ATLAS_DIR} && atlas migrate new --env ${ATLAS_ENV} $(NAME)

# Generate a migration by comparing schema files with the current database state
# This will create a new migration file based on the diff
# Usage: make atlas-diff NAME=add_users_table
atlas-diff:
	@echo "Generating migration from schema diff..."
	cd ${ATLAS_DIR} && atlas migrate diff --env ${ATLAS_ENV} $(NAME)

# Apply pending migrations to the database
# Usage: make atlas-apply [ENV=docker]
atlas-apply:
	@echo "Applying migrations to database (env: ${ATLAS_ENV})..."
	cd ${ATLAS_DIR} && atlas migrate apply --env ${ATLAS_ENV}

# Apply migrations with auto-approval (use with caution)
atlas-apply-auto:
	@echo "Applying migrations with auto-approval (env: ${ATLAS_ENV})..."
	cd ${ATLAS_DIR} && atlas migrate apply --env ${ATLAS_ENV} --auto-approve

# Show migration status (applied vs pending)
atlas-status:
	@echo "Migration status (env: ${ATLAS_ENV}):"
	cd ${ATLAS_DIR} && atlas migrate status --env ${ATLAS_ENV}

# Rehash migration directory (update atlas.sum)
# Run this after manually creating or editing migration files
atlas-hash:
	@echo "Rehashing migration directory..."
	cd ${ATLAS_DIR} && atlas migrate hash

# Lint migration files for issues
atlas-lint:
	@echo "Linting migrations..."
	cd ${ATLAS_DIR} && atlas migrate lint --env ${ATLAS_ENV}

# Validate migration directory integrity
atlas-validate:
	@echo "Validating migration directory..."
	cd ${ATLAS_DIR} && atlas migrate validate

# Inspect current database schema
atlas-inspect:
	@echo "Inspecting database schema (env: ${ATLAS_ENV})..."
	cd ${ATLAS_DIR} && atlas schema inspect --env ${ATLAS_ENV}

# Apply schema directly (declarative approach - bypasses migrations)
# WARNING: This can be destructive. Use for development only.
atlas-schema-apply:
	@echo "Applying schema directly (env: ${ATLAS_ENV})..."
	cd ${ATLAS_DIR} && atlas schema apply --env ${ATLAS_ENV} --auto-approve

# Clean/reset dev database (for development)
# This will drop and recreate the dev database
atlas-clean:
	@echo "Cleaning dev database..."
	@echo "WARNING: This will drop all data in the dev database."
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		cd ${ATLAS_DIR} && atlas schema clean --env ${ATLAS_ENV} --auto-approve; \
	fi

#############
# Git Hooks #
#############
.PHONY: hooks hooks-install hooks-uninstall

# Install git hooks
hooks-install:
	@echo "Installing git hooks..."
	@mkdir -p .githooks
	@printf '#!/bin/sh\necho "Running pre-commit hooks..."\nmake fmt && make vet\n' > .githooks/pre-commit
	@printf '#!/bin/sh\necho "Running pre-push hooks..."\nmake test\n' > .githooks/pre-push
	@chmod +x .githooks/pre-commit .githooks/pre-push
	@if [ -f .githooks/post-commit ]; then chmod +x .githooks/post-commit; fi
	@git config core.hooksPath .githooks
	@echo "Git hooks installed successfully!"

# Uninstall git hooks (preserves tracked files like post-commit)
hooks-uninstall:
	@echo "Uninstalling git hooks..."
	@git config --unset core.hooksPath || true
	@rm -f .githooks/pre-commit .githooks/pre-push
	@echo "Git hooks uninstalled (post-commit preserved)."

# Alias for hooks-install
hooks: hooks-install

#############
# Terraform #
#############
.PHONY: tf-fmt

tf-fmt:
	cd ${IAC_DIR} && terraform fmt -recursive
