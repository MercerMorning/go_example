# Тесты для User API

Этот каталог содержит тесты для User API, включая unit-тесты и integration-тесты.

## Структура тестов

- `create_test.go` - Unit-тесты с моками
- `integration_test.go` - Integration-тесты с реальной БД

## Запуск тестов

### Unit-тесты (с моками)
```bash
make test-unit
# или
go test -v ./internal/api/user/tests/create_test.go
```

### Integration-тесты (с реальной БД)

#### С локальной БД
```bash
# Убедитесь, что PostgreSQL запущен на порту 54321
make test-integration-local
```

#### С Docker контейнером
```bash
# Автоматически создает и удаляет тестовую БД
make test-integration
```

#### Все тесты
```bash
make test-all
```

## Настройка

### Переменные окружения

- `TEST_PG_DSN` - DSN для тестовой БД
- `SKIP_INTEGRATION_TESTS` - пропустить integration тесты (для CI/CD)

### Пример настройки
```bash
export TEST_PG_DSN="postgres://postgres:postgres@localhost:54321/auth_test?sslmode=disable"
```

## Что тестируется

### Unit-тесты
- Логика API слоя
- Валидация входных данных
- Обработка ошибок
- Моки сервисов

### Integration-тесты
- Полный цикл: API → Service → Repository → Database
- Создание пользователей
- Обработка дублирующихся email
- Валидация паролей
- Различные роли пользователей
- Производительность

## Требования

- Go 1.19+
- PostgreSQL 14+
- Docker (для автоматических тестов)
- testify (автоматически устанавливается)
