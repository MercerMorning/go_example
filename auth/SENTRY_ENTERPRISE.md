# Sentry Enterprise Integration

Руководство по интеграции Sentry в большие проекты с enterprise-функциями.

## 🏗️ Архитектура для больших проектов

### 1. Многоуровневая структура

```
internal/monitoring/
├── sentry_enterprise.go      # Основная enterprise логика
├── sentry_middleware.go      # Middleware для HTTP/gRPC
├── sentry_configs.go         # Конфигурации для окружений
└── sentry_metrics.go         # Метрики и аналитика
```

### 2. Конфигурация по окружениям

```go
// Разные настройки для разных окружений
development: 100% sampling, debug включен
staging:     50% sampling, фильтрация debug
production:  10% sampling, агрессивная фильтрация
```

## 🚀 Enterprise функции

### 1. Расширенный контекст ошибок

```go
enterprise.CaptureErrorWithContext(
    ctx,
    err,
    map[string]string{
        "component": "payment_service",
        "operation": "process_payment",
    },
    map[string]interface{}{
        "amount": 99.99,
        "currency": "USD",
    },
    &sentry.User{ID: "user_123"},
)
```

### 2. Мониторинг производительности

```go
// Автоматическое отслеживание медленных операций
enterprise.CapturePerformanceIssue(
    "user_creation",
    duration,
    100*time.Millisecond, // threshold
    context,
)
```

### 3. Бизнес-события

```go
enterprise.CaptureBusinessEvent(
    "user_registration",
    "user_123",
    map[string]interface{}{
        "plan": "premium",
        "source": "website",
    },
)
```

### 4. Транзакции с контекстом

```go
transaction := enterprise.StartTransactionWithContext(
    ctx,
    "user_onboarding",
    "user.create",
    map[string]string{"user_id": "user_123"},
)
defer transaction.Finish()

span := enterprise.CreateSpan(
    transaction,
    "database.insert",
    "Insert user",
    map[string]string{"table": "users"},
)
defer span.Finish()
```

## 🔧 Middleware интеграция

### 1. HTTP Middleware

```go
// Автоматический перехват HTTP ошибок
httpMiddleware := middleware.HTTPMiddleware()
http.Handle("/api/", httpMiddleware(apiHandler))
```

### 2. gRPC Interceptors

```go
// Unary interceptor
grpcServer := grpc.NewServer(
    grpc.UnaryInterceptor(middleware.GRPCUnaryInterceptor()),
    grpc.StreamInterceptor(middleware.GRPCStreamInterceptor()),
)
```

### 3. Database Middleware

```go
dbMiddleware := middleware.DatabaseMiddleware()

err := dbMiddleware(func() error {
    return db.Query("SELECT * FROM users")
})
```

### 4. Business Logic Middleware

```go
businessMiddleware := middleware.BusinessLogicMiddleware("user_validation")

err := businessMiddleware(func() error {
    return validateUser(user)
})
```

## 📊 Конфигурация для больших проектов

### 1. Переменные окружения

```bash
# Основные настройки
SENTRY_DSN=https://your-dsn@sentry.io/project-id
SENTRY_ENVIRONMENT=production
SENTRY_RELEASE=1.2.3
SENTRY_SERVER_NAME=api-server-01

# Производительность
SENTRY_SAMPLE_RATE=0.1
SENTRY_TRACES_SAMPLE_RATE=0.05
SENTRY_MAX_BREADCRUMBS=100
SENTRY_MAX_SPANS=1000

# Дополнительные теги
SENTRY_ADDITIONAL_TAGS=datacenter:us-east-1,team:backend
```

### 2. Фильтрация событий

```go
BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
    // Фильтрация в зависимости от окружения
    if event.Level == sentry.LevelDebug {
        return nil // Игнорируем debug в production
    }
    
    // Фильтрация частых ошибок
    if isFrequentError(event) {
        return nil
    }
    
    return event
}
```

### 3. Rate Limiting

```go
// Автоматическое ограничение частоты отправки
// Настраивается в Sentry dashboard
```

## 🏢 Паттерны для больших проектов

### 1. Микросервисы

```go
// Каждый сервис имеет свою конфигурацию
type ServiceConfig struct {
    Name        string
    Environment string
    DSN         string
    Tags        map[string]string
}

// Инициализация для каждого сервиса
func InitServiceSentry(config ServiceConfig) *SentryEnterprise {
    enterpriseConfig := &SentryEnterpriseConfig{
        DSN: config.DSN,
        Environment: config.Environment,
        Tags: map[string]string{
            "service": config.Name,
            "version": getVersion(),
        },
    }
    
    return NewSentryEnterprise(enterpriseConfig)
}
```

### 2. Distributed Tracing

```go
// Передача trace context между сервисами
func CallExternalService(ctx context.Context, url string) error {
    // Создаем span для внешнего вызова
    span := sentry.StartSpan(ctx, "http.client")
    span.SetTag("url", url)
    defer span.Finish()
    
    // Передаем trace context в HTTP заголовках
    req, _ := http.NewRequestWithContext(span.Context(), "GET", url, nil)
    
    client := &http.Client{}
    resp, err := client.Do(req)
    
    if err != nil {
        span.SetTag("error", true)
        span.SetTag("error.message", err.Error())
    }
    
    return err
}
```

### 3. Error Aggregation

```go
// Группировка похожих ошибок
type ErrorAggregator struct {
    errors map[string]*ErrorInfo
    mutex  sync.RWMutex
}

type ErrorInfo struct {
    Count     int
    FirstSeen time.Time
    LastSeen  time.Time
    Samples   []*sentry.Event
}

func (ea *ErrorAggregator) ProcessError(event *sentry.Event) {
    key := generateErrorKey(event)
    
    ea.mutex.Lock()
    defer ea.mutex.Unlock()
    
    if info, exists := ea.errors[key]; exists {
        info.Count++
        info.LastSeen = time.Now()
        if len(info.Samples) < 5 {
            info.Samples = append(info.Samples, event)
        }
    } else {
        ea.errors[key] = &ErrorInfo{
            Count:     1,
            FirstSeen: time.Now(),
            LastSeen:  time.Now(),
            Samples:   []*sentry.Event{event},
        }
    }
}
```

## 📈 Мониторинг и алерты

### 1. Настройка алертов

```yaml
# sentry-alerts.yml
alerts:
  - name: "High Error Rate"
    condition: "error_rate > 5%"
    action: "slack_notification"
    
  - name: "Performance Degradation"
    condition: "avg_response_time > 1s"
    action: "pagerduty_alert"
    
  - name: "New Error Type"
    condition: "new_error_type"
    action: "email_notification"
```

### 2. Дашборды

```go
// Создание кастомных метрик
func TrackCustomMetric(name string, value float64, tags map[string]string) {
    sentry.WithScope(func(scope *sentry.Scope) {
        scope.SetTag("metric_name", name)
        for key, value := range tags {
            scope.SetTag(key, value)
        }
        scope.SetContext("metric", map[string]interface{}{
            "value": value,
            "timestamp": time.Now(),
        })
        sentry.CaptureMessage(fmt.Sprintf("Metric: %s = %f", name, value))
    })
}
```

## 🔒 Безопасность

### 1. Фильтрация чувствительных данных

```go
BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
    // Удаляем чувствительные данные
    if event.Request != nil {
        // Удаляем пароли из заголовков
        if event.Request.Headers != nil {
            delete(event.Request.Headers, "Authorization")
            delete(event.Request.Headers, "X-API-Key")
        }
        
        // Фильтруем данные формы
        if event.Request.Data != nil {
            filterSensitiveData(event.Request.Data)
        }
    }
    
    return event
}
```

### 2. PII Detection

```go
func filterSensitiveData(data interface{}) {
    // Реализация фильтрации PII данных
    // (email, phone, credit card, etc.)
}
```

## 🚀 Развертывание

### 1. Docker

```dockerfile
# Dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o main ./cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /app/.env .
CMD ["./main"]
```

### 2. Kubernetes

```yaml
# k8s-deployment.yml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: auth-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: auth-service
  template:
    metadata:
      labels:
        app: auth-service
    spec:
      containers:
      - name: auth-service
        image: auth-service:latest
        env:
        - name: SENTRY_DSN
          valueFrom:
            secretKeyRef:
              name: sentry-secrets
              key: dsn
        - name: SENTRY_ENVIRONMENT
          value: "production"
        - name: SENTRY_RELEASE
          value: "1.2.3"
```

## 📊 Примеры использования

Запустите enterprise пример:

```bash
# Установите переменные окружения
export SENTRY_DSN="https://your-dsn@sentry.io/project-id"
export SENTRY_ENVIRONMENT="development"

# Запустите пример
go run cmd/example_enterprise_sentry/main.go
```

## 🎯 Лучшие практики

1. **Настройте правильные sample rates** для каждого окружения
2. **Используйте теги** для группировки и фильтрации
3. **Настройте алерты** для критических ошибок
4. **Мониторьте производительность** Sentry интеграции
5. **Регулярно проверяйте** и очищайте неиспользуемые теги
6. **Используйте release tracking** для отслеживания деплоев
7. **Настройте PII фильтрацию** для соответствия GDPR
