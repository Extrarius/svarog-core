# Rules (repo invariants)

Use this folder for **extra** constraints that are too granular for [`doc/AGENTS.md`](../doc/AGENTS.md) or that you want to version separately.

Hard rules already live in [`doc/AGENTS.md`](../doc/AGENTS.md) (e.g. `internal/app` import policy). Add files such as:

- `git-commits.md` — no Cursor co-author trailers; no screenshots in git
- `proto-naming.md` — package and RPC naming beyond the lint config
- `migration-safety.md` — table-lock expectations for this service

If a rule duplicates `doc/AGENTS.md`, prefer editing `doc/AGENTS.md` instead.
