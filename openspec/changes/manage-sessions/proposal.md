## Why

Users can log in and out but cannot see or revoke their active sessions from other devices. Session management is a common security feature that lets users audit logins and terminate suspicious sessions.

## What Changes

- Add `ListSessions` RPC — returns active sessions for the authenticated user with metadata (user agent, IP, timestamps) and a flag marking the current session.
- Add `RevokeSession` RPC — lets the authenticated user revoke one of their own sessions by ID.
- Add `RevokeAllOtherSessions` RPC — revokes all sessions except the current one ("logout everywhere else").
- Expose all via grpc-gateway HTTP routes (`GET /v1/sessions`, `DELETE /v1/sessions/{session_id}`, `POST /v1/sessions/revoke-others`).
- Register new methods as protected in the gRPC auth interceptor.

## Capabilities

### New Capabilities

- `session-management`: List and revoke active user sessions via authenticated API.

### Modified Capabilities

- (none)

## Impact

- `api/proto/auth/v1/auth.proto` — new RPCs and messages
- `internal/app` — new use cases and `SessionRepo` methods
- `internal/adapters/repo/sessions.go` — list and revoke-by-owner queries
- `internal/api/grpc` — new handlers, protected method registration
- Generated stubs under `api/gen/go/`
