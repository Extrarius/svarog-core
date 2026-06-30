# OpenSpec — итерация 2: изменение спеки и apply (RevokeAllOtherSessions)

Диалог агента при итерации change `manage-sessions` — добавление эндпоинта «выйти везде, кроме текущей сессии».

## Контекст

- Change: `openspec/changes/manage-sessions/` (тот же change, обновлённые артефакты)
- Цель итерации: показать SDD-цикл «изменил спеку → apply снова»

## Диалог (сжатая выгрузка)

**Агент (изменение спеки):** Обновлены артефакты без ручного кода:

- `proposal.md` — добавлен `RevokeAllOtherSessions`
- `specs/session-management/spec.md` — новое требование «Revoke all other sessions» + 3 сценария
- `design.md` — решение `RevokeAllExcept` в repo, route `POST /v1/sessions/revoke-others`
- `tasks.md` — группа 6 (6 задач)

**Агент (apply v2):** Реализация по обновлённой спеке:

1. `auth.proto` — RPC `RevokeAllOtherSessions`, `RevokeAllOtherSessionsResponse.revoked_count`
2. `easyp generate`
3. `SessionRepo.RevokeAllExcept` — SQL `UPDATE ... WHERE user_id AND id != except`
4. `Handlers.RevokeAllOtherSessions` — use case с логированием `revoked_count`
5. `auth_service.go` + `interceptor.go` — protected method + handler
6. `go vet ./...` — OK

## Результат итерации 2

- `POST /v1/sessions/revoke-others` — отзывает все сессии пользователя, кроме текущей
- Ответ: `{ "revoked_count": N }`
- Задачи группы 6 в `tasks.md` отмечены `[x]`

## Заметка про OpenSpec v1.5

В задании упоминаются `/opsx:new` и `/opsx:ff` (расширенный профиль). В установленной версии 1.5 доступен только профиль `core`; эквивалентный поток:

| Задание | OpenSpec 1.5 |
|---------|--------------|
| `/opsx:new` + `/opsx:ff` | `openspec new change` + ручная генерация артефактов по `openspec instructions` (или `/opsx:propose`) |
| `/opsx:apply` | `/opsx:apply` (skill `openspec-apply-change`) |

Каталог спеки: `openspec/` (не `.openspec/` из формулировки задания).
