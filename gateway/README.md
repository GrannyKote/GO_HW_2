# Gateway

HTTP-API на `:8080`. Эндпоинты из ДЗ 5 не менялись; добавлен отчёт `GET /api/reports/summary`. Gateway вызывает пакет `ledger/ledger`, который читает и пишет в **PostgreSQL** и кеширует отчёты в **Redis** — отдельный процесс Ledger для API не обязателен, но его можно запускать для проверки доменной логики.

## Запуск

1. PostgreSQL + Redis, миграции Goose (см. [ledger/README.md](../ledger/README.md)).
2. Переменные окружения (те же, что у Ledger):

```powershell
$env:DATABASE_URL = "postgres://postgres:postgres@localhost:5432/cashapp?sslmode=disable"
$env:REDIS_ADDR = "localhost:6379"
cd gateway
go run .
```

## Проверка cURL

Бюджет и транзакция:

```powershell
curl.exe -X POST http://localhost:8080/api/budgets -H "Content-Type: application/json" -d "{\"category\":\"food\",\"limit\":5000}"
curl.exe -X POST http://localhost:8080/api/transactions -H "Content-Type: application/json" -d "{\"amount\":450,\"category\":\"food\",\"description\":\"lunch\",\"date\":\"2025-09-10\"}"
curl.exe http://localhost:8080/api/transactions
curl.exe http://localhost:8080/api/budgets
```

Превышение бюджета — `409` и `{"error":"budget exceeded"}`.

Отчёт (второй запрос подряд быстрее за счёт Redis):

```powershell
curl.exe "http://localhost:8080/api/reports/summary?from=2025-01-01&to=2025-12-31"
curl.exe "http://localhost:8080/api/reports/summary?from=2025-01-01&to=2025-12-31"
```

Перезапустите Gateway и снова выполните `GET /api/transactions` — данные должны остаться в PostgreSQL.
