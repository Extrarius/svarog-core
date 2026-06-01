# AGENTS.md

Instructions for AI coding agents working in the `svarog-core` repository.
Agent workspace layout (tasks, checklists, skills, …) lives in [`.agents/`](../.agents/) — see [`.agents/README.md`](../.agents/README.md).

This file lives under [`doc/`](./) so the **repository root contains only [`README.md`](../README.md)** for humans; agent docs are grouped here. The content still follows the [agents.md](https://agents.md/) convention.

## Project overview

- Go 1.25 modular monolith, module `github.com/Extrarius/svarog-core`.
- Provides user **registration**, **authentication** via opaque session cookies, and a **full LGTM observability stack** (Loki, Grafana, Tempo, Mimir) wired through OpenTelemetry.
- Transport is **gRPC + grpc-gateway**, contract-first. Source of truth lives in `api/proto/`.
- Storage is **PostgreSQL 17.7** accessed via `pgx/v5` (no ORM, no sqlc — hand-written SQL inside `internal/adapters/repo`).
- Migrations are managed with **golang-migrate**.

## Setup commands

```bash
make bootstrap        # install CLI tools (easyp, golang-migrate)
cp .env.example .env  # local environment
make up               # full stack: Postgres, LGTM, migrate, app, mcp-http
make migrate          # apply migrations (local CLI, outside compose migrate job)
make proto-gen        # generate Go/grpc/gateway/openapi stubs
make run              # run the application locally (without Docker app service)
make run-mcp-stdio    # MCP over stdio for Cursor
make run-mcp-http     # MCP streamable HTTP on :8000/mcp
```

MCP details: [`doc/MCP.md`](./MCP.md). Cursor config: [`.cursor/mcp.json`](../.cursor/mcp.json).

## Architecture rules (Clean Architecture)

The single most important rule:

> **`internal/app` MUST NOT import any non-stdlib package, except other `internal/*` packages that are themselves stdlib-only.**

- `internal/app/contracts.go` defines the **ports** (interfaces) that the use cases need: `UserRepo`, `SessionRepo`, `Hasher`, `Clock`, etc.
- `internal/app/handlers.go` contains the **use cases**: `Register`, `Login`, `Logout`, `Me`. They receive ports through a constructor and never reach for globals.
- `internal/adapters/repo` provides **pgx-based** implementations of `UserRepo` and `SessionRepo`. Bcrypt-based `Hasher` lives here too.
- `internal/auth` is a **stdlib-only** helper for session tokens (generate, parse cookie, SHA-256 hash). `internal/app` may import it freely.
- `internal/api/grpc` and `internal/api/gateway` are **transport adapters** that translate gRPC/HTTP calls into use case invocations.
- `internal/config`, `internal/logger`, `internal/observability` are infrastructure adapters wired up from `cmd/main.go`.
- `cmd/main.go` is the **composition root**: it constructs concrete adapters and injects them into use cases.
- `internal/mcp` implements the **MCP server** (tools, resource, prompt) shared by `cmd/mcp-stdio` and `cmd/mcp-http`. It may use `pgx` and HTTP clients but is not part of `internal/app`.

Dependency arrows always point inward, toward `internal/app`.

## Code style

- Go 1.25, formatted with `gofmt`.
- `golangci-lint run` must pass; see `.golangci.yml` (TODO: add).
- Package naming: short, lowercase, no underscores.
- Errors: wrap with `fmt.Errorf("...: %w", err)` and surface canonical gRPC codes at the transport boundary only.
- Structured logging via `log/slog` from `internal/logger`. Do not use the global `slog.Default()` in business code — accept a `*slog.Logger` parameter.
- No business logic in transport handlers — they parse requests, call use cases, map results.

## Proto / API conventions

- Proto packages live under `api/proto/<service>/v<N>/`. Current contract: `auth/v1`.
- Generated code goes to `api/gen/go/` and `api/gen/openapi/`. Regenerate with `make proto-gen`.
- Every PR that touches `.proto` files MUST run:
  - `easyp lint`
  - `easyp breaking --against main`
- HTTP routes are described via `google.api.http` annotations and surfaced through grpc-gateway.

## Database conventions

- Migrations are forward-and-backward: every `NNNN_*.up.sql` has a matching `.down.sql`.
- Use `make migrate` / `make migrate-down`.
- All SQL lives inside `internal/adapters/repo/*.go` files (raw `pgx` queries). No SQL strings outside the repo package.
- `internal/app` interacts with the database exclusively through the `UserRepo` / `SessionRepo` interfaces.

## Auth & secrets

- Cookies: `HttpOnly`, `SameSite=Lax`, `Secure` controlled by `SESSION_COOKIE_SECURE` (must be `true` in production).
- Session token format: 32 random bytes (`crypto/rand`), base64url-encoded; only `sha256(token)` is persisted in `sessions.token_hash`.
- Passwords are hashed with bcrypt inside the `Hasher` adapter (`internal/adapters/repo`).
- Never commit `.env`, credentials, or private keys. `.env.example` shows the schema.

## Testing

- Run all tests with `make test` (`go test ./...`).
- Prefer table-driven tests.
- Integration tests against Postgres are planned with `testcontainers-go`; out of scope for the initial skeleton.

## Pull request instructions

- Title format: `[svarog-core] <subject>`.
- Required checks before merge: `make lint`, `make test`. If `.proto` changed, also `easyp breaking --against main`.
- Migrations are append-only on `main`. Never mutate existing `*.up.sql` files post-merge.

## Where to look for task-specific instructions

The authoritative agent doc for this repo is this file (`doc/AGENTS.md`). Typical entry points:

- [`.agents/tasks/new-endpoint.md`](../.agents/tasks/new-endpoint.md) — add a new gRPC + HTTP endpoint
- [`.agents/tasks/new-migration.md`](../.agents/tasks/new-migration.md) — create and apply a new migration
- [`.agents/checklists/pr-review.md`](../.agents/checklists/pr-review.md) — PR review checklist
- [`.agents/checklists/observability-check.md`](../.agents/checklists/observability-check.md) — verify logs / metrics / traces reach Grafana
