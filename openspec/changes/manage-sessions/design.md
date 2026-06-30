## Context

`svarog-core` already persists sessions in PostgreSQL (`sessions` table) with `revoked_at`, `expires_at`, `user_agent`, and `ip`. Auth uses opaque cookie `sid`; the gRPC interceptor resolves identity via `app.Handlers.Me` and injects `MeOutput` (user ID, email, session ID) into context for protected RPCs.

Existing protected methods: `Logout`, `Me`. New methods follow the same pattern.

## Goals / Non-Goals

**Goals:**

- Add `ListSessions` and `RevokeSession` RPCs with grpc-gateway HTTP mapping.
- Extend `SessionRepo` with `ListActiveByUserID` and `RevokeOwned`.
- Mark current session via comparison with interceptor-injected `SessionID`.
- Clear cookie when user revokes their current session.
- Add `RevokeAllOtherSessions` for bulk revoke of all sessions except current.

**Non-Goals:**

- Admin cross-user session management.
- Pagination (v1 returns all active sessions; typical count is small).

## Decisions

1. **Reuse auth interceptor** — add `ListSessions` and `RevokeSession` to `protectedMethods`; identity comes from `IdentityFromContext`, no duplicate token parsing in handlers.

2. **Repo ownership check in SQL** — `RevokeOwned` updates only when `id = $1 AND user_id = $2 AND revoked_at IS NULL`; zero rows → `ErrSessionNotFound` at app layer.

3. **Session summary DTO** — app layer exposes `SessionSummary` without `TokenHash` (never leaves repo for list).

4. **HTTP routes** — `GET /v1/sessions`, `DELETE /v1/sessions/{session_id}`, `POST /v1/sessions/revoke-others`.

5. **Bulk revoke** — `RevokeAllExcept` repo method updates all rows where `user_id = $1 AND id != $2 AND revoked_at IS NULL`; returns affected row count.

## Risks / Trade-offs

- [Listing all sessions without pagination] → Acceptable for v1; users rarely have dozens of sessions.
- [Revoking current session requires cookie clear] → Reuse `MetaClearSession` metadata same as Logout.
