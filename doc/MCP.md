# MCP-серверы svarog-core

Проект включает собственный MCP-сервер на [`mark3labs/mcp-go`](https://github.com/mark3labs/mcp-go) с общей логикой в `internal/mcp/`.

## Транспорты

| Бинарь | Транспорт | Назначение |
|--------|-----------|------------|
| `cmd/mcp-stdio` | STDIO | Локально в Cursor / VS Code (процесс `go run`) |
| `cmd/mcp-http` | Streamable HTTP `/mcp` | Сеть, Docker Compose (порт **8000**) |

## Tools (минимум 3)

| Tool | Источник данных |
|------|-----------------|
| `weather_forecast` | Внешний API [Open-Meteo](https://open-meteo.com/) (без ключа) |
| `list_users` | Postgres `users` (read-only) |
| `list_active_sessions` | Postgres `sessions` + join `users` |
| `register_user` | HTTP `POST /v1/auth/register` на svarog-core |

У каждого tool и аргумента — развёрнутый `description` (UX для агента).

## Resource

- URI: `svarog://service/info`
- JSON: версия, базовый URL API, список endpoints и tools.

## Prompt

- `investigate_user` — чеклист расследования аккаунта по email (связка resource + DB tools).

## Переменные окружения

См. `.env.example` (`MCP_HTTP_ADDR`, `SVAROG_HTTP_BASE`, те же `POSTGRES_*`, что у приложения).

## Подключение в Cursor

Файл [`.cursor/mcp.json`](../.cursor/mcp.json):

1. **svarog-stdio** — `go run ./cmd/mcp-stdio` (нужны Postgres и поднятый `make run` для `register_user`).
2. **svarog-http** — `http://localhost:8000/mcp` (после `docker compose up` или `make run-mcp-http`).
3. **context7** — удалённый MCP `https://mcp.context7.com/mcp` (без `npx`; важно для WSL, где Windows-`npx` падает на UNC-путях проекта).
4. **obsidian-shared** — filesystem MCP только на `AI-Shared/` в Obsidian vault (`/mnt/e/Obsidian/life/AI-Shared`). Основной vault **не** подключён.

Опционально: API-ключ в `.cursor/mcp.json`:

```json
"context7": {
  "url": "https://mcp.context7.com/mcp",
  "headers": { "CONTEXT7_API_KEY": "ваш-ключ" }
}
```

Если в **Settings → MCP** есть второй context7 от marketplace-плагина — отключите дубликат, оставьте только запись из `.cursor/mcp.json`.

Перезагрузите MCP в Cursor (Command Palette → *MCP: List Servers* / перезапуск окна).

## Сценарий проверки (два MCP + агент)

1. Поднять стек: `make up` (Postgres, LGTM, **migrate**, **app**, **mcp-http**).
2. Убедиться: `curl -s http://localhost:8080/v1/auth/me` → 401 без cookie.
3. В чате Cursor попросить агента:
   - через **context7** найти актуальную документацию по grpc-gateway или OTel;
   - через **svarog-stdio** вызвать `weather_forecast` для вашего города;
   - через **svarog-stdio** `register_user`, затем `list_users`;
   - прочитать resource `svarog://service/info`;
   - выполнить prompt `investigate_user` для созданного email.
4. Опционально подключить **svarog-http** вместо stdio и повторить вызов tool по HTTP.

## Локальный запуск без Docker

```bash
make up          # только инфра (если app не в compose)
make migrate
make run         # :8080 / :9090
make run-mcp-stdio   # второй терминал
# или
make run-mcp-http    # :8000/mcp
```
