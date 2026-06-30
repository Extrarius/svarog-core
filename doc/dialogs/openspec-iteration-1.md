# OpenSpec — итерация 1: спека и apply (ListSessions + RevokeSession)

Диалог агента при создании фичи `manage-sessions` через SDD-цикл OpenSpec.

## Контекст

- Ветка: `feature/home-four` (от `feature/home-three`)
- OpenSpec CLI: `@fission-ai/openspec@1.5.0`, `openspec init --tools cursor`
- Change: `openspec/changes/manage-sessions/`

## Диалог (сжатая выгрузка)

**Пользователь:** Домашка 4 — SDD через OpenSpec. Установить OpenSpec, создать фичу через `/opsx:new` + `/opsx:ff`, затем `/opsx:apply`, итерация спеки, зафиксировать диалоги.

**Агент (план):** Выбрана фича **управление сессиями** — `ListSessions` и `RevokeSession`. OpenSpec v1.5 использует профиль `core` и команды `/opsx:propose` / `/opsx:apply` (вместо устаревших `/opsx:new` + `/opsx:ff` из задания).

**Агент (инициализация):**

```bash
git checkout -b feature/home-four
npm install -g @fission-ai/openspec@latest
openspec init --tools cursor --force
openspec new change "manage-sessions"
```

**Агент (спека v1):** Сгенерированы артефакты:

| Файл | Содержание |
|------|------------|
| `proposal.md` | Зачем: пользователь не видит/не отзывает сессии с других устройств |
| `specs/session-management/spec.md` | Требования: ListSessions, RevokeSession + сценарии |
| `design.md` | Interceptor, SessionRepo, HTTP routes |
| `tasks.md` | Чек-лист из 5 групп задач |

**Агент (apply v1):** Реализация строго по `tasks.md`:

1. `auth.proto` — RPC `ListSessions`, `RevokeSession`, сообщения `SessionInfo`
2. `easyp generate` — регенерация stubs
3. `internal/app` — `SessionSummary`, `ListSessions`, `RevokeSession`
4. `internal/adapters/repo/sessions.go` — `ListActiveByUserID`, `RevokeOwned`
5. `internal/api/grpc` — protected methods + handlers, cookie clear при revoke current
6. `go vet ./...` — OK

## Результат итерации 1

- `GET /v1/sessions` — список активных сессий с `is_current`
- `DELETE /v1/sessions/{session_id}` — отзыв своей сессии
- Все задачи в `tasks.md` (группы 1–5) отмечены `[x]`
