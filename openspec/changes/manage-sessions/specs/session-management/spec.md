## ADDED Requirements

### Requirement: List active sessions

The system SHALL return all active (non-revoked, non-expired) sessions belonging to the authenticated user. Each session entry MUST include: `id`, `user_agent`, `ip`, `created_at`, `last_seen_at`, and `is_current` (true when the session matches the caller's session cookie).

#### Scenario: Authenticated user lists sessions

- **WHEN** an authenticated user calls `ListSessions`
- **THEN** the system returns all active sessions for that user ordered by `last_seen_at` descending
- **AND** exactly one session has `is_current` set to true (the caller's session)

#### Scenario: Unauthenticated list attempt

- **WHEN** a request without a valid session calls `ListSessions`
- **THEN** the system responds with HTTP 401 / gRPC `Unauthenticated`

### Requirement: Revoke own session

The system SHALL allow an authenticated user to revoke a specific session by ID. The session MUST belong to the authenticated user. Revoking an already-revoked or unknown session owned by another user MUST return `NotFound`.

#### Scenario: Revoke another device session

- **WHEN** an authenticated user calls `RevokeSession` with a session ID that belongs to them and is active
- **THEN** the system marks that session as revoked
- **AND** returns success

#### Scenario: Revoke non-owned session

- **WHEN** an authenticated user calls `RevokeSession` with a session ID that does not belong to them
- **THEN** the system responds with HTTP 404 / gRPC `NotFound`

#### Scenario: Revoke current session

- **WHEN** an authenticated user revokes their own current session
- **THEN** the system marks the session as revoked
- **AND** the gateway clears the session cookie (same behaviour as Logout)

### Requirement: Revoke all other sessions

The system SHALL revoke all active sessions belonging to the authenticated user except the current session. The current session MUST remain valid.

#### Scenario: Logout everywhere else

- **WHEN** an authenticated user calls `RevokeAllOtherSessions`
- **THEN** all other active sessions for that user are revoked
- **AND** the current session remains active
- **AND** the response includes the count of revoked sessions

#### Scenario: No other sessions

- **WHEN** an authenticated user calls `RevokeAllOtherSessions` and has no other active sessions
- **THEN** the system returns success with revoked count zero

#### Scenario: Unauthenticated revoke-all attempt

- **WHEN** a request without a valid session calls `RevokeAllOtherSessions`
- **THEN** the system responds with HTTP 401 / gRPC `Unauthenticated`
