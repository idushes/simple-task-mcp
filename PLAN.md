# План разработки MCP Task Manager Server

## Текущий статус проекта
**Версия**: 0.1.2  
**Статус**: Завершен  
**Завершенные этапы**: 13 из 13  
**Текущий этап**: Все этапы завершены

### Реализованные возможности:
- ✅ Базовая инфраструктура MCP сервера
- ✅ Подключение к PostgreSQL
- ✅ JWT авторизация
- ✅ Создание пользователей (create_user)
- ✅ Создание задач (create_task)
- ✅ Список созданных задач (list_created_tasks)
- ✅ Получение задач по фильтру (get_next_task)
- ✅ Завершение задач с результатом (complete_task)
- ✅ Отмена задач с причиной (cancel_task)
- ✅ Система комментариев и ожидание человека (wait_for_user)
- ✅ Список пользователей (list_users)

### Реализованные возможности (продолжение):
- ✅ Генерация токенов для существующих пользователей (generate_token)
- ✅ Получение информации о текущем токене (get_token_info)
- ✅ Совместимость с MCP клиентами (краткие текстовые сообщения в content)

## Описание проекта
MCP сервер для управления задачами, предназначенный для использования AI агентами. Сервер предоставляет инструменты для создания, обновления, архивирования и управления задачами через MCP протокол.

## Технологии
- **Язык**: Go
- **MCP библиотека**: mcp-go
- **База данных**: PostgreSQL
- **Конфигурация**: .env файл
- **Авторизация**: JWT токены
- **Особенности**: Graceful shutdown

## Структура проекта
```
simple-task-mcp/
├── .env                    # Конфигурация (не коммитится)
├── .env.example           # Пример конфигурации
├── .gitignore            # Git игнор файл
├── go.mod                # Go модуль
├── go.sum                # Go dependencies
├── main.go               # Точка входа
├── config/               # Конфигурация
│   └── config.go        # Загрузка конфигурации
├── database/            # База данных
│   ├── connection.go    # Подключение к БД
│   └── migrations/      # Миграции
│       ├── 001_users.sql
│       ├── 002_tasks.sql
│       ├── 004_add_result_to_tasks.sql
│       └── 005_task_comments.sql  # ✅ Таблица комментариев
├── models/              # Модели данных
│   ├── user.go         # Модель пользователя
│   ├── task.go         # Модель задачи
│   └── comment.go      # ✅ Модель комментария
├── auth/               # Авторизация
│   └── jwt.go          # JWT middleware
├── tools/              # MCP инструменты (каждый в отдельном файле)
│   ├── create_user.go     # ✅ Создание пользователей (админ)
│   ├── create_task.go     # ✅ Создание задач
│   ├── get_next_task.go   # ✅ Получение следующей задачи
│   ├── complete_task.go   # ✅ Завершение задач с результатом
│   ├── cancel_task.go         # ✅ Отмена задач с причиной
│   └── wait_for_user.go   # ✅ Ожидание человека с комментариями
├── tests/              # Тестовые скрипты и документация
│   ├── test_connection.sh
│   ├── test_create_task.sh
│   ├── test_get_next_task.sh
│   ├── test_complete_task.sh  # ✅ Тест завершения задач
│   ├── test_cancel_task.sh    # ✅ Тест отмены задач
│   ├── test_wait_for_user.sh  # ✅ Тест системы комментариев
│   ├── TESTING_STAGE_3.md
│   ├── TESTING_STAGE_4.md
│   └── README.md
├── cmd/               # Утилиты командной строки
│   └── create-admin/  # Создание админ пользователя
└── README.md          # Документация

```

## Модели данных

### Users
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    is_admin BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Индекс для быстрого поиска по имени
CREATE INDEX idx_users_name ON users(name);
```

### Tasks
```sql
CREATE TABLE tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    description TEXT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    created_by UUID NOT NULL REFERENCES users(id),
    assigned_to UUID NOT NULL REFERENCES users(id),
    is_archived BOOLEAN DEFAULT FALSE,
    result TEXT, -- Результат выполнения задачи
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP WITH TIME ZONE,
    archived_at TIMESTAMP WITH TIME ZONE
);

-- Индексы для производительности
CREATE INDEX idx_tasks_status ON tasks(status) WHERE NOT is_archived;
CREATE INDEX idx_tasks_assigned_to ON tasks(assigned_to) WHERE NOT is_archived;
CREATE INDEX idx_tasks_created_by ON tasks(created_by) WHERE NOT is_archived;
```

### Task Comments
```sql
CREATE TABLE task_comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    created_by UUID NOT NULL REFERENCES users(id),
    comment TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Индексы для быстрого получения комментариев по задаче
CREATE INDEX idx_task_comments_task_id ON task_comments(task_id);
CREATE INDEX idx_task_comments_created_at ON task_comments(created_at);
```

### Статусы задач:
- `pending` - новая задача
- `in_progress` - в работе
- `waiting_for_user` - ожидает реакции пользователя (человека)
- `completed` - завершена
- `cancelled` - отменена

## Список этапов разработки

### MCP инструменты:
- [x] `create_user` - админский инструмент создания пользователей
- [x] `create_task` - Создание новой задачи
- [x] `list_created_tasks` - Получение списка задач, созданных пользователем
- [x] `get_next_task` - Получение следующей задачи по фильтру статусов
- [x] `complete_task` - Завершение задачи с результатом
- [x] `cancel_task` - Отмена задачи с указанием причины
- [x] `wait_for_user` - Отправка задачи в ожидание человека с комментарием
- [x] `generate_token` - Генерация нового JWT токена для существующего пользователя (админ)
- [x] `get_token_info` - Получение информации о текущем JWT токене
- [x] `list_users` - Получение списка всех пользователей

### Инфраструктура:
- [x] Docker контейнеризация
- [x] GitHub Actions CI/CD

## Этапы разработки

### Этап 1: Базовая инфраструктура и ping ✅
**Цель**: Настроить проект и базовый MCP сервер без подключения к БД

**Статус**: Завершен

**Задачи**:
1. ✅ Создать структуру проекта
2. ✅ Инициализировать Go модуль
3. ✅ Настроить .env конфигурацию (только MCP_SERVER_PORT и LOG_LEVEL)
4. ✅ Настроить базовый MCP сервер
5. ✅ Реализовать graceful shutdown

**Проверка**: ✅ Завершена
- ✅ Сервер запускается на порту из конфигурации
- ✅ Корректно завершается по Ctrl+C (graceful shutdown)
- ✅ Логирование работает

### Этап 2: Подключение к БД и JWT авторизация ✅
**Цель**: Добавить подключение к БД и JWT авторизацию

**Статус**: Завершен

**Задачи**:
1. ✅ Добавить подключение к PostgreSQL
2. ✅ Создать миграции для таблиц users и tasks
3. ✅ Создать модели User и Task
4. ✅ Реализовать JWT middleware
5. ✅ Создать инструмент create_user для админов

**Проверка**: ✅ Завершена
- ✅ Подключается к БД по DATABASE_URL
- ✅ Миграции применяются успешно
- ✅ JWT токен валидируется корректно

### Этап 2.1: Список пользователей (list_users) ✅
**Цель**: Добавить инструмент для просмотра списка пользователей и поле описания пользователя

**Статус**: Завершен

**Задачи**:
1. ✅ Добавить миграцию для добавления поля description в таблицу users
2. ✅ Обновить модель User с новым полем
3. ✅ Создать MCP инструмент list_users
4. ✅ Зарегистрировать инструмент в сервере

**MCP инструмент**:

<details>
<summary><b>list_users</b> - Получение списка всех пользователей</summary>

```go
listUsersTool := mcp.NewTool("list_users",
    mcp.WithDescription("List all users in the system"),
    mcp.WithNumber("limit",
        mcp.Description("Maximum number of users to return (default: 100, max: 1000)"),
        mcp.DefaultNumber(100),
    ),
)
```

**Реализованные возможности**:
- Требует валидный JWT токен
- Возвращает список пользователей с поддержкой лимита
- Параметр limit (опциональный, по умолчанию 100, максимум 1000)
- Включает опциональное поле description
- Сортирует пользователей по имени (name)
- Возвращает полную информацию о пользователях включая ID, имя, описание, права админа и даты создания/обновления
- Возвращает количество найденных пользователей и примененный лимит

**Формат возвращаемых данных**:
```json
{
  "users": [
    {
      "id": "uuid-пользователя",
      "name": "имя_пользователя",
      "description": "описание пользователя (опционально)",
      "is_admin": true/false,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "count": 5,
  "limit": 100
}
```

**Поля пользователя**:
- `id` (string): UUID пользователя
- `name` (string): уникальное имя пользователя
- `description` (string, optional): описание пользователя (может отсутствовать)
- `is_admin` (boolean): права администратора
- `created_at` (string): дата создания в формате ISO 8601
- `updated_at` (string): дата обновления в формате ISO 8601

**Мета-информация**:
- `count` (number): количество возвращенных пользователей
- `limit` (number): примененный лимит

**Файлы**: `tools/list_users.go`, `database/migrations/006_add_user_description.sql`
</details>

**Проверка**: ✅ Завершена
- ✅ Добавлено поле description в таблицу users
- ✅ Обновлена модель User с новым полем
- ✅ Создан и зарегистрирован инструмент list_users
- ✅ Инструмент возвращает список всех пользователей
- ✅ Включает опциональное поле description
- ✅ Сортирует пользователей по имени
- ✅ Детально описан формат возвращаемых данных с примерами JSON
- ✅ Документированы все поля пользователя и мета-информация

### Этап 3: Создание задач (create_task) ✅
**Цель**: Реализовать создание новых задач

**Статус**: Завершен

**MCP инструмент**:

<details>
<summary><b>create_task</b> - Создание новой задачи</summary>

```go
createTaskTool := mcp.NewTool("create_task",
    mcp.WithDescription("Create a new task and assign it to a user"),
    mcp.WithString("description",
        mcp.Required(),
        mcp.Description("Task description"),
    ),
    mcp.WithString("assigned_to",
        mcp.Required(),
        mcp.Description("Username to assign the task to"),
    ),
)
```

**Реализованные возможности**:
- Требует валидный JWT токен
- Создает задачу с created_by из токена
- Возвращает ID созданной задачи и имена пользователей
- Находит пользователя по имени для assigned_to
- Проверяет существование пользователя по имени
- Автоматически устанавливает статус "pending"
- Возвращает детальную информацию о созданной задаче

**Файлы**: `tools/create_task.go`, `tests/test_create_task.sh`, `tests/TESTING_STAGE_3.md`
</details>

**Проверка**: ✅ Завершена

### Этап 3.1: Список созданных задач (list_created_tasks) ✅
**Цель**: Реализовать получение полного списка задач, созданных пользователем

**Статус**: Завершен

**MCP инструмент**:

<details>
<summary><b>list_created_tasks</b> - Получение списка задач, созданных пользователем</summary>

```go
listCreatedTasksTool := mcp.NewTool("list_created_tasks",
    mcp.WithDescription("Get a list of tasks created by the current user or specified user (admins only)"),
    mcp.WithString("user_name",
        mcp.Description("Username to get tasks for. If not provided, uses current user. Only admins can specify other users."),
    ),
    mcp.WithNumber("limit",
        mcp.Description("Maximum number of tasks to return (default: 50, max: 1000)"),
    ),
    mcp.WithArray("statuses",
        mcp.Description("Array of statuses to filter by. Available statuses: pending, in_progress, waiting_for_user, completed, cancelled. If not provided, returns tasks with all statuses."),
        mcp.Items(map[string]any{"type": "string"}),
    ),
)
```

**Реализованные возможности**:
- Требует валидный JWT токен
- Возвращает список задач, созданных текущим пользователем
- Поддержка параметра `user_name` для просмотра задач других пользователей (только для админов)
- Ограничение доступа: обычные пользователи видят только свои созданные задачи
- Параметр `limit` для ограничения количества результатов (по умолчанию 50, максимум 1000)
- Параметр `statuses` для фильтрации по статусам задач
- Доступные статусы для фильтрации: pending, in_progress, waiting_for_user, completed, cancelled
- Возвращает полную информацию о задачах включая комментарии
- Сортировка по дате создания (новые первые)
- Включает архивные и неархивные задачи
- Возвращает статистику: общее количество задач и примененный лимит

**Файлы**: `tools/list_created_tasks.go`
</details>

**Проверка**: ✅ Завершена
- ✅ Возвращает задачи, созданные текущим пользователем
- ✅ Админы могут просматривать задачи других пользователей через параметр user_name
- ✅ Обычные пользователи не могут указывать других пользователей
- ✅ Поддержка лимитирования результатов
- ✅ Фильтрация по статусам задач через параметр statuses
- ✅ Валидация статусов с проверкой на корректность, пустые строки и дублирование
- ✅ Информативные сообщения об ошибках валидации
- ✅ Включает комментарии к задачам
- ✅ Сортировка по дате создания
- ✅ Показывает как архивные, так и активные задачи
- ✅ Корректная JSON Schema с указанием типа элементов массива (mcp.Items)

### Этап 4: Получение следующей задачи (get_next_task) ✅
**Цель**: Реализовать получение задач по фильтру

**Статус**: Завершен

**MCP инструмент**:

<details>
<summary><b>get_next_task</b> - Получение следующей задачи по фильтру статусов</summary>

```go
getNextTaskTool := mcp.NewTool("get_next_task",
    mcp.WithDescription("Get one task where the current user is assignee, filtered by status"),
    mcp.WithArray("statuses",
        mcp.Description("Array of statuses to filter by. Available statuses: pending, in_progress, waiting_for_user, completed, cancelled. If not provided, defaults to [\"pending\"]"),
        mcp.Items(map[string]any{"type": "string"}),
    ),
)
```

**Реализованные возможности**:
- Возвращает одну задачу, где текущий пользователь является исполнителем (assigned_to)
- Задачи сортируются по created_at (старые первые)
- Поддерживает фильтрацию по массиву статусов
- Значение по умолчанию: ["pending"] если статусы не указаны
- Возвращает как имена пользователей, так и их UUID
- Перечисляет доступные статусы в описании параметра
- Возвращает null если подходящих задач нет

**Файлы**: `tools/get_next_task.go`, `tests/test_get_next_task.sh`, `tests/TESTING_STAGE_4.md`
</details>

**Проверка**: ✅ Завершена
- ✅ Возвращает только неархивные задачи
- ✅ Фильтрует по статусам из массива
- ✅ Возвращает задачи где пользователь является исполнителем (assigned_to)
- ✅ Возвращает null если задач нет
- ✅ Использует значение по умолчанию ["pending"]
- ✅ Включает имена и UUID пользователей
- ✅ Валидация статусов с проверкой на корректность, пустые строки и дублирование
- ✅ Информативные сообщения об ошибках валидации
- ✅ Корректная JSON Schema с указанием типа элементов массива (mcp.Items)

### Этап 5: Завершение задач (complete_task) ✅
**Цель**: Реализовать завершение задач с добавлением результата

**Статус**: Завершен

**MCP инструмент**:

<details>
<summary><b>complete_task</b> - Завершение задачи с результатом</summary>

```go
completeTaskTool := mcp.NewTool("complete_task",
    mcp.WithDescription("Mark task as completed and optionally add result"),
    mcp.WithString("id",
        mcp.Required(),
        mcp.Description("Task ID (UUID)"),
    ),
    mcp.WithString("result",
        mcp.Description("Task completion result or notes"),
    ),
)
```
</details>

**Реализованные возможности**:
- Требует валидный JWT токен
- Проверяет права доступа (только создатель или исполнитель)
- Устанавливает статус "completed" и время завершения
- Сохраняет результат выполнения задачи (опциональный параметр)
- Обновляет updated_at автоматически
- Не позволяет завершать уже завершенные или архивные задачи
- Возвращает детальную информацию о завершенной задаче

**Файлы**: `tools/complete_task.go`, `tests/test_complete_task.sh`, `database/migrations/004_add_result_to_tasks.sql`

**Проверка**: ✅ Завершена
- ✅ Завершает задачи с установкой статуса "completed"
- ✅ Сохраняет результат выполнения (опционально)
- ✅ Проверяет права доступа
- ✅ Не позволяет завершать уже завершенные задачи
- ✅ Не позволяет завершать архивные задачи
- ✅ Обновляет поля completed_at и updated_at

### Этап 6: Отмена задач (cancel_task) ✅
**Цель**: Реализовать отмену задач с указанием причины

**Статус**: Завершен

**MCP инструмент**:

<details>
<summary><b>cancel_task</b> - Отмена задачи с указанием причины</summary>

```go
cancelTaskTool := mcp.NewTool("cancel_task",
    mcp.WithDescription("Cancel a task with cancellation reason"),
    mcp.WithString("id",
        mcp.Required(),
        mcp.Description("Task ID (UUID)"),
    ),
    mcp.WithString("reason",
        mcp.Required(),
        mcp.Description("Reason for task cancellation"),
    ),
)
```
</details>

**Реализованные возможности**:
- Требует валидный JWT токен
- Проверяет права доступа (только создатель или исполнитель)
- Устанавливает статус "cancelled"
- Добавляет причину отмены в конец поля result
- Обновляет updated_at автоматически
- Не позволяет отменять уже завершенные или архивные задачи
- Возвращает детальную информацию об отмененной задаче

**Файлы**: `tools/cancel_task.go`, `tests/test_cancel_task.sh`

**Проверка**: ✅ Завершена
- ✅ Отменяет задачи с установкой статуса "cancelled"
- ✅ Сохраняет причину отмены в поле result
- ✅ Добавляет причину к существующему результату (если есть)
- ✅ Проверяет права доступа (только создатель или исполнитель)
- ✅ Не позволяет отменять уже завершенные задачи
- ✅ Не позволяет отменять уже отмененные задачи
- ✅ Не позволяет отменять архивные задачи
- ✅ Обновляет поле updated_at

### Этап 7: Система комментариев и ожидание человека (wait_for_user) ✅
**Цель**: Реализовать систему комментариев к задачам и возможность отправки задач в ожидание человека

**Статус**: Завершен

**Компоненты**:
1. Таблица `task_comments` для хранения комментариев
2. MCP инструмент `wait_for_user` для отправки задач в ожидание с комментарием
3. Обновление `get_next_task` для включения комментариев

**MCP инструмент**:

<details>
<summary><b>wait_for_user</b> - Отправка задачи в ожидание человека с комментарием</summary>

```go
waitForUserTool := mcp.NewTool("wait_for_user",
    mcp.WithDescription("Send task to waiting for user status with comment"),
    mcp.WithString("id",
        mcp.Required(),
        mcp.Description("Task ID (UUID)"),
    ),
    mcp.WithString("comment",
        mcp.Required(),
        mcp.Description("Comment explaining why task needs user attention"),
    ),
)
```
</details>

**Реализованные возможности**:
- Требует валидный JWT токен
- Проверяет права доступа (только создатель или исполнитель)
- Устанавливает статус "waiting_for_user"
- Добавляет комментарий в таблицу task_comments
- Обновляет updated_at автоматически
- Не позволяет отправлять уже завершенные или архивные задачи в ожидание
- Возвращает детальную информацию о задаче

**Файлы**: 
- `tools/wait_for_user.go`
- `models/comment.go`
- `database/migrations/005_task_comments.sql`
- `tests/test_wait_for_user.sh`

**Проверка**: ✅ Завершена
- ✅ Создана таблица task_comments с индексами
- ✅ wait_for_user отправляет задачи в статус "waiting_for_user"
- ✅ Комментарии сохраняются в базе данных
- ✅ Проверяются права доступа (только создатель или исполнитель)
- ✅ get_next_task возвращает комментарии вместе с задачами
- ✅ Защита от отправки завершенных/отмененных задач в ожидание
- ✅ Поддержка множественных комментариев к одной задаче
- ✅ Транзакционная безопасность при добавлении комментариев

### Этап 7.5: Генерация токенов для существующих пользователей (generate_token) ✅
**Цель**: Реализовать админский инструмент для генерации новых JWT токенов существующим пользователям

**Статус**: Завершен

**MCP инструмент**:

<details>
<summary><b>generate_token</b> - Генерация нового JWT токена для существующего пользователя</summary>

```go
generateTokenTool := mcp.NewTool("generate_token",
    mcp.WithDescription("Generate a new JWT token for an existing user (admin only)"),
    mcp.WithString("user_id",
        mcp.Required(),
        mcp.Description("User ID (UUID) to generate token for"),
    ),
)
```
</details>

**Реализованные возможности**:
- Требует валидный JWT токен с правами администратора
- Проверяет существование пользователя по ID
- Генерирует новый JWT токен с годовым сроком действия
- Возвращает информацию о пользователе и новый токен
- Не изменяет данные пользователя в базе

**Файлы**: 
- `tools/generate_token.go`

**Проверка**: ✅ Завершена
- ✅ Только админ может генерировать токены
- ✅ Токен генерируется с правильными claims (user_id, is_admin)
- ✅ Возвращается ошибка при попытке сгенерировать токен для несуществующего пользователя
- ✅ Новый токен действителен и работает с другими инструментами

### Этап 7.6: Получение информации о текущем токене (get_token_info) ✅
**Цель**: Реализовать инструмент для получения информации о текущем JWT токене пользователя

**Статус**: Завершен

**MCP инструмент**:

<details>
<summary><b>get_token_info</b> - Получение информации о текущем JWT токене</summary>

```go
getTokenInfoTool := mcp.NewTool("get_token_info",
    mcp.WithDescription("Get information about the current JWT token"),
    mcp.WithInputSchema[struct{}](),
)
```
</details>

**Реализованные возможности**:
- Требует валидный JWT токен
- Не требует дополнительных параметров
- Декодирует JWT токен и возвращает информацию о пользователе
- Возвращает время истечения токена и оставшийся срок действия
- Возвращает информацию о правах администратора

**Файлы**: 
- `tools/get_token_info.go`

**Проверка**: ✅ Завершена
- ✅ Корректно декодирует JWT токен
- ✅ Возвращает информацию о пользователе (ID, имя, права)
- ✅ Показывает время истечения токена и оставшийся срок действия
- ✅ Обрабатывает ошибки при невалидном токене
- ✅ Корректная JSON Schema с пустой схемой входных данных для инструмента без параметров

### Этап 8: Docker контейнеризация ✅
**Цель**: Создать Docker образ для удобного развертывания

**Статус**: Завершен

**Задачи**:
1. ✅ Создать Dockerfile для сборки приложения
2. ✅ Добавить .dockerignore
3. ✅ Оптимизировать образ (multi-stage build)

**Файлы**:

<details>
<summary><b>Dockerfile</b> - Multi-stage build для оптимального размера</summary>

```dockerfile
# Build stage
FROM golang:1.24-alpine AS builder
RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
# Create default .env.example if it doesn't exist in the build context
RUN echo '# Database\nDATABASE_URL=postgres://user:password@postgres:5432/task_manager?sslmode=disable\n\n# MCP Server\nMCP_SERVER_PORT=8080\n\n# JWT\nJWT_SECRET=your-secret-key-here\n\n# Logging\nLOG_LEVEL=info' > .env.example
CMD ["./main"]
```
</details>

<details>
<summary><b>.dockerignore</b> - Исключение ненужных файлов</summary>

```
.git
.gitignore
.env
*.md
.github/
docker-compose.yml
Dockerfile
.dockerignore
```
</details>


**Проверка**: ✅ Завершена
- ✅ Docker образ собирается успешно
- ✅ Размер образа минимальный (alpine + бинарник)
- ✅ Приложение запускается из образа
- ✅ Переменные окружения передаются корректно


## Конфигурация .env
```env
# Database
DATABASE_URL=postgres://user:password@localhost:5432/task_manager?sslmode=disable

# MCP Server
MCP_SERVER_PORT=8080

# JWT
JWT_SECRET=your-secret-key-here

# Logging
LOG_LEVEL=info
```

## Примечания по реализации

### Совместимость с MCP клиентами
- Все инструменты используют `mcp.NewToolResultStructured` с краткими текстовыми сообщениями в content
- Структурированные данные содержат полную информацию для программного доступа
- Текстовые сообщения содержат только суть операции для быстрого понимания
- Полная совместимость с любыми MCP клиентами (структурированные + текстовые данные)

### JWT токен
JWT токен должен содержать следующие claims:
```json
{
  "user_id": "UUID пользователя",
  "is_admin": true/false,
  "exp": "время истечения токена"
}
```

### Обработка ошибок
- Все ошибки должны возвращаться в понятном формате
- Использовать правильные коды ошибок MCP
- Логировать ошибки для отладки

### Валидация
- Проверять UUID формат для ID
- Ограничивать длину строк
- Валидировать enum значения
- Проверять права доступа (только создатель или исполнитель может изменять задачу)
- Имена пользователей должны быть уникальными
- При работе с пользователями использовать имена вместо UUID для удобства

### Безопасность
- Использовать подготовленные запросы для защиты от SQL инъекций
- Не выводить чувствительную информацию в логи
- Проверять JWT токен для каждого запроса

### Производительность
- Использовать индексы для часто используемых полей
- get_next_task должен быть оптимизирован для быстрого получения одной задачи
- Индекс на users.name для быстрого поиска пользователей по имени

## Тестирование каждого этапа

Для каждого этапа создан простой скрипт или использован MCP клиент для проверки функциональности. Документированы примеры запросов и ожидаемые ответы.


## Быстрый старт

### 1. Настройка окружения
```bash
# Скопировать пример конфигурации
cp .env.example .env

# Отредактировать .env файл с вашими настройками
# DATABASE_URL, JWT_SECRET, MCP_SERVER_PORT
```

### 2. Запуск сервера
```bash
# Установить зависимости
go mod tidy

# Запустить сервер
go run main.go
```

### 3. Создание админ пользователя
```bash
# Использовать утилиту создания админа
go run cmd/create-admin/main.go
```



## Доступные MCP инструменты

### create_user (только для админов)
Создание новых пользователей в системе.

### create_task
Создание новой задачи и назначение её пользователю.
- Параметры: `description` (обязательный), `assigned_to` (обязательный)
- Автоматически устанавливает статус "pending"

### list_created_tasks
Получение списка задач, созданных пользователем.
- Параметры: `user_name` (опциональный), `limit` (опциональный, по умолчанию 50, максимум 1000), `statuses` (опциональный)
- Возвращает задачи, созданные текущим пользователем
- Админы могут просматривать задачи других пользователей через параметр `user_name`
- Фильтрация по статусам: pending, in_progress, waiting_for_user, completed, cancelled
- Включает архивные и неархивные задачи с комментариями
- Сортировка по дате создания (новые первые)

### get_next_task
Получение следующей задачи по фильтру статусов.
- Параметры: `statuses` (по умолчанию ["pending"])
- Доступные статусы: pending, in_progress, waiting_for_user, completed, cancelled
- Возвращает задачи где пользователь является исполнителем (assigned_to)
- Включает поле `result` для завершенных задач

### complete_task
Завершение задачи с возможностью добавления результата.
- Параметры: `id` (обязательный), `result` (опциональный)
- Устанавливает статус "completed" и время завершения
- Проверяет права доступа (только создатель или исполнитель)
- Не позволяет завершать уже завершенные или архивные задачи

### cancel_task
Отмена задачи с указанием причины.
- Параметры: `id` (обязательный), `reason` (обязательный)
- Устанавливает статус "cancelled"
- Добавляет причину отмены в конец поля result
- Проверяет права доступа (только создатель или исполнитель)
- Не позволяет отменять уже завершенные или архивные задачи

### wait_for_user
Отправка задачи в ожидание человека с комментарием.
- Параметры: `id` (обязательный), `comment` (обязательный)
- Устанавливает статус "waiting_for_user"
- Добавляет комментарий в таблицу task_comments
- Проверяет права доступа (только создатель или исполнитель)
- Не позволяет отправлять уже завершенные или архивные задачи в ожидание
- get_next_task возвращает комментарии вместе с задачами

### generate_token (только для админов)
Генерация нового JWT токена для существующего пользователя.
- Параметры: `user_id` (обязательный)
- Требует права администратора
- Проверяет существование пользователя
- Возвращает новый JWT токен с годовым сроком действия

### get_token_info
Получение информации о текущем JWT токене.
- Не требует параметров
- Декодирует текущий JWT токен
- Возвращает информацию о пользователе и сроке действия токена
- Показывает оставшееся время до истечения токена

### list_users
Получение списка пользователей с поддержкой лимита.
- Параметры: `limit` (опциональный - число, по умолчанию 100, минимум 1, максимум 1000)
- Возвращает список пользователей с информацией о количестве и лимите
- Включает опциональное поле description
- Сортирует пользователей по имени (name)
- Формат ответа: `{"users": [...], "count": число, "limit": число}`
- Каждый пользователь: `{"id": "uuid", "name": "string", "description": "string?", "is_admin": boolean, "created_at": "ISO-date", "updated_at": "ISO-date"}`
