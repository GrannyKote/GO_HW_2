# Ledger

Бизнес-логика, PostgreSQL и Redis-кеш. Gateway импортирует пакет `ledger/ledger` и использует те же функции хранения.

## PostgreSQL и миграции (Goose)

1. Поднимите PostgreSQL (локально или через Docker из корня репозитория):

```powershell
docker compose up -d postgres redis
```

2. Создайте БД `cashapp` (в `docker-compose.yml` она создаётся автоматически).

3. Установите Goose:

```powershell
go install github.com/pressly/goose/v3/cmd/goose@latest
```

4. Задайте строку подключения и примените миграции:

```powershell
$env:DATABASE_URL = "postgres://postgres:postgres@localhost:5432/cashapp?sslmode=disable"
goose -dir ./migrations postgres $env:DATABASE_URL up
```

Откат:

```powershell
goose -dir ./migrations postgres $env:DATABASE_URL down
```

Миграции лежат в `ledger/migrations/` (таблицы `budgets`, `expenses`).

## Переменные окружения

| Переменная | По умолчанию | Описание |
|------------|--------------|----------|
| `DATABASE_URL` | — | Полный DSN PostgreSQL (приоритет) |
| `DB_HOST` | `localhost` | Хост БД |
| `DB_PORT` | `5432` | Порт |
| `DB_USER` | `postgres` | Пользователь |
| `DB_PASS` | `postgres` | Пароль |
| `DB_NAME` | `cashapp` | Имя БД |
| `REDIS_ADDR` | `localhost:6379` | Redis |
| `REDIS_DB` | `0` | Номер БД Redis |
| `REDIS_PASSWORD` | — | Пароль Redis (опционально) |

## Запуск

После `goose up`:

```powershell
cd ledger
$env:DATABASE_URL = "postgres://postgres:postgres@localhost:5432/cashapp?sslmode=disable"
$env:REDIS_ADDR = "localhost:6379"
go run .
```

При успешном старте в логе: подключение к PostgreSQL и Redis. Данные транзакций и бюджетов сохраняются в БД и переживают перезапуск.

## Кеш

- `ListBudgets()` — ключ `budgets:all`, TTL ~20 с; после `SetBudget` ключ удаляется.
- `GetReportSummary(from, to)` — ключ `report:summary:<from>:<to>`, TTL 30 с.
