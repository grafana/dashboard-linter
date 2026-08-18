# TODO

Задачи по ревью кода dashboard-linter (ветка `fix/linter-bugs-review`).
Порядок — по приоритету. Каждое исправление — отдельный коммит.

## Высокий приоритет

- [x] main.go: ошибка вместо паники при `lint` без аргументов и без `--stdin`
- [x] lint.go: безопасный type assert для поля `query` у шаблонной переменной (паника на `{"query": 123}`)
- [x] rule_panel_units.go: безопасный type assert для `unit` override (паника на нестроковом значении)
- [x] Починить `--fix`: `editable` остаётся `true` из-за `omitempty` + conflate, дубли строк в `rows`, мусорные `null`-поля, права `0600`
- [x] configuration.go: yaml-тег `targetIdx` (документированный ключ не парсится)
- [x] rules.go: `fixPanel`/`fixTarget` пишут в `dashboard.Panels[pi]` по индексу из плоского списка — сломано для вложенных и row-панелей
- [x] Тесты: main.go (нет аргументов), configuration.go (YAML-ключи), сценарии «мусорного» JSON

## Средний приоритет

- [x] results.go: `MaximumSeverity` учитывает `Fixed` (5) > `Warning` (3) — `--strict --fix` заведомо падает
- [x] Убрать `panic()` из конструкторов правил rate_interval и logql_auto

## Низкий приоритет

- [x] Убрать мёртвое поле `Calcs` с битым тегом `[]calcs`
- [x] Объединить дублирующиеся `expandVariables` и `expandLogQLVariables`
- [x] Makefile: `check-fmt` не должен модифицировать файлы перед проверкой
- [x] `TtyPrint` не должен выводить ANSI-эскейпы в не-TTY окружении

## История

### 2026-08-18, ветка fix/linter-bugs-review

Все задачи выполнены. Коммиты (12 исправлений + todo.md + gofmt):

1. `b9e03bc` — ошибка вместо паники при lint без аргументов (+ тесты CLI)
2. `ec5c1bd` — паника при нестроковом query у шаблонной переменной (+ тесты)
3. `6036fa4` — паника при нестроковом unit в override панели (+ тест)
4. `ab41135` — починен механизм `--fix` (editable, дубли rows, null-поля, права файла, conflate → собственный merge)
5. `640a074` — yaml-тег `targetIdx` в конфиге (+ тесты загрузки YAML)
6. `3bff002` — автоправки вложенных панелей через `PanelAt` (+ тесты)
7. `727f20e` — `MaximumSeverity` игнорирует Fixed (+ тест)
8. `0775a58` — убран panic() из конструкторов правил
9. `bb48f60` — удалено мёртвое поле Calcs с битым тегом
10. `17e94de` — объединены expandVariables/expandLogQLVariables
11. `9da0c81` — check-fmt без модификации файлов; `2e8587d` — gofmt
12. `0d5a878` — без ANSI-эскейпов в не-TTY

Проверено: `go build ./...`, `go vet ./...`, `go test ./...`, `make check-fmt` — без ошибок. `go mod tidy` убрал неиспользуемую зависимость conflate из go.mod/go.sum.
