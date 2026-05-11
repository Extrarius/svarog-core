# Домашнее задание — выгрузка диалога и ретро

Документ зафиксировал методологию: **что / где / ограничения / критерий**, условный **PLAN** (с уточняющими вопросами), **точный список путей**, проверку `**doc/AGENTS.md` + `.agents/`** и короткое **ретро** (что сработало / что нет).

Файл лежит в `[doc/HOMEWORK_DIALOG.md](./HOMEWORK_DIALOG.md)`; в **корне репозитория** остаётся только `[README.md](../README.md)` для людей.

---

## 1. Задача в формате что / где / ограничения / критерий

### Что

Инициализировать **собственный backend-template (скелет монолита)** под работу с ИИ-агентами и под выбранный технологический стек: регистрация и вход, **сессионная аутентификация** (cookie, без JWT), **PostgreSQL**, **gRPC + grpc-gateway**, **LGTM observability** (Loki, Grafana, Tempo, Mimir + OpenTelemetry), **Clean Architecture** с жёсткой границей `internal/app`.

### Где

Репозиторий `svarog-core` (module `github.com/Extrarius/svarog-core`):

- `**[doc/AGENTS.md](./AGENTS.md)`** — инструкции для агентов (английский).
- `**doc/`** — прочая проектная документация в Markdown (в т. ч. этот файл).
- `**.agents/**` — навыки, воркфлоу, правила, задачи, заметки, чеклисты (английский).
- Каркас кода: `cmd/`, `internal/{app,adapters/repo,auth,api/{grpc,gateway},config,logger,observability}`, `api/proto/`, `api/gen/`, `migrations/`, `deploy/`.

### Ограничения

- **Go 1.25** в `go.mod` (toolchain на машине может быть новее).
- **Без `go build` для проверки** — проверка через `go vet` / `go run` (по договорённости с пользователем).
- **Без полной бизнес-логики** на первом проходе: use-cases могут возвращать `ErrNotImplemented`.
- **Без sqlc** — SQL только в `internal/adapters/repo`.
- **Миграции** — `golang-migrate`, пары `up`/`down`.
- **Proto** — `easyp`, сгенерированный код в `api/gen/` (генерация — отдельный шаг `make proto-gen`).
- `**doc/AGENTS.md` и материалы в `.agents/`** — на **английском**; **корневой `README.md`** для людей — **русский**.
- В **корне репозитория** только `**README.md`**; остальные `.md` (кроме вложенных в `.agents/` и др. служебных каталогов) — в `**doc/`**.

### Критерий приёмки

1. В `**doc/**` есть **заполненный** `[AGENTS.md](./AGENTS.md)` по духу конвенции [agents.md](https://agents.md/).
2. Каталог `.agents/` содержит набор шаблонов.
3. Заложена **архитектура** (правила в `doc/AGENTS.md` отражены в структуре пакетов).
4. Проект **компилируемо проверяем** (`go vet ./...`) и **стартует** (`go run ./cmd`) с wiring config + logger + observability + серверов.
5. Есть **docker-compose** для Postgres + LGTM + OTel Collector в `deploy/`.
6. В **корне** только `**README.md`** среди пользовательских MD-документов проекта.

---

## 2. PLAN-режим — уточняющие вопросы (как должны были прозвучать)


| #   | Вопрос                  | Зафиксированный ответ                                                                        |
| --- | ----------------------- | -------------------------------------------------------------------------------------------- |
| 1   | Транспорт API?          | gRPC + grpc-gateway                                                                          |
| 2   | Модель аутентификации?  | Session cookies, сессии в Postgres, без Redis                                                |
| 3   | Observability?          | LGTM + OTel Collector; метрики в Mimir                                                       |
| 4   | Стиль архитектуры?      | Clean Architecture; `internal/app` без внешних импортов                                      |
| 5   | Расположение proto/gen? | `api/proto/`, `api/gen/`; транспорт в `internal/api/`                                        |
| 6   | Миграции?               | golang-migrate                                                                               |
| 7   | Работа с БД?            | pgx, ручной SQL в адаптерах                                                                  |
| 8   | Содержимое `.agents/`?  | `skills/`, `workflows/`, `rules/`, `tasks/`, `notes/`, `checklists/` (+ `.agents/README.md`) |
| 9   | Язык agent-доков?       | English для `doc/AGENTS.md` и `.agents/`                                                     |
| 10  | Корневой README?        | Russian; **единственный `.md` в корне**                                                      |
| 11  | `cmd/main`?             | Hello + полный wiring без полной бизнес-логики                                               |
| 12  | Module path?            | `github.com/Extrarius/svarog-core`                                                           |
| 13  | Proto tooling?          | easyp                                                                                        |


---

## 3. PLAN-проход — точный список файлов и папок

```
.
├── README.md                     # только этот md в корне (для людей)
├── Makefile
├── easyp.yaml
├── go.mod
├── go.sum
├── .gitignore
├── .dockerignore
├── .env.example
├── doc/
│   ├── README.md
│   ├── AGENTS.md
│   └── HOMEWORK_DIALOG.md       # этот файл
├── .agents/
│   ├── README.md
│   ├── skills/README.md
│   ├── workflows/README.md
│   ├── rules/README.md
│   ├── tasks/
│   │   ├── new-endpoint.md
│   │   └── new-migration.md
│   ├── notes/README.md
│   └── checklists/
│       ├── pr-review.md
│       └── observability-check.md
├── cmd/main.go
├── api/
│   ├── gen/.gitkeep
│   └── proto/auth/v1/auth.proto
├── internal/                     # см. дерево выше в репо при необходимости
├── migrations/
│   ├── 0001_init.up.sql
│   └── 0001_init.down.sql
└── deploy/
    ├── docker-compose.yml
    ├── otel-collector-config.yaml
    ├── grafana/provisioning/...
    ├── loki/loki.yaml
    ├── mimir/mimir.yaml
    └── tempo/tempo.yaml
```

---

## 4. Реализация (AGENT) — статус

Выполнено: `doc/AGENTS.md`, `doc/HOMEWORK_DIALOG.md`, каркас кода, `deploy/`, полный скелет `.agents/`.

---

## 5. Проверка критериев домашки


| Критерий                                     | Результат                             |
| -------------------------------------------- | ------------------------------------- |
| `doc/AGENTS.md` заполнен                     | **Да**                                |
| Корень: только `README.md` среди ключевых MD | **Да** (прочее в `doc/` и `.agents/`) |
| Подпапки `.agents/`                          | **Да**                                |
| Архитектура и стек                           | **Да**                                |


---

## 6. Ретро: что сработало, что нет

### Сработало

- [agents.md](https://agents.md/)-ориентированный документ агента в `**doc/`** + чеклисты в `**.agents/`**.
- Конвенция «только `**README.md**` в корне» — проще навигация для людей.
- Правило `**internal/app**` + OTel LGTM skeleton.

### Не идеально / ограничения

- Тулинги IDE (напр. автозагрузка правил Cursor), которые ожидают `**AGENTS.md` строго в корне**, нужно переуказать на `**doc/AGENTS.md`** в настройках проекта.
- `make proto-gen` и регистрация handlers в gateway/grpc — см. TODO в коде.

---

## 7. Минимальная проверка

```bash
go vet ./...
go run ./cmd
```

---

*Приложение к отчёту по домашней работе.*