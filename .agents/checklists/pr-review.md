# Task: review a pull request

Use this checklist when reviewing — or auto-reviewing — a PR in `svarog-core`.

## 1. Architectural rules

- [ ] `internal/app` contains **no** imports outside the stdlib and other stdlib-only `internal/*` packages.
- [ ] No SQL strings leak outside `internal/adapters/repo`.
- [ ] Transport (`internal/api/grpc`, `internal/api/gateway`) does not contain business logic.
- [ ] New external dependencies (in `go.mod`) are justified in the PR description.

## 2. Proto / API

- [ ] If `.proto` changed, `easyp lint` and `easyp breaking --against main` pass.
- [ ] HTTP routes for new methods have `google.api.http` annotations.
- [ ] Generated code (`api/gen/`) is not hand-edited.

## 3. Database

- [ ] Each `*.up.sql` migration has a matching `*.down.sql`.
- [ ] Migration is numbered sequentially after the latest on `main`.
- [ ] PR description calls out any potentially long-running statements (`ALTER TABLE`, large index builds).

## 4. Authentication & security

- [ ] No secrets in code or fixtures.
- [ ] Cookie attributes are correct (`HttpOnly`, `SameSite=Lax`, `Secure` driven by config).
- [ ] Session tokens are never logged. Only `token_hash` is persisted.
- [ ] Passwords are hashed with bcrypt at the adapter boundary.

## 5. Observability

- [ ] New code paths produce spans where they cross a boundary (DB call, outgoing HTTP/gRPC).
- [ ] Logs use the injected `*slog.Logger` and include relevant attributes (no PII).
- [ ] No `fmt.Println` outside `cmd/main.go`.

## 6. Tests & docs

- [ ] New behavior has tests (table-driven where possible).
- [ ] `make lint` and `make test` pass.
- [ ] User-visible changes are reflected in root `README.md` or [`doc/AGENTS.md`](../../doc/AGENTS.md).

## 7. PR hygiene

- [ ] Title follows `[svarog-core] <subject>`.
- [ ] Commits are scoped and readable; squash if necessary before merge.
- [ ] Description explains *why*, not just *what*.
