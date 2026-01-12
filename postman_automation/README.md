# Task Manager API

Простой REST API для демонстрации автоматизации тестирования в Postman.

## Запуск

```bash
go run cmd/server/main.go
```

Сервер запустится на `http://localhost:8080`

Или через Docker:
```bash
docker-compose up
```

## Тестирование в Postman

### Шаг 1: Импорт коллекции

1. Откройте Postman
2. Нажмите **Import**
3. Выберите файл `postman-examples/collection-example.json`
4. Коллекция появится в вашем workspace

### Шаг 2: Запуск тестов

1. Откройте импортированную коллекцию "Task Manager API"
2. Нажмите **Run collection**
3. Нажмите **Run Task Manager API**
4. Смотрите результаты тестов

Все переменные (`auth_token`, `user_id`, `task_id`) подставляются автоматически через тесты.

## Альтернатива: Импорт из OpenAPI

1. Postman → **Import** → `api/openapi.yaml`
2. Postman создаст коллекцию автоматически
3. Настройте Environment:
   - `base_url` = `http://localhost:8080`
   - `auth_token` = (заполнится автоматически)

## Эндпоинты

| Метод | URL | Описание |
|-------|-----|----------|
| POST | `/api/v1/auth/register` | Регистрация → получить user_id |
| POST | `/api/v1/auth/login` | Логин → получить token |
| POST | `/api/v1/tasks` | Создать задачу → получить task_id |
| GET | `/api/v1/tasks` | Список задач |
| GET | `/api/v1/tasks/{id}` | Получить задачу |
| PUT | `/api/v1/tasks/{id}` | Обновить задачу |
| PATCH | `/api/v1/tasks/{id}/status` | Изменить статус |
| DELETE | `/api/v1/tasks/{id}` | Удалить задачу |

## Что демонстрирует

1. **Request Chaining** - автоматическая подстановка значений между запросами
2. **JWT авторизация** - Bearer token в заголовках
3. **Автоматические тесты** - проверка ответов в Postman
4. **OpenAPI → Postman** - генерация коллекции из Swagger

## Порт занят?

```bash
PORT=8081 go run cmd/server/main.go
```
