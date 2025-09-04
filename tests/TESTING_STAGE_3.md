# Тестирование этапа 3: create_task

## Что реализовано

Добавлен инструмент `create_task` для создания новых задач с следующими возможностями:

- Требует JWT авторизацию
- Принимает имя пользователя в `assigned_to` вместо UUID
- Автоматически находит UUID пользователя по имени
- Проверяет существование пользователя, которому назначается задача
- Возвращает имена пользователей в ответе (created_by_name и assigned_to_name)
- Создает задачу со статусом `pending`
- Автоматически заполняет `created_by` из JWT токена
- Возвращает созданную задачу с полной информацией

## Как протестировать

### 1. Запустите сервер

```bash
./simple-task-mcp
```

### 2. Создайте админского пользователя (если еще не создан)

```bash
./create-admin
```

Сохраните токен и ID пользователя из вывода.

### 3. Создайте обычного пользователя для назначения задач

Используя токен админа:

```bash
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <ADMIN_TOKEN>" \
  -d '{
    "jsonrpc": "2.0",
    "method": "tools/call",
    "params": {
      "name": "create_user",
      "arguments": {
        "name": "Test User",
        "is_admin": false
      }
    },
    "id": 1
  }'
```

Сохраните ID созданного пользователя.

### 4. Создайте задачу

Используя любой валидный токен (админа или обычного пользователя):

```bash
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <TOKEN>" \
  -d '{
    "jsonrpc": "2.0",
    "method": "tools/call",
    "params": {
      "name": "create_task",
      "arguments": {
        "description": "Implement user authentication module",
        "assigned_to": "Test User"
      }
    },
    "id": 1
  }'
```

### 5. Проверьте в базе данных

```sql
SELECT * FROM tasks ORDER BY created_at DESC LIMIT 5;
```

## Ожидаемые результаты

### Успешное создание задачи:
```json
{
  "id": "generated-uuid",
  "description": "Implement user authentication module",
  "status": "pending",
  "created_by": "user-id-from-token",
  "created_by_name": "Admin User",
  "assigned_to": "specified-user-id",
  "assigned_to_name": "Test User",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

### Ошибки валидации:

1. **Без авторизации**: "Authorization header is required"
2. **Невалидный токен**: "invalid token: ..."
3. **Несуществующий пользователь**: "user 'NonExistentUser' does not exist"
4. **Отсутствующие параметры**: "description is required" / "assigned_to is required"

## Альтернативный способ тестирования

Используйте готовый скрипт:

```bash
cd tests
./test_create_task.sh <TOKEN> "Username"
```

## Что проверить

1. ✅ Задача создается с корректными данными
2. ✅ `created_by` берется из JWT токена
3. ✅ `status` устанавливается в `pending`
4. ✅ Временные метки заполняются автоматически
5. ✅ Поиск пользователя по имени работает корректно
6. ✅ Проверяется существование пользователя для `assigned_to`
7. ✅ Возвращаются имена пользователей в ответе
8. ✅ Требуется авторизация
