# Домашнее задание 4 — SDD через OpenSpec

## Цели

1. Установить [OpenSpec](https://github.com/Fission-AI/OpenSpec) и инициализировать в проекте.
2. Описать новую фичу через спеку (proposal → specs → design → tasks) **до** написания кода.
3. Реализовать по спеке (`/opsx:apply`), изменить спеку (добавить эндпоинт), повторить цикл.
4. Зафиксировать оба диалога агента в `.md`.

## Статус

| Часть | Статус |
|-------|--------|
| `npm install -g @fission-ai/openspec` + `openspec init` | ✅ |
| Change `manage-sessions` (proposal, specs, design, tasks) | ✅ |
| Apply v1: `ListSessions` + `RevokeSession` | ✅ |
| Итерация спеки: `RevokeAllOtherSessions` | ✅ |
| Apply v2 | ✅ |
| Диалоги в `doc/dialogs/` | ✅ |
| Документация | ✅ |

## Фича: управление сессиями

| RPC | HTTP | Описание |
|-----|------|----------|
| `ListSessions` | `GET /v1/sessions` | Активные сессии пользователя |
| `RevokeSession` | `DELETE /v1/sessions/{session_id}` | Отзыв одной сессии |
| `RevokeAllOtherSessions` | `POST /v1/sessions/revoke-others` | «Выйти везде», кроме текущей |

Все три RPC требуют аутентификации (session cookie `sid`).

## OpenSpec в репозитории

```
openspec/
├── config.yaml
└── changes/
    └── manage-sessions/
        ├── proposal.md
        ├── design.md
        ├── tasks.md
        ├── specs/session-management/spec.md
        └── .openspec.yaml
```

Cursor: слэш-команды в [`.cursor/commands/`](../.cursor/commands/) (`/opsx:propose`, `/opsx:apply`, …), skills в [`.cursor/skills/openspec-*`](../.cursor/skills/).

**Примечание:** в задании указана папка `.openspec/`; актуальная конвенция OpenSpec 1.5 — `openspec/`.

## Диалоги агента

| Итерация | Файл |
|----------|------|
| Спека + apply v1 | [`doc/dialogs/openspec-iteration-1.md`](./dialogs/openspec-iteration-1.md) |
| Изменение спеки + apply v2 | [`doc/dialogs/openspec-iteration-2.md`](./dialogs/openspec-iteration-2.md) |

## Установка (повтор)

```bash
export NVM_DIR="$HOME/.nvm" && . "$NVM_DIR/nvm.sh"
npm install -g @fission-ai/openspec@latest
cd ~/svarog-core
openspec init --tools cursor
openspec new change "my-feature"
# … артефакты по openspec instructions …
# реализация: /opsx:apply в Cursor
```

## Проверка API (при поднятом `make up`)

```bash
# register + login (сохранить cookie)
curl -c /tmp/sid.txt -X POST http://localhost:8080/v1/auth/register \
  -H 'Content-Type: application/json' -d '{"email":"s@test.com","password":"secret123"}'
curl -c /tmp/sid.txt -b /tmp/sid.txt -X POST http://localhost:8080/v1/auth/login \
  -H 'Content-Type: application/json' -d '{"email":"s@test.com","password":"secret123"}'

# list sessions
curl -b /tmp/sid.txt http://localhost:8080/v1/sessions

# revoke all other sessions
curl -b /tmp/sid.txt -X POST http://localhost:8080/v1/sessions/revoke-others \
  -H 'Content-Type: application/json' -d '{}'
```

## Ссылки

- [OpenSpec](https://github.com/Fission-AI/OpenSpec)
- [openspec.pro](https://openspec.pro)
- [Spec Kit](https://github.com/github/spec-kit) (альтернатива)
- [sipki-tech/sdd](https://github.com/sipki-tech/sdd) (альтернатива)
