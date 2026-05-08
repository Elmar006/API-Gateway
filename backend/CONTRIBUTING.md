# Contributing

Спасибо за интерес к проекту! Этот документ описывает, как работать над кодом.

## Окружение разработчика

- Go 1.25+
- Docker (для интеграционных тестов через testcontainers и для локальной PostgreSQL)
- `golangci-lint` (`go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`)
- `migrate` CLI (опционально): `go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest`

Подготовка:

```bash
git clone <repo>
cd backend
cp .env.example .env   # при необходимости отредактировать
make tidy
```

## Как сделать изменение

1. Создайте ветку от `main`: имя в формате `feat/<short>` или `fix/<short>`.
2. Внесите изменения в код. Старайтесь не выходить за рамки текущей задачи.
3. Запустите перед коммитом:

   ```bash
   make lint
   make test
   ```

4. Если затрагиваете схему БД — добавьте пару миграций `NNNNNN_<name>.up.sql` / `<name>.down.sql` в `migrations/` (не редактируйте уже выпущенные миграции).
5. Пишите коммиты по [Conventional Commits](https://www.conventionalcommits.org/) (например, `feat(router): support {param} segments`).
6. Откройте Pull Request, опишите мотивацию, тестирование, риски.

## Стиль кода

- Effective Go + `gofmt`/`goimports`. Линтеры включены в CI.
- Никаких naked-return; ошибки обрабатываются явно (`if err != nil { ... }`).
- `context.Context` всегда передаётся первым аргументом, не сохраняется в struct.
- Интерфейсы определяем рядом с потребителем (consumer-side interface).
- Не использовать `any`/`reflect`/`getattr-style` доступ ради удобства.

## Тестирование

- Юнит-тесты: табличный стиль, без выхода в сеть/диск/БД (моки).
- Интеграционные тесты: build-tag `integration`, поднимают postgres через testcontainers.
- CI запускает линтеры, юнит-тесты и интеграционные тесты на сервисном postgres-контейнере.

```bash
make test                # unit
make test-integration    # integration (требует Docker)
make cover               # HTML coverage
```

## Безопасность

- Никаких секретов в коммитах. Используйте `.env` (он в `.gitignore`).
- Все SQL-запросы — параметризованные.
- Пароли — bcrypt с cost ≥ 12.
- При обнаружении уязвимости — пишите maintainers'ам приватно, не открывайте public issue.

## Лицензия и авторские права

Внося изменения, вы соглашаетесь, что они распространяются на условиях лицензии проекта.
