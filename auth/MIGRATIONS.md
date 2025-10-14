# Миграции базы данных

Этот проект использует `golang-migrate` для управления миграциями базы данных PostgreSQL.

## Структура

- `migrations/` - директория с файлами миграций
- `cmd/migrate/` - CLI утилита для управления миграциями
- `internal/client/db/migrate/` - пакет для работы с миграциями

## Команды

### Запуск миграций

```bash
# Применить все миграции
make migrate-up

# Откатить все миграции
make migrate-down

# Проверить текущую версию миграции
make migrate-version

# Принудительно установить версию миграции (используется при проблемах)
make migrate-force

# Создать новую миграцию
make migrate-create
```

### Ручной запуск

```bash
# Установить переменную окружения
export PG_DSN="postgres://postgres:postgres@localhost:54321/auth?sslmode=disable"

# Запустить миграции
cd cmd/migrate && go run main.go -command=up

# Откатить миграции
cd cmd/migrate && go run main.go -command=down

# Проверить версию
cd cmd/migrate && go run main.go -command=version

# Принудительно установить версию
cd cmd/migrate && go run main.go -command=force -version=20251014162220
```

## Создание новых миграций

1. Запустите команду:
   ```bash
   make migrate-create
   ```

2. Введите название миграции (например: `add_user_phone_field`)

3. Отредактируйте созданные файлы:
   - `migrations/YYYYMMDDHHMMSS_add_user_phone_field.up.sql` - для применения изменений
   - `migrations/YYYYMMDDHHMMSS_add_user_phone_field.down.sql` - для отката изменений

## Формат файлов миграций

### Up миграция (применение)
```sql
-- +migrate Up
CREATE TABLE example (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL
);
```

### Down миграция (откат)
```sql
-- +migrate Down
DROP TABLE example;
```

## Требования

- PostgreSQL 14+
- Go 1.24+
- Docker и Docker Compose (для локальной разработки)

## Переменные окружения

- `PG_DSN` - строка подключения к PostgreSQL

Пример:
```
PG_DSN=postgres://postgres:postgres@localhost:54321/auth?sslmode=disable
```

## Существующие миграции

### 20251014162220_create_users_table
Создает таблицу `users` с полями:
- `id` - первичный ключ (BIGSERIAL)
- `name` - имя пользователя (VARCHAR(255))
- `email` - email пользователя (VARCHAR(255), UNIQUE)
- `password` - пароль (VARCHAR(255))
- `role` - роль пользователя (VARCHAR(50), по умолчанию 'user')
- `created_at` - время создания (TIMESTAMP WITH TIME ZONE)
- `updated_at` - время обновления (TIMESTAMP WITH TIME ZONE)

Индексы:
- `idx_users_email` - по полю email
- `idx_users_role` - по полю role
- `idx_users_created_at` - по полю created_at
