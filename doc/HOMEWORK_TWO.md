# Домашнее задание 2 — compose-up и MCP

## Цели

1. **Docker Compose up** — Postgres, миграции one-shot, приложение svarog-core (auth), observability, MCP HTTP.
2. **Собственный MCP** — stdio + streamable HTTP, tools / resource / prompt.
3. **IDE** — `.cursor/mcp.json` с svarog + Context7.

## Статус реализации

| Часть | Статус |
|-------|--------|
| Auth end-to-end (Register/Login/Logout/Me) | ✅ |
| `api/gen` закоммичен, gRPC + gateway + cookies | ✅ |
| `Dockerfile` (app, mcp-stdio, mcp-http) | ✅ |
| `deploy/docker-compose.yml` migrate + app + mcp-http | ✅ |
| `internal/mcp` + `cmd/mcp-*` | ✅ |
| `.cursor/mcp.json` | ✅ |
| `doc/MCP.md`, обновления README / AGENTS | ✅ |

## Проверка

```bash
make bootstrap
cp .env.example .env
cd deploy && docker compose up -d   # или make up из корня
# app: http://localhost:8080
# mcp: http://localhost:8000/mcp
curl -X POST http://localhost:8080/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"demo@example.com","password":"secret123"}'
```

Cursor: включить MCP servers, сценарий из `doc/MCP.md`.

## Ретро (кратко)

- Proto-gen потребовал `M`-маппинги googleapis в `easyp.yaml`.
- Session auth: cookie ↔ gRPC metadata ↔ interceptor для `Me`/`Logout`.
- MCP: одна `internal/mcp.Build`, два транспорта.
