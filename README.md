# org-api-test-task

REST API организационной структуры компании: подразделения с иерархией (дерево) и сотрудники.

## Стек

- **Go 1.24**, `net/http` (стандартный мультиплексор Go 1.22+ с pattern matching)
- **PostgreSQL 18**
- **GORM** — ORM
- **goose** — миграции (запускаются автоматически при старте)
- **slog** — структурированное JSON-логирование
- **Docker** + **docker compose**
- **testify** — тесты сервисного слоя

## Структура проекта

Чёткое разделение слоёв: handler → service → repository. Сервис не знает про GORM,
зависит только от интерфейсов репозиториев и `Transactor`. Транзакции прокидываются
через `context.Context` (паттерн transaction-in-context), что позволяет одному
сервисному методу атомарно вызывать несколько репозиториев в одной транзакции
без нарушения слоистости.

## Запуск

Требуется Docker и docker compose.

```bash
cp .env.example .env
docker compose up --build -d
```

Или через Makefile:

```bash
make up      # запустить
make logs    # смотреть логи api
make down    # остановить
make restart # перезапустить
make test    # запустить тесты
```

API будет доступно на `http://localhost:8080` (порт настраивается через `APP_PORT`
в `.env`). Миграции применяются автоматически при старте.

## Эндпоинты

### 1. Создать подразделение

```bash
curl -X POST http://localhost:8080/departments/ \
  -H 'Content-Type: application/json' \
  -d '{"name": "Engineering"}'
```

Ответ `201 Created`:
```json
{"id": 1, "name": "Engineering", "parent_id": null, "created_at": "2026-05-16T10:00:00Z"}
```

С родителем:
```bash
curl -X POST http://localhost:8080/departments/ \
  -H 'Content-Type: application/json' \
  -d '{"name": "Backend", "parent_id": 1}'
```

### 2. Создать сотрудника в подразделении

```bash
curl -X POST http://localhost:8080/departments/2/employees/ \
  -H 'Content-Type: application/json' \
  -d '{"full_name": "Иван Иванов", "position": "Senior Go Developer", "hired_at": "2024-03-15"}'
```

`hired_at` опционален, формат `YYYY-MM-DD`.

### 3. Получить подразделение с поддеревом

```bash
curl 'http://localhost:8080/departments/1?depth=3&include_employees=true'
```

Параметры:
- `depth` (int, по умолчанию 1, максимум 5) — глубина дочерних подразделений в ответе
- `include_employees` (bool, по умолчанию true) — включать ли сотрудников

Ответ:
```json
{
  "department": {"id": 1, "name": "Engineering", "parent_id": null, "created_at": "..."},
  "employees": [],
  "children": [
    {
      "department": {"id": 2, "name": "Backend", "parent_id": 1, "created_at": "..."},
      "employees": [
        {"id": 1, "department_id": 2, "full_name": "Иван Иванов", "position": "Senior Go Developer", "hired_at": "2024-03-15T00:00:00Z", "created_at": "..."}
      ],
      "children": []
    }
  ]
}
```

### 4. Переименовать / переместить подразделение (PATCH)

```bash
# переименовать
curl -X PATCH http://localhost:8080/departments/2 \
  -H 'Content-Type: application/json' \
  -d '{"name": "Backend Team"}'

# переместить в другого родителя
curl -X PATCH http://localhost:8080/departments/2 \
  -H 'Content-Type: application/json' \
  -d '{"parent_id": 5}'

# вынести в корень (parent_id = null)
curl -X PATCH http://localhost:8080/departments/2 \
  -H 'Content-Type: application/json' \
  -d '{"parent_id": null}'
```

PATCH-семантика реализована честно: можно отличить «поле не передано» от
«передано `null`» через кастомный `UnmarshalJSON`. Передача `parent_id: null`
переносит подразделение в корень дерева.

### 5. Удалить подразделение

Два режима:

**cascade** — удалить подразделение со всеми дочерними подразделениями и сотрудниками
(каскад выполняется на уровне БД через `ON DELETE CASCADE`):

```bash
curl -X DELETE 'http://localhost:8080/departments/2?mode=cascade'
```

**reassign** — удалить подразделение, а его сотрудников и дочерние подразделения
перенести в `reassign_to_department_id`:

```bash
curl -X DELETE 'http://localhost:8080/departments/2?mode=reassign&reassign_to_department_id=1'
```

Оба возвращают `204 No Content`.

## Валидация и бизнес-правила

- `name`, `full_name`, `position` — не пустые, длина 1..200, пробелы по краям тримятся
- Имя подразделения уникально в пределах одного родителя (PostgreSQL constraint
  `UNIQUE NULLS NOT DISTINCT (name, parent_id)` — современный синтаксис PG 15+,
  корректно работает и для корневых подразделений с `parent_id IS NULL`)
- Подразделение не может быть собственным родителем → `400 Bad Request`
- Перенос подразделения внутрь своего же поддерева → `409 Conflict` (защита от циклов
  обходом по предкам с safety bound)
- Создание сотрудника в несуществующем подразделении → `404 Not Found`
- Reassign в подразделение, которое находится в поддереве удаляемого → `409 Conflict`

## Формат ошибок

Все ошибки возвращаются в JSON:
```json
{"error": "human-readable message", "code": "machine_code"}
```

Коды: `invalid_json`, `invalid_id`, `invalid_depth`, `invalid_query`,
`not_found`, `conflict`, `bad_request`, `internal_error`.

## Тесты

```bash
make test
# или
go test ./... -v
```

Тесты сервисного слоя написаны с использованием мок-репозиториев (без поднятия БД),
проверяют валидацию и бизнес-правила.

## Что реализовано из необязательного

- Структурное JSON-логирование (slog) с middleware для HTTP-запросов
- Graceful shutdown по SIGINT/SIGTERM с таймаутом
- Server timeouts (ReadHeaderTimeout / ReadTimeout / WriteTimeout / IdleTimeout)
- Non-root user и ca-certificates в финальном Docker-образе
- Healthcheck для Postgres + `depends_on: service_healthy` для api
- `restart: unless-stopped` для обоих сервисов
- Тесты сервиса с моком (не только тесты парсинга на handler-уровне)
