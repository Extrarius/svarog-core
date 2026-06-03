# `.agents/` — agent workspace (skeleton)

This directory complements **[`doc/AGENTS.md`](../doc/AGENTS.md)**. The repository root keeps only **`README.md`** for humans. Keep agent-facing prose here **in English** unless a course explicitly requires otherwise.

Base layout (mirrors the course template):

| Subfolder | Purpose |
|-----------|---------|
| [`skills/`](skills/) | Agent skills ([skills.sh](https://skills.sh/) + project playbooks). See [`skills/README.md`](skills/README.md). |
| [`workflows/`](workflows/) | Multi-step processes that chain several files or commands. |
| [`rules/`](rules/) | Extra constraints and invariants **for this repo** (do not duplicate [`doc/AGENTS.md`](../doc/AGENTS.md) wholesale — link or extract deltas). |
| [`tasks/`](tasks/) | Task formulations: “do X with steps Y–Z”. |
| [`notes/`](notes/) | Session notes, decisions, scratch context (gitignored patterns optional). |
| [`checklists/`](checklists/) | Repeatable checklists (reviews, releases, observability smoke). |

Start here for common work:

- [`tasks/new-endpoint.md`](tasks/new-endpoint.md)
- [`tasks/new-migration.md`](tasks/new-migration.md)
- [`checklists/pr-review.md`](checklists/pr-review.md)
- [`checklists/observability-check.md`](checklists/observability-check.md)
