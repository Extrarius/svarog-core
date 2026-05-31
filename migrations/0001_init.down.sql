-- 0001_init.down.sql
-- Reverse the initial schema.

DROP INDEX IF EXISTS sessions_user_active_idx;
DROP INDEX IF EXISTS sessions_token_hash_uniq;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS users;

-- The citext and pgcrypto extensions are left in place: they are cheap, may be
-- in use by other migrations, and are not safe to drop blindly.
