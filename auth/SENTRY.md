# Sentry Integration

Этот проект интегрирован с Sentry для мониторинга ошибок и производительности.

## Настройка

### 1. Переменные окружения

Создайте файл `.env` в корне проекта со следующими переменными:

```bash
# Sentry Configuration
SENTRY_DSN=https://your-dsn@sentry.io/project-id
SENTRY_ENVIRONMENT=development
SENTRY_RELEASE=1.0.0
SENTRY_DEBUG=false
```

### 2. Получение DSN

1. Зарегистрируйтесь на [sentry.io](https://sentry.io)
2. Создайте новый проект
3. Скопируйте DSN из настроек проекта
4. Вставьте DSN в переменную `SENTRY_DSN`

## Использование

### Автоматическая интеграция

Sentry автоматически интегрирован с логгером Zap. Все ошибки уровня `Error` и выше будут автоматически отправляться в Sentry.

### Перехват паник

Sentry также может перехватывать паники и отправлять их в Sentry:

```go
// Перехват паники с продолжением работы приложения
defer logger.RecoverPanicSilent()

// Перехват паники с завершением приложения
defer logger.RecoverPanic()

// Обертка для функций
logger.WithPanicRecoverySilent(func() {
    // код, который может паниковать
})
```

```go
// Обычное логирование - будет отправлено в Sentry как breadcrumb
logger.Info("User logged in", zap.String("user_id", "123"))

// Ошибка - будет отправлена в Sentry как exception
logger.Error("Database connection failed", zap.String("error", "timeout"))
```

### Прямая отправка в Sentry

```go
// Отправка ошибки
err := errors.New("custom error")
logger.CaptureError(err, map[string]string{
    "component": "user_service",
}, map[string]interface{}{
    "user_id": "123",
})

// Отправка сообщения
logger.CaptureMessage("Important event", sentry.LevelInfo, 
    map[string]string{"event": "user_registration"},
    map[string]interface{}{"user_id": "123"})
```

### Транзакции для отслеживания производительности

```go
ctx := context.Background()
transaction := logger.StartTransaction(ctx, "user_creation", "user.create")
defer transaction.Finish()

// Создание span внутри транзакции
span := transaction.StartChild("database.insert")
// ... выполнение операции
span.Finish()
```

## Примеры

Запустите пример использования:

```bash
# Основной пример
go run cmd/example_sentry/main.go

# Пример с перехватом паник
go run cmd/example_panic_recovery/main.go
```

## Что логируется автоматически

### ✅ Автоматически отправляется в Sentry:

1. **Ошибки через логгер** - все вызовы `logger.Error()`, `logger.Fatal()`
2. **Паники** - если используется `defer logger.RecoverPanic()` или `defer logger.RecoverPanicSilent()`
3. **Breadcrumbs** - Info и Warn логи для контекста
4. **Ошибки в горутинах** - если добавлен `defer logger.RecoverPanicSilent()`

### ❌ НЕ отправляется автоматически:

1. **Необработанные ошибки** - если вы просто возвращаете `error` без логирования
2. **Паники без перехвата** - если не используется `defer logger.RecoverPanic()`
3. **Ошибки в горутинах без перехвата** - если не добавлен `defer logger.RecoverPanicSilent()`

### Рекомендации:

```go
// ✅ Хорошо - ошибка будет отправлена в Sentry
if err != nil {
    logger.Error("Operation failed", zap.Error(err))
    return err
}

// ❌ Плохо - ошибка НЕ будет отправлена в Sentry
if err != nil {
    return err // Только возвращаем ошибку
}

// ✅ Хорошо - паника будет перехвачена
defer logger.RecoverPanicSilent()
riskyOperation()

// ❌ Плохо - паника НЕ будет перехвачена
riskyOperation() // Может упасть без логирования
```

## Конфигурация

### Уровни логирования

- `Debug` - не отправляется в Sentry
- `Info` - отправляется как breadcrumb
- `Warn` - отправляется как breadcrumb
- `Error` - отправляется как exception
- `Fatal` - отправляется как exception

### Настройки производительности

- `SampleRate` - процент событий для отправки (по умолчанию 1.0)
- `TracesSampleRate` - процент транзакций для отправки (по умолчанию 0.1)

## Структура файлов

```
internal/
├── config/
│   └── sentry.go          # Конфигурация Sentry
├── logger/
│   ├── logger.go          # Основной логгер
│   └── sentry.go          # Интеграция с Sentry
└── app/
    └── app.go             # Инициализация в приложении

cmd/
└── example_sentry/
    └── main.go            # Пример использования
```

## Отключение Sentry

Для отключения Sentry просто не устанавливайте переменную `SENTRY_DSN` или оставьте её пустой. Приложение будет работать без Sentry интеграции.
