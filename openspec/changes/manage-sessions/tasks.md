## 1. API Contract

- [x] 1.1 Add `ListSessions` and `RevokeSession` RPCs with messages to `api/proto/auth/v1/auth.proto`
- [x] 1.2 Add grpc-gateway HTTP annotations (`GET /v1/sessions`, `DELETE /v1/sessions/{session_id}`)
- [x] 1.3 Regenerate stubs with `easyp generate`

## 2. Domain Layer

- [x] 2.1 Add `SessionSummary` type and extend `SessionRepo` with `ListActiveByUserID` and `RevokeOwned`
- [x] 2.2 Implement `ListSessions` and `RevokeSession` use cases in `internal/app/handlers.go`
- [x] 2.3 Add `ErrSessionNotOwned` or reuse `ErrSessionNotFound` for cross-user revoke attempts

## 3. Persistence

- [x] 3.1 Implement `ListActiveByUserID` in `internal/adapters/repo/sessions.go`
- [x] 3.2 Implement `RevokeOwned` in `internal/adapters/repo/sessions.go`

## 4. Transport

- [x] 4.1 Register `ListSessions` and `RevokeSession` in `protectedMethods` (interceptor)
- [x] 4.2 Implement gRPC handlers in `internal/api/grpc/auth_service.go`
- [x] 4.3 Clear session cookie via `MetaClearSession` when revoking current session

## 5. Verification

- [x] 5.1 Run `go vet ./...`
- [x] 5.2 Run `go run ./cmd` (smoke, no build)

## 6. Iteration 2 — RevokeAllOtherSessions

- [x] 6.1 Add `RevokeAllOtherSessions` RPC to `auth.proto` with `POST /v1/sessions/revoke-others`
- [x] 6.2 Regenerate stubs with `easyp generate`
- [x] 6.3 Add `RevokeAllExcept` to `SessionRepo` and implement in repo
- [x] 6.4 Implement `RevokeAllOtherSessions` use case in handlers
- [x] 6.5 Register protected method and implement gRPC handler
- [x] 6.6 Run `go vet ./...`
