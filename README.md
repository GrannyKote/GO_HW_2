# GO_HW_2

Два сервиса в одном репозитории: **Gateway** (HTTP-шлюз) и **Ledger** (учёт транзакций и бюджетов в памяти).

## Требования

- Go 1.21+ (или любая актуальная версия с поддержкой modules)

## Gateway

HTTP-сервер на порту `8080`. Использует пакет `ledger` из этого репозитория (in-memory хранилище). Все ответы — JSON с заголовком `Content-Type: application/json; charset=utf-8`.

```powershell
cd gateway
go run .
```

### Эндпоинты

| Метод | Путь | Описание |
|-------|------|----------|
| `POST` | `/api/transactions` | Создать транзакцию |
| `GET` | `/api/transactions` | Список транзакций |
| `POST` | `/api/budgets` | Создать или обновить бюджет категории |
| `GET` | `/api/budgets` | Список бюджетов |

Ошибки возвращаются в формате `{"error":"..."}` со статусами `400` (валидация), `409` (превышение бюджета), `500` (внутренняя ошибка).

### Примеры cURL

Создать транзакцию:

```powershell
curl.exe -X POST http://localhost:8080/api/transactions `
  -H "Content-Type: application/json" `
  -d "{\"amount\":450,\"category\":\"food\",\"description\":\"lunch\",\"date\":\"2025-09-10\"}"
```

Список транзакций:

```powershell
curl.exe http://localhost:8080/api/transactions
```

Задать бюджет:

```powershell
curl.exe -X POST http://localhost:8080/api/budgets `
  -H "Content-Type: application/json" `
  -d "{\"category\":\"food\",\"limit\":5000}"
```

Список бюджетов:

```powershell
curl.exe http://localhost:8080/api/budgets
```

Превышение бюджета (пример — сумма больше лимита категории):

```powershell
curl.exe -X POST http://localhost:8080/api/transactions `
  -H "Content-Type: application/json" `
  -d "{\"amount\":10000,\"category\":\"food\",\"description\":\"too much\",\"date\":\"2025-09-10\"}"
```

Ожидается `409 Conflict` и `{"error":"budget exceeded"}`.

## Ledger

Сервис хранит транзакции и бюджеты по категориям в памяти.

- Бюджеты загружаются из `ledger/budgets.json` (JSON-массив объектов с полями `Category` и `Limit`).
- `AddTransaction` проверяет лимит для категорий, для которых задан бюджет; при превышении возвращает ошибку `budget exceeded` и не сохраняет транзакцию.
- `SetBudget` добавляет или обновляет бюджет категории.

```powershell
cd ledger
go run .
```

Ожидаемое поведение при старте:

1. Сообщение о загрузке бюджетов из `budgets.json`.
2. Успешное добавление транзакций в пределах лимита (`Food within budget`, `Transport within budget`, и т.д.).
3. Отказ для транзакции `Food over budget` с сообщением `budget exceeded`.
4. Успех для категории без бюджета (`Uncategorized`).
5. Ожидаемая ошибка для транзакции с нулевой суммой.
6. Итоговый список транзакций **без** отклонённой записи по категории `food` на 4000.

Формат `budgets.json`:

```json
[
  {"Category": "food", "Limit": 5000},
  {"Category": "transport", "Limit": 3000}
]
```
