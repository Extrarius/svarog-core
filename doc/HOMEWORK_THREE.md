# Домашнее задание 3 — skills.sh и Agent Skills

## Цели

1. Установить экосистему [skills.sh](https://skills.sh/) (CLI `npx skills`).
2. Установить готовые скиллы (дизайн + Obsidian + Go).
3. Создать собственные скиллы для инфраструктуры и код-стайла svarog-core.
4. Безопасно подключить Obsidian (только папка `AI-Shared`).

## Статус

| Часть | Статус |
|-------|--------|
| nvm + Node LTS в WSL (нативный npx) | ✅ |
| `npx skills` (find / add / list) | ✅ |
| Готовые скиллы в `.agents/skills/` | ✅ |
| Obsidian `AI-Shared` + MCP `obsidian-shared` | ✅ |
| `svarog-infra-up`, `svarog-go-codestyle`, `svarog-mcp-tool` | ✅ |
| Документация | ✅ |

## Установленные готовые скиллы (skills.sh)

| Скилл | Источник | Назначение |
|-------|----------|------------|
| `web-design-guidelines` | `vercel-labs/agent-skills` | UI / design review |
| `obsidian-markdown` | `kepano/obsidian-skills` | Obsidian markdown, wikilinks |
| `golang-code-style` | `samber/cc-skills-golang` | Общий Go code style |

Установка (повтор):

```bash
export NVM_DIR="$HOME/.nvm" && . "$NVM_DIR/nvm.sh"
cd ~/svarog-core
npx skills add vercel-labs/agent-skills -a cursor -s web-design-guidelines --copy -y
npx skills add kepano/obsidian-skills -a cursor -s obsidian-markdown --copy -y
npx skills add samber/cc-skills-golang -a cursor -s golang-code-style --copy -y
```

## Собственные скиллы

| Скилл | Файл |
|-------|------|
| `svarog-infra-up` | [`.agents/skills/svarog-infra-up/SKILL.md`](../.agents/skills/svarog-infra-up/SKILL.md) |
| `svarog-go-codestyle` | [`.agents/skills/svarog-go-codestyle/SKILL.md`](../.agents/skills/svarog-go-codestyle/SKILL.md) |
| `svarog-mcp-tool` | [`.agents/skills/svarog-mcp-tool/SKILL.md`](../.agents/skills/svarog-mcp-tool/SKILL.md) |

## Obsidian (безопасность)

- Vault: `E:\Obsidian\life` (в WSL: `/mnt/e/Obsidian/life`).
- Для агентов доступна **только** `AI-Shared/` (создана с `golang/practices.md`).
- MCP: `obsidian-shared` в [`.cursor/mcp.json`](../.cursor/mcp.json) → `@modelcontextprotocol/server-filesystem` с одним аргументом-путём.
- **Не** монтировать весь vault.

## Проверка в Cursor

1. **Settings → MCP** — включить `obsidian-shared` (после Reload Window).
2. **Agent mode** — в списке skills должны быть все 6 project skills (`npx skills list`).
3. Запросы:
   - «Используй skill svarog-infra-up и проверь стек»
   - «По svarog-go-codestyle: добавь endpoint …»
   - «Прочитай golang/practices.md через obsidian-shared»

## WSL: почему nvm

Windows `npx` (`/mnt/c/nvm4w/...`) ломается на UNC-путях WSL. Linux Node через **nvm** (`~/.nvm`) — для `npx skills` и MCP `obsidian-shared`.

## Ретро

- Cursor читает project skills из `.agents/skills/` (совпадает с `npx skills -a cursor`).
- `--copy` вместо symlink — удобно для git.
- `skills-lock.json` фиксирует версии установленных skills.sh пакетов.
