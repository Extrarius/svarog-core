# Skills

Reusable **Agent Skills** for Cursor and other agents ([skills.sh](https://skills.sh/), [agents.md](https://agents.md/) format).

Each skill is a folder with `SKILL.md` (YAML frontmatter + instructions).

## Installed from skills.sh (`npx skills add`)

| Skill | Source |
|-------|--------|
| [web-design-guidelines](./web-design-guidelines/) | `vercel-labs/agent-skills` |
| [obsidian-markdown](./obsidian-markdown/) | `kepano/obsidian-skills` |
| [golang-code-style](./golang-code-style/) | `samber/cc-skills-golang` |

Refresh / add more:

```bash
export NVM_DIR="$HOME/.nvm" && . "$NVM_DIR/nvm.sh"
npx skills find <query>
npx skills add <owner/repo> -a cursor -s <skill> --copy -y
npx skills list
```

## Project-specific (authored in-repo)

| Skill | When to use |
|-------|-------------|
| [svarog-infra-up](./svarog-infra-up/) | `make up`, Docker WSL, auth/MCP smoke |
| [svarog-go-codestyle](./svarog-go-codestyle/) | Go changes; links `doc/AGENTS.md` + Obsidian `AI-Shared` |
| [svarog-mcp-tool](./svarog-mcp-tool/) | Add a new tool to `internal/mcp` |

## Obsidian notes (optional)

Personal Go practices: vault folder `AI-Shared/golang/` (see MCP `obsidian-shared` in `.cursor/mcp.json`). **Do not** put secrets there.

Homework write-up: [`doc/HOMEWORK_THREE.md`](../../doc/HOMEWORK_THREE.md).
