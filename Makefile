SHELL := /bin/bash

# -----------------------------------------------------------------------------
# Project metadata
# -----------------------------------------------------------------------------
BIN_DIR        := bin
BIN_NAME       := svarog
DB_DSN         ?= $(shell grep -E '^DB_DSN=' .env 2>/dev/null | sed 's/^DB_DSN=//')
COMPOSE_FILE   := deploy/docker-compose.yml
MIGRATIONS_DIR := migrations

GO             ?= go

# Docker: native binary in PATH, or Docker Desktop on Windows (WSL2 without integration).
DOCKER_BIN     ?= $(shell command -v docker 2>/dev/null || true)
ifeq ($(DOCKER_BIN),)
  DOCKER_BIN   := /mnt/c/Program Files/Docker/Docker/resources/bin/docker.exe
endif
DOCKER_COMPOSE ?= "$(DOCKER_BIN)" compose

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-22s\033[0m %s\n", $$1, $$2}'

# -----------------------------------------------------------------------------
# Tooling
# -----------------------------------------------------------------------------
.PHONY: bootstrap
bootstrap: ## Install CLI tools (easyp, golang-migrate)
	$(GO) install github.com/easyp-tech/easyp/cmd/easyp@latest
	$(GO) install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# -----------------------------------------------------------------------------
# Build & run
# -----------------------------------------------------------------------------
.PHONY: build
build: ## Build the application binary into $(BIN_DIR)/$(BIN_NAME)
	@mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN_DIR)/$(BIN_NAME) ./cmd

.PHONY: run
run: ## Run the application locally
	$(GO) run ./cmd

.PHONY: run-mcp-stdio
run-mcp-stdio: ## Run MCP server over stdio (for Cursor)
	$(GO) run ./cmd/mcp-stdio

.PHONY: run-mcp-http
run-mcp-http: ## Run MCP server over streamable HTTP on :8000/mcp
	$(GO) run ./cmd/mcp-http

.PHONY: test
test: ## Run unit tests
	$(GO) test ./...

.PHONY: lint
lint: ## Run linters (golangci-lint + easyp)
	@command -v golangci-lint >/dev/null 2>&1 && golangci-lint run || echo "golangci-lint not installed, skipping"
	@command -v easyp >/dev/null 2>&1 && easyp lint || echo "easyp not installed, skipping"

# -----------------------------------------------------------------------------
# Proto
# -----------------------------------------------------------------------------
.PHONY: proto-gen
proto-gen: ## Generate Go/gRPC/gateway/openapi from .proto
	easyp generate

.PHONY: proto-lint
proto-lint: ## Lint .proto files
	easyp lint

# -----------------------------------------------------------------------------
# Migrations
# -----------------------------------------------------------------------------
.PHONY: migrate
migrate: ## Apply all pending migrations
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_DSN)" up

.PHONY: migrate-down
migrate-down: ## Roll back the most recent migration
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_DSN)" down 1

.PHONY: migrate-new
migrate-new: ## Create a new migration (usage: make migrate-new name=add_avatar)
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)

# -----------------------------------------------------------------------------
# Docker compose (Postgres + LGTM + OTel Collector)
# -----------------------------------------------------------------------------
.PHONY: up
up: ## Bring up full stack (infra + migrate + app + mcp-http)
	$(DOCKER_COMPOSE) -f $(COMPOSE_FILE) up -d --build

.PHONY: down
down: ## Tear down local infrastructure
	$(DOCKER_COMPOSE) -f $(COMPOSE_FILE) down

.PHONY: logs
logs: ## Follow docker-compose logs
	$(DOCKER_COMPOSE) -f $(COMPOSE_FILE) logs -f

.PHONY: ps
ps: ## Show docker-compose status
	$(DOCKER_COMPOSE) -f $(COMPOSE_FILE) ps
