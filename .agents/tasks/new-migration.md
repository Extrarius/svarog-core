# Task: add a new database migration

We use [`golang-migrate`](https://github.com/golang-migrate/migrate). Migration files live in `migrations/`.

## 1. Create the migration files

```bash
migrate create -ext sql -dir migrations -seq <descriptive_name>
```

This produces two files:

- `migrations/NNNN_<name>.up.sql`
- `migrations/NNNN_<name>.down.sql`

Conventions:

- File naming is snake_case and describes the change (e.g. `0002_add_user_avatar_url`).
- Always provide a working `down` migration. If a destructive change cannot be safely reversed, document the reason in the file header.

## 2. Write the SQL

- Forward migration goes into `*.up.sql`.
- Reverse migration goes into `*.down.sql`.
- Use `IF NOT EXISTS` / `IF EXISTS` where it makes the migration idempotent.
- Prefer `timestamptz` over `timestamp`. Prefer `text` over `varchar(N)` unless there is a strong reason for the bound.

## 3. Apply locally

```bash
make migrate            # apply all pending migrations
make migrate-down       # roll back the most recent one
```

## 4. Reflect changes in the repo layer

If you added or changed tables, update:

- `internal/adapters/repo/*.go` — pgx queries that touch the changed table(s).
- `internal/app/contracts.go` — repository interface signatures, if the data model now exposes new fields.

## 5. Test

```bash
make test
```

Add or update tests covering the new schema where reasonable.

## 6. PR checklist

- Migration files are append-only on `main`: never edit a merged `*.up.sql`.
- New migrations are reviewed for **locking implications** on large tables (`ALTER TABLE ... ADD COLUMN` is usually fast; adding non-`NOT NULL` defaults to existing rows can be expensive on Postgres < 11; we are on 17, so most operations are fast — still, mention any potential long lock in the PR description).
- The PR description explains *why* the migration is needed.
