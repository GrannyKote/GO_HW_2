# GO_HW_2

**Gateway** — HTTP-API для клиентов. **Ledger** — доменная логика, PostgreSQL и Redis. Gateway импортирует пакет `ledger/ledger`: те же эндпоинты, что в ДЗ 5, но данные хранятся в БД и переживают перезапуск; сводные отчёты кешируются в Redis.

Подробнее: [gateway/README.md](gateway/README.md), [ledger/README.md](ledger/README.md).

## Требования

- Go 1.21+
- PostgreSQL, Redis
- [Goose](https://github.com/pressly/goose) для миграций

## Быстрый старт

```powershell
docker compose up -d
go install github.com/pressly/goose/v3/cmd/goose@latest

$env:DATABASE_URL = "postgres://postgres:postgres@localhost:5432/cashapp?sslmode=disable"
$env:REDIS_ADDR = "localhost:6379"

cd ledger
goose -dir ./migrations postgres $env:DATABASE_URL up
go run .
```

В другом терминале:

```powershell
$env:DATABASE_URL = "postgres://postgres:postgres@localhost:5432/cashapp?sslmode=disable"
$env:REDIS_ADDR = "localhost:6379"
cd gateway
go run .
```

### Миграции

```powershell
goose -dir ./ledger/migrations postgres $env:DATABASE_URL up
goose -dir ./ledger/migrations postgres $env:DATABASE_URL down
```

Таблицы: `budgets` (категория, лимит), `expenses` (транзакции).

### Переменные окружения

- **PostgreSQL:** `DATABASE_URL` или `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASS`, `DB_NAME` (по умолчанию `localhost:5432`, `postgres`/`postgres`, БД `cashapp`, `sslmode=disable`).
- **Redis:** `REDIS_ADDR` (`localhost:6379`), `REDIS_DB` (`0`), `REDIS_PASSWORD` (опционально).

## Gateway — эндпоинты

| Метод | Путь | Описание |
|-------|------|----------|
| `POST` | `/api/transactions` | Создать транзакцию → таблица `expenses` |
| `GET` | `/api/transactions` | Список из БД |
| `POST` | `/api/budgets` | UPSERT бюджета → таблица `budgets` |
| `GET` | `/api/budgets` | Список (кеш Redis ~20 с) |
| `GET` | `/api/reports/summary?from=&to=` | Отчёт по категориям (кеш Redis 30 с) |

Примеры:

```powershell
curl.exe -X POST http://localhost:8080/api/budgets -H "Content-Type: application/json" -d "{\"category\":\"food\",\"limit\":5000}"
curl.exe -X POST http://localhost:8080/api/transactions -H "Content-Type: application/json" -d "{\"amount\":450,\"category\":\"food\",\"description\":\"lunch\",\"date\":\"2025-09-10\"}"
curl.exe http://localhost:8080/api/transactions
curl.exe "http://localhost:8080/api/reports/summary?from=2025-01-01&to=2025-12-31"
```

Второй подряд `GET /api/reports/summary` с теми же датами обслуживается из Redis (cache hit). После перезапуска Gateway данные остаются в PostgreSQL.

## Тесты

Нужны PostgreSQL и Redis (те же `DATABASE_URL` / `REDIS_ADDR`). Интеграционные тесты пропускаются, если БД недоступна.

```powershell
go test ./gateway/internal/api ./ledger/ledger -cover -coverprofile=cover.out
go tool cover -func=cover.out
go tool cover -html=cover.out -o cover.html
```

Покрытие **~55%** (зависит от окружения; unit-тесты валидации не требуют БД).

## Ledger (демо)

```powershell
cd ledger
go run .
```

Загружает `budgets.json`, демонстрирует добавление транзакций с проверкой лимитов в PostgreSQL.
