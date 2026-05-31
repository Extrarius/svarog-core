-- 0001_init.up.sql
-- Initial schema: users + sessions.
-- RBAC tables are intentionally out of scope for the initial skeleton.

CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "citext";

CREATE TABLE IF NOT EXISTS users (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email         citext UNIQUE NOT NULL,
    password_hash text NOT NULL,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS sessions (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash    bytea NOT NULL,
    expires_at    timestamptz NOT NULL,
    created_at    timestamptz NOT NULL DEFAULT now(),
    last_seen_at  timestamptz NOT NULL DEFAULT now(),
    revoked_at    timestamptz,
    user_agent    text NOT NULL DEFAULT '',
    ip            inet
);

CREATE UNIQUE INDEX IF NOT EXISTS sessions_token_hash_uniq ON sessions(token_hash);
CREATE INDEX IF NOT EXISTS sessions_user_active_idx ON sessions(user_id) WHERE revoked_at IS NULL;
