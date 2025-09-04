# План разработки MCP Task Manager Server

## Текущий статус проекта
**Версия**: 0.1.0  
**Статус**: В разработке  
**Завершенные этапы**: 5 из 9  
**Текущий этап**: Готов к этапу 6 (update_task_status)

### Реализованные возможности:
- ✅ Базовая инфраструктура MCP сервера
- ✅ Подключение к PostgreSQL
- ✅ JWT авторизация
- ✅ Создание пользователей (create_user)
- ✅ Создание задач (create_task)
- ✅ Получение задач по фильтру (get_next_task)
- ✅ Завершение задач с результатом (complete_task)

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
│       └── 002_tasks.sql
├── models/              # Модели данных
│   ├── user.go         # Модель пользователя
│   └── task.go         # Модель задачи
├── auth/               # Авторизация
│   └── jwt.go          # JWT middleware
├── tools/              # MCP инструменты (каждый в отдельном файле)
│   ├── create_user.go     # ✅ Создание пользователей (админ)
│   ├── create_task.go     # ✅ Создание задач
│   ├── get_next_task.go   # ✅ Получение следующей задачи
│   ├── complete_task.go   # ✅ Завершение задач с результатом
│   ├── update_task_status.go  # 🔄 Изменение статуса (в планах)
│   └── archive_task.go    # 🔄 Архивирование (в планах)
├── tests/              # Тестовые скрипты и документация
│   ├── test_connection.sh
│   ├── test_create_task.sh
│   ├── test_get_next_task.sh
│   ├── test_complete_task.sh  # ✅ Тест завершения задач
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
- [x] `get_next_task` - Получение следующей задачи по фильтру статусов
- [x] `complete_task` - Завершение задачи с результатом
- [ ] `update_task_status` - Изменение статуса задачи
- [ ] `archive_task` - Архивирование задачи

### Инфраструктура:
- [ ] Docker контейнеризация
- [ ] GitHub Actions CI/CD

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

### Этап 4: Получение следующей задачи (get_next_task) ✅
**Цель**: Реализовать получение задач по фильтру

**Статус**: Завершен

**MCP инструмент**:

<details>
<summary><b>get_next_task</b> - Получение следующей задачи по фильтру статусов</summary>

```go
getNextTaskTool := mcp.NewTool("get_next_task",
    mcp.WithDescription("Get one task where the current user is either creator or assignee, filtered by status"),
    mcp.WithArray("statuses",
        mcp.Description("Array of statuses to filter by. Available statuses: pending, in_progress, waiting_for_user, completed, cancelled. If not provided, defaults to [\"pending\"]"),
    ),
)
```

**Реализованные возможности**:
- Возвращает одну задачу, где текущий пользователь является либо создателем (created_by), либо исполнителем (assigned_to)
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
- ✅ Возвращает задачи где пользователь создатель или исполнитель
- ✅ Возвращает null если задач нет
- ✅ Использует значение по умолчанию ["pending"]
- ✅ Включает имена и UUID пользователей

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

### Этап 6: Изменение статуса (update_task_status)
**Цель**: Реализовать изменение статуса задач

**MCP инструмент**:

<details>
<summary><b>update_task_status</b> - Изменение статуса задачи</summary>

```go
updateTaskStatusTool := mcp.NewTool("update_task_status",
    mcp.WithDescription("Update task status"),
    mcp.WithString("id",
        mcp.Required(),
        mcp.Description("Task ID (UUID)"),
    ),
    mcp.WithString("status",
        mcp.Required(),
        mcp.Description("New status"),
        mcp.Enum("pending", "in_progress", "waiting_for_user", "completed", "cancelled"),
    ),
)
```
</details>

**Проверка**:
- Проверяет права доступа
- Обновляет статус
- Устанавливает completed_at при переходе в completed
- Не позволяет менять статус архивных задач

### Этап 7: Архивирование задач (archive_task)
**Цель**: Реализовать архивирование задач

**MCP инструмент**:

<details>
<summary><b>archive_task</b> - Архивирование задачи</summary>

```go
archiveTaskTool := mcp.NewTool("archive_task",
    mcp.WithDescription("Archive a task (soft delete)"),
    mcp.WithString("id",
        mcp.Required(),
        mcp.Description("Task ID (UUID)"),
    ),
)
```
</details>

**Проверка**:
- Проверяет права доступа
- Устанавливает is_archived=true
- Устанавливает archived_at
- Не позволяет архивировать уже архивные задачи

### Этап 8: Docker контейнеризация
**Цель**: Создать Docker образ для удобного развертывания

**Задачи**:
1. Создать Dockerfile для сборки приложения
2. Добавить .dockerignore
3. Оптимизировать образ (multi-stage build)

**Файлы**:

<details>
<summary><b>Dockerfile</b> - Multi-stage build для оптимального размера</summary>

```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder
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
COPY --from=builder /app/.env.example .env.example
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

**Проверка**:
- Docker образ собирается успешно
- Размер образа минимальный (alpine + бинарник)
- Приложение запускается из образа
- Переменные окружения передаются корректно

### Этап 9: GitHub Actions CI/CD
**Цель**: Настроить автоматическую сборку и публикацию Docker образа в Docker Hub

**Задачи**:
1. Создать workflow для автоматической сборки
2. Настроить публикацию в Docker Hub при push в main
3. Настроить правильное тегирование образов

**Файлы**:

<details>
<summary><b>.github/workflows/docker-publish.yml</b> - Автоматическая сборка и публикация</summary>

```yaml
name: Docker Build and Push

on:
  push:
    branches: [ main ]
    tags:
      - 'v*'

env:
  DOCKER_USERNAME: ${{ secrets.DOCKER_USERNAME }}
  IMAGE_NAME: ${{ secrets.DOCKER_USERNAME }}/simple-task-mcp

jobs:
  build-and-push:
    runs-on: ubuntu-latest
    
    steps:
    - name: Checkout repository
      uses: actions/checkout@v3

    - name: Log in to Docker Hub
      uses: docker/login-action@v2
      with:
        username: ${{ secrets.DOCKER_USERNAME }}
        password: ${{ secrets.DOCKER_PASSWORD }}

    - name: Extract metadata
      id: meta
      uses: docker/metadata-action@v4
      with:
        images: ${{ env.IMAGE_NAME }}
        tags: |
          type=ref,event=branch
          type=semver,pattern={{version}}
          type=semver,pattern={{major}}.{{minor}}
          type=raw,value=latest,enable={{is_default_branch}}

    - name: Build and push Docker image
      uses: docker/build-push-action@v4
      with:
        context: .
        push: true
        tags: ${{ steps.meta.outputs.tags }}
        labels: ${{ steps.meta.outputs.labels }}
```
</details>

**Необходимые секреты в GitHub**:
- `DOCKER_USERNAME` - имя пользователя Docker Hub
- `DOCKER_PASSWORD` - токен доступа Docker Hub

**Проверка**:
- Push в main запускает автоматическую сборку
- Docker образ публикуется в Docker Hub
- Образ доступен по тегу latest для main ветки
- При создании git тега v* образ публикуется с соответствующей версией



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

### Тестовые файлы:
- `tests/test_connection.sh` - Тест подключения к серверу
- `tests/test_create_task.sh` - Тест создания задач
- `tests/test_get_next_task.sh` - Тест получения задач
- `tests/TESTING_STAGE_3.md` - Документация по тестированию этапа 3
- `tests/TESTING_STAGE_4.md` - Документация по тестированию этапа 4

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

### 4. Тестирование
```bash
# Запустить тесты
./tests/test_connection.sh
./tests/test_create_task.sh
./tests/test_get_next_task.sh
```

## Доступные MCP инструменты

### create_user (только для админов)
Создание новых пользователей в системе.

### create_task
Создание новой задачи и назначение её пользователю.
- Параметры: `description` (обязательный), `assigned_to` (обязательный)
- Автоматически устанавливает статус "pending"

### get_next_task
Получение следующей задачи по фильтру статусов.
- Параметры: `statuses` (по умолчанию ["pending"])
- Доступные статусы: pending, in_progress, waiting_for_user, completed, cancelled
- Возвращает задачи где пользователь создатель или исполнитель
- Включает поле `result` для завершенных задач

### complete_task
Завершение задачи с возможностью добавления результата.
- Параметры: `id` (обязательный), `result` (опциональный)
- Устанавливает статус "completed" и время завершения
- Проверяет права доступа (только создатель или исполнитель)
- Не позволяет завершать уже завершенные или архивные задачи
