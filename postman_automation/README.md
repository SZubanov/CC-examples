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

1. Postman → **Import** → выберите `postman_collection.json`
2. Postman → **Import** → выберите `postman_environment.json`
3. Откройте коллекцию "Task Manager API - Complete Test Suite"
4. Нажмите **Run collection**
5. Нажмите **Run**

## Что демонстрирует

1. **Request Chaining** - автоматическая подстановка значений между запросами
2. **JWT авторизация** - Bearer token в заголовках
3. **Автоматические тесты** - проверка ответов в Postman
4. **OpenAPI → Postman** - генерация коллекции из Swagger

## Порт занят?

```bash
PORT=8081 go run cmd/server/main.go
```
