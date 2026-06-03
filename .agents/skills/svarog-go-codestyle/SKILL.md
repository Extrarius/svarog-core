---
name: svarog-go-codestyle
description: Apply svarog-core Go coding standards — Clean Architecture boundary for internal/app, error wrapping, slog injection, gRPC status mapping at transport only, pgx SQL in repo adapters. Use when writing or reviewing Go in this repo, before opening a PR, or when aligning with project AGENTS.md and personal Go notes.
---

# svarog-go-codestyle

Coding rules for **github.com/Extrarius/svarog-core**. Authoritative project doc: [`doc/AGENTS.md`](../../../doc/AGENTS.md).

## Architecture (mandatory)

```
cmd/main.go          → composition root only
internal/app         → use cases; imports stdlib + stdlib-only internal/* (e.g. internal/auth)
internal/adapters    → pgx, bcrypt, clocks
internal/api/grpc|gateway → transport; maps app errors → gRPC codes
internal/mcp         → MCP tools (NOT part of internal/app)
```

> **`internal/app` MUST NOT import** pgx, grpc, otel, or other third-party packages.

## Errors

- Wrap: `fmt.Errorf("register: create user: %w", err)`.
- Sentinels in `internal/app`: `ErrEmailTaken`, `ErrInvalidCredentials`, `ErrSessionNotFound`, etc.
- **Only** `internal/api/grpc` converts to `status.Error(codes.*, ...)`.
- Never return password hashes or raw session tokens in API responses.

## Logging

- Accept `*slog.Logger` in constructors; use `log.InfoContext`, `WarnContext` with structured fields.
- Do not call `slog.SetDefault` inside `internal/app`.

## Database

- All SQL in `internal/adapters/repo/*.go`.
- Migrations: paired `*.up.sql` / `*.down.sql` under `migrations/`.
- When reading `inet` from Postgres into Go, use `COALESCE(host(ip), '')` in SELECT (see `sessions.go`).

## Proto / HTTP

- Edit `api/proto/**/*.proto`, run `make proto-gen`.
- HTTP routes via `google.api.http` annotations; session cookie bridged in `internal/api/gateway`.

## Personal Go notes (Obsidian, non-secret)

When the user mentions vault notes or Go practices beyond AGENTS.md:

1. Use MCP server **`obsidian-shared`** (scoped to `AI-Shared/` only).
2. Read `golang/practices.md` under that folder — e.g. `/mnt/e/Obsidian/life/AI-Shared/golang/practices.md`.
3. **Never** request access outside `AI-Shared`; the main vault is not mounted.

Also apply installed skill **`golang-code-style`** in `.agents/skills/golang-code-style/` for general Go clarity (line breaks, comments, control flow).

## PR checklist (quick)

- [ ] `go vet ./...` (no `go build` required by project convention unless user asks)
- [ ] No new imports in `internal/app` beyond allowed packages
- [ ] Proto changes regenerated if `.proto` touched
- [ ] [`doc/AGENTS.md`](../../../doc/AGENTS.md) conventions respected

## References

- [`doc/AGENTS.md`](../../../doc/AGENTS.md)
- [`.agents/checklists/pr-review.md`](../../checklists/pr-review.md)
- Obsidian AI-Shared: `AI-Shared/golang/practices.md` (via `obsidian-shared` MCP)
