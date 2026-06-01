# svarog-core

Каркас Go 1.25 монолита с регистрацией/авторизацией через session cookies и полным observability-стеком LGTM (Loki + Grafana + Tempo + Mimir) на базе OpenTelemetry.

## Стек

- **Язык**: Go 1.25, module `github.com/Extrarius/svarog-core`
- **Архитектура**: Clean Architecture со строгой границей — `internal/app` не импортирует внешние пакеты (только stdlib и `internal/*`).
- **Транспорт**: gRPC + grpc-gateway, контракт — `api/proto/auth/v1/auth.proto`.
- **Proto-tooling**: [`easyp`](https://github.com/easyp-tech/easyp).
- **БД**: Postgres 17.7, `pgx/v5`. Запросы вручную в `internal/adapters/repo` (без sqlc).
- **Миграции**: [`golang-migrate`](https://github.com/golang-migrate/migrate).
- **Аутентификация**: opaque session-токены в cookie `sid`, в БД храним SHA-256(token). Без Redis, без JWT.
- **Observability**: OpenTelemetry SDK → OTel Collector → Loki / Tempo / Mimir → Grafana.

## Быстрый старт

**WSL2 + Docker Desktop:** запустите Docker Desktop в Windows, затем `make up`. Если в WSL нет `make`, добавьте в PATH локальную копию: `export PATH="$PWD/tools/apt-make/usr/bin:$PATH"` (или `sudo apt install make`). Альтернатива без make: `./scripts/up.sh`.

```bash
# 1. Установить вспомогательные CLI (easyp, migrate)
make bootstrap

# 2. Подготовить окружение
cp .env.example .env

# 3. Поднять весь стек (Postgres, LGTM, migrate, app, mcp-http)
make up

# 4. (Локально без Docker для app) накатить миграции и запустить:
make migrate
make run

# 5. Сгенерировать proto-код (если меняли .proto)
make proto-gen
```

После `make up`:

- HTTP API: http://localhost:8080 (`/v1/auth/*`)
- gRPC: `:9090`
- MCP (HTTP): http://localhost:8000/mcp
- Grafana: http://localhost:3000

Подключение MCP в Cursor — [`doc/MCP.md`](doc/MCP.md), конфиг [`.cursor/mcp.json`](.cursor/mcp.json).

## Раскладка проекта

```
.
├── README.md                   # единственный .md в корне (этот файл)
├── doc/
│   ├── README.md               # оглавление doc/
│   ├── AGENTS.md               # инструкции для AI-агентов (EN)
│   ├── HOMEWORK_DIALOG.md      # выгрузка для домашки 1
│   ├── HOMEWORK_TWO.md         # домашка 2 (compose + MCP)
│   └── MCP.md                  # MCP-серверы и сценарий в Cursor
├── .agents/                    # skills, workflows, rules, tasks, notes, checklists (EN)
├── cmd/
│   ├── main.go                 # svarog-core API
│   ├── mcp-stdio/              # MCP over stdio
│   └── mcp-http/               # MCP over HTTP /mcp
├── internal/mcp/               # общая логика MCP
├── internal/
│   ├── app/                    # бизнес-логика (stdlib only)
│   ├── adapters/repo/          # pgx-реализации портов
│   ├── auth/                   # session tokens (stdlib only)
│   ├── api/{grpc,gateway}/     # транспортный слой
│   ├── config/                 # envconfig
│   ├── logger/                 # slog + OTel bridge
│   └── observability/          # OTel SDK
├── api/
│   ├── proto/                  # .proto contracts
│   └── gen/                    # generated code
├── migrations/                 # golang-migrate
└── deploy/                     # docker-compose + LGTM configs
```

## Полезные команды

| Команда | Что делает |
|---|---|
| `make up` / `make down` | Поднять / погасить весь стек (Docker) |
| `./scripts/up.sh` | То же, без GNU make |
| `make migrate` / `make migrate-down` | Накатить / откатить миграции |
| `make proto-gen` | Сгенерировать Go/gateway/openapi из `.proto` |
| `make lint` | `golangci-lint` + `easyp lint` |
| `make test` | Запустить тесты |
| `make build` | Собрать бинарь в `bin/svarog` |
| `make run` | Локальный запуск приложения |
| `make run-mcp-stdio` / `make run-mcp-http` | Локальный MCP |
