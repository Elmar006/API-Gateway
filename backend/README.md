# API Gateway (Go 1.25)

Высокопроизводительный API Gateway на Go: динамическая маршрутизация из БД, JWT-аутентификация админки, rate-limiting по IP, WebSocket-прокси, асинхронное батчевое логирование, Prometheus-метрики, health-checks и graceful shutdown.

Стек: Go 1.25, chi, pgx/v5 (pgxpool), bytedance/sonic, gofrs/uuid, logrus, golang-migrate, prometheus/client_golang, JWT (golang-jwt).

## Содержание

- [Архитектура](#архитектура)
- [Быстрый старт](#быстрый-старт)
- [Переменные окружения](#переменные-окружения)
- [Сборка бинарника](#сборка-бинарника)
- [Запуск тестов](#запуск-тестов)
- [Миграции БД](#миграции-бд)
- [HTTP API](#http-api)
- [Метрики и health-checks](#метрики-и-health-checks)
- [TLS](#tls)
- [WebSocket](#websocket)
- [Структура проекта](#структура-проекта)

## Архитектура

Чистая архитектура (Clean Architecture):

- `internal/domain` — сущности (`entity`) и интерфейсы репозиториев (`repository`).
- `internal/application` — сценарии использования (use cases).
- `internal/infrastructure` — реализации репозиториев (PostgreSQL), JWT, rate limiter, reverse proxy, конфигурация, логгер, метрики, миграции, trie-router.
- `internal/presentation` — HTTP-обработчики и middleware (публичный прокси и admin API).
- `cmd/main.go` — точка входа.

Ключевые решения:

- Два независимых HTTP-сервера: публичный (`PORT`, по умолчанию `:8080`) и админский (`ADMIN_PORT`, по умолчанию `:9090`).
- Динамические маршруты хранятся в БД и матчатся через собственный trie-роутер с поддержкой `{param}` и `*`.
- Совпавшие path-параметры пробрасываются в апстрим в виде HTTP-заголовков `X-Path-Param-<name>`.
- Логи запросов собираются в кольцевой буфер и сбрасываются батчами через `pgx.CopyFrom` (см. [`internal/infrastructure/log/async.go`](internal/infrastructure/log/async.go)).
- Rate limiter — token bucket, шардированный (16 шардов) по FNV32a от IP, c очисткой бакетов по TTL.
- Graceful shutdown обоих серверов параллельно через `errgroup`, далее закрываются: async-логгер → rate limiter cleanup → пул БД.
- Миграции применяются на старте через `golang-migrate` с `embed.FS`.

## Быстрый старт

1. Поднять PostgreSQL (например, локально: `docker run -d --name pg -p 5432:5432 -e POSTGRES_PASSWORD=postgres postgres:16-alpine`).
2. Создать `.env` (см. [`.env.example`](.env.example)) и выставить минимум: `JWT_SECRET`, `DB_PASSWORD`, `ADMIN_USERNAME`, `ADMIN_PASSWORD`.
3. Запустить шлюз — миграции применятся автоматически:

   ```bash
   make run
   ```

4. Получить JWT-токен админа:

   ```bash
   curl -s -X POST http://localhost:9090/login \
     -H 'Content-Type: application/json' \
     -d '{"username":"admin","password":"admin-pass"}'
   ```

5. Создать маршрут и проверить публичный прокси:

   ```bash
   TOKEN=...  # значение из шага 4
   curl -s -X POST http://localhost:9090/routes \
     -H "Authorization: Bearer $TOKEN" \
     -H 'Content-Type: application/json' \
     -d '{"method":"GET","path_pattern":"/api/users/{id}","target_url":"https://httpbin.org","is_active":true,"priority":10}'

   curl -i http://localhost:8080/api/users/42
   ```

## Переменные окружения

| Переменная | По умолчанию | Описание |
|---|---|---|
| `PORT` | `8080` | Порт публичного прокси |
| `ADMIN_PORT` | `9090` | Порт admin API |
| `LOG_LEVEL` | `info` | `trace`/`debug`/`info`/`warn`/`error` |
| `LOG_FORMAT` | `text` | `text` или `json` |
| `JWT_SECRET` | — (обязательно) | Секрет для подписи JWT (≥16 символов) |
| `ADMIN_USERNAME` | — (обязательно) | Логин bootstrap-админа (создаётся при пустой таблице) |
| `ADMIN_PASSWORD` | — (обязательно) | Пароль bootstrap-админа (хешируется bcrypt cost=12) |
| `MAX_BODY_BYTES` | `10485760` (10 MiB) | Лимит тела запроса публичного прокси |
| `RATE_LIMIT_TTL` | `5m` | TTL неактивного бакета rate limiter'а |
| `LOG_BUFFER_SIZE` | `4096` | Размер кольцевого буфера async-логгера |
| `LOG_BATCH_SIZE` | `200` | Размер батча для `pgx.CopyFrom` (≤ buffer) |
| `LOG_FLUSH_PERIOD` | `500ms` | Период принудительного flush'а |
| `TLS_ENABLED` | `false` | Включает TLS для обоих серверов |
| `CERT_FILE` | — | Путь к сертификату (если `TLS_ENABLED=true`) |
| `KEY_FILE` | — | Путь к ключу (если `TLS_ENABLED=true`) |
| `DB_HOST` | `localhost` | Хост PostgreSQL |
| `DB_PORT` | `5432` | Порт PostgreSQL |
| `DB_USER` | `postgres` | Пользователь БД |
| `DB_PASSWORD` | — (обязательно) | Пароль БД |
| `DB_NAME` | `postgres` | Имя БД |
| `DB_SSLMODE` | `disable` | SSL-режим pgx |
| `DB_MAX_OPEN_CONNS` | `25` | Размер пула |
| `DB_MAX_IDLE_CONNS` | `10` | Idle в пуле |
| `DB_MAX_LIFETIME_MIN` | `30` | Время жизни соединения, минут |

## Сборка бинарника

```bash
make build         # → bin/gateway
go build -o bin/gateway ./cmd
```

Кросс-сборка под Linux/amd64:

```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o bin/gateway ./cmd
```

## Запуск тестов

Юнит-тесты (без Docker):

```bash
make test          # go test ./... -count=1 -cover
```

Интеграционные тесты (требуют Docker, поднимают postgres через testcontainers, помечены build-tag `integration`):

```bash
make test-integration   # go test ./... -tags=integration -count=1
```

Покрытие в HTML:

```bash
make cover
open coverage.html
```

Линтеры:

```bash
make lint          # go vet + golangci-lint
```

## Миграции БД

Миграции лежат в [`migrations/`](migrations/) и встроены в бинарник через `embed.FS`. На старте сервиса автоматически выполняется `MigrateUp`.

Ручные операции (используется CLI `golang-migrate`):

```bash
make migrate-up
make migrate-down  # откат на одну версию
```

CLI читает те же `.env`-переменные. Установить `migrate` можно так:

```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

## HTTP API

Все эндпоинты admin API — на `ADMIN_PORT` (`9090` по умолчанию). Префикса `/admin` нет.

### Аутентификация

```bash
# Получить токен
curl -s -X POST http://localhost:9090/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"admin-pass"}'
# {"token":"eyJhbGciOi..."}
```

Все защищённые эндпоинты требуют `Authorization: Bearer <token>`. Открытыми остаются `/login`, `/health`, `/live`, `/ready`.

### Маршруты

```bash
# Список
curl -s http://localhost:9090/routes -H "Authorization: Bearer $TOKEN"

# Создание
curl -s -X POST http://localhost:9090/routes \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"method":"GET","path_pattern":"/api/users/{id}","target_url":"https://httpbin.org","is_active":true,"priority":10,"rate_limit":60,"require_auth":false}'

# Обновление
curl -s -X PUT http://localhost:9090/routes/1 \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"method":"GET","path_pattern":"/api/users/{id}","target_url":"https://httpbin.org","is_active":true,"priority":20}'

# Toggle (активный/нет)
curl -s -X PATCH http://localhost:9090/routes/1/toggle -H "Authorization: Bearer $TOKEN"

# Удаление
curl -s -X DELETE http://localhost:9090/routes/1 -H "Authorization: Bearer $TOKEN"
```

Поддержка path-параметров:

- `*` — суффикс любого пути (например, `/static/*`).
- `{name}` — именованный параметр; в апстрим уйдёт заголовок `X-Path-Param-Name: <value>` (имя приведено к Title-Case с разделителем `-`).

### Логи и сводные метрики

```bash
# Логи (фильтры: path, status, from, to – RFC3339; пагинация limit/offset)
curl -s "http://localhost:9090/logs?status=500&limit=20" -H "Authorization: Bearer $TOKEN"

# Сводная метрика (period=hour|day|week)
curl -s "http://localhost:9090/metrics-summary?period=day" -H "Authorization: Bearer $TOKEN"
```

### Public-эндпоинты

На `PORT` (по умолчанию `8080`) работают:

- Любой динамический маршрут из БД (см. выше).
- `GET /auth/register`, `POST /auth/login` — регистрация и вход обычных пользователей (если используете user-auth).
- `GET /health`, `GET /live`, `GET /ready` — health-checks (см. ниже).

## Метрики и health-checks

- `GET /metrics` (на admin-порту, **под JWT**) — Prometheus-экспортёр (`http_requests_total`, `http_request_duration_seconds`, `rate_limit_hits_total`, `active_routes_count`, `db_connections`, `goroutines_count` + дефолтные Go/process collectors).
- `GET /health` — всегда `200 OK`.
- `GET /live` — алиас `/health`.
- `GET /ready` — `200`, если `db.Ping` проходит; иначе `503`.

## TLS

Включается переменными окружения:

```bash
TLS_ENABLED=true
CERT_FILE=/etc/ssl/gateway.crt
KEY_FILE=/etc/ssl/gateway.key
```

При `TLS_ENABLED=true` без указания `CERT_FILE`/`KEY_FILE` сервис не стартует (валидация конфигурации).

## WebSocket

Прокси автоматически распознаёт `Connection: Upgrade` + `Upgrade: websocket` и пересобирает заголовки. Для маршрута достаточно прописать `target_url: http://backend:8080`.

## Структура проекта

```
.
├── cmd/main.go
├── internal/
│   ├── domain/
│   │   ├── entity/                 # Route, RequestLog, AdminUser, User, Metrics
│   │   └── repository/             # интерфейсы и общие ошибки
│   ├── application/                # use cases (Route, AdminAuth, AuthUseCase, Log, Metric)
│   ├── infrastructure/
│   │   ├── auth/                   # JWT-сервис
│   │   ├── config/                 # загрузка/валидация конфигурации
│   │   ├── db/                     # pgxpool, миграции (golang-migrate + embed.FS)
│   │   ├── limiter/                # шардированный IP rate limiter
│   │   ├── log/                    # logrus + батчевый async-логгер
│   │   ├── metrics/                # Prometheus registry/middleware
│   │   ├── proxy/                  # reverse proxy builder (WS-aware)
│   │   ├── repo/                   # PostgreSQL-репозитории
│   │   └── router/                 # trie-роутер для динамических маршрутов
│   └── presentation/
│       ├── admin/                  # admin API + middleware (CORS, JWT, logging, recovery, metrics)
│       └── public/                 # публичный прокси + middleware (security, request-id, recovery, max-body, rate-limit, logging, auth)
├── migrations/                     # 000001..000004_*.up.sql/.down.sql + embed
├── .github/workflows/ci.yml
├── Makefile
├── README.md
├── CONTRIBUTING.md
├── CODE_OF_CONDUCT.md
└── .env.example
```

## Лицензия

См. файл `LICENSE`, если он присутствует в репозитории. По умолчанию проект распространяется на условиях MIT, если иное не указано.
