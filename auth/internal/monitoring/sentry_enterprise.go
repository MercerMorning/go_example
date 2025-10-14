package monitoring

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/getsentry/sentry-go"
)

// SentryEnterprise - расширенная конфигурация Sentry для больших проектов
type SentryEnterprise struct {
	config *SentryEnterpriseConfig
}

// SentryEnterpriseConfig - конфигурация для enterprise использования
type SentryEnterpriseConfig struct {
	DSN              string
	Environment      string
	Release          string
	Debug            bool
	SampleRate       float64
	TracesSampleRate float64

	// Enterprise настройки
	MaxBreadcrumbs   int
	BeforeSend       func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event
	BeforeBreadcrumb func(breadcrumb *sentry.Breadcrumb, hint *sentry.BreadcrumbHint) *sentry.Breadcrumb

	// Настройки для больших проектов
	ServerName   string
	Dist         string
	Tags         map[string]string
	Integrations []sentry.Integration

	// Настройки производительности
	MaxSpans         int
	MaxTraceFileSize int64
	FlushTimeout     time.Duration
}

// NewSentryEnterprise создает новую enterprise конфигурацию Sentry
func NewSentryEnterprise(config *SentryEnterpriseConfig) *SentryEnterprise {
	return &SentryEnterprise{
		config: config,
	}
}

// Init инициализирует Sentry с enterprise настройками
func (se *SentryEnterprise) Init() error {
	options := sentry.ClientOptions{
		Dsn:              se.config.DSN,
		Environment:      se.config.Environment,
		Release:          se.config.Release,
		Debug:            se.config.Debug,
		SampleRate:       se.config.SampleRate,
		TracesSampleRate: se.config.TracesSampleRate,
		MaxBreadcrumbs:   se.config.MaxBreadcrumbs,
		BeforeSend:       se.config.BeforeSend,
		BeforeBreadcrumb: se.config.BeforeBreadcrumb,
		ServerName:       se.config.ServerName,
		Dist:             se.config.Dist,
		Tags:             se.config.Tags,
		MaxSpans:         se.config.MaxSpans,
	}

	return sentry.Init(options)
}

// CaptureErrorWithContext отправляет ошибку с расширенным контекстом
func (se *SentryEnterprise) CaptureErrorWithContext(
	ctx context.Context,
	err error,
	tags map[string]string,
	extra map[string]interface{},
	user *sentry.User,
) {
	sentry.WithScope(func(scope *sentry.Scope) {
		// Устанавливаем контекст
		scope.SetContext("request", map[string]interface{}{
			"method":  getContextValue(ctx, "method"),
			"url":     getContextValue(ctx, "url"),
			"headers": getContextValue(ctx, "headers"),
		})

		// Устанавливаем пользователя
		if user != nil {
			scope.SetUser(*user)
		}

		// Добавляем теги
		for key, value := range tags {
			scope.SetTag(key, value)
		}

		// Добавляем дополнительную информацию
		for key, value := range extra {
			scope.SetContext(key, map[string]interface{}{"value": value})
		}

		// Добавляем системную информацию
		scope.SetContext("system", map[string]interface{}{
			"go_version": runtime.Version(),
			"os":         runtime.GOOS,
			"arch":       runtime.GOARCH,
			"goroutines": runtime.NumGoroutine(),
		})

		sentry.CaptureException(err)
	})
}

// CapturePerformanceIssue отправляет информацию о проблемах производительности
func (se *SentryEnterprise) CapturePerformanceIssue(
	operation string,
	duration time.Duration,
	threshold time.Duration,
	context map[string]interface{},
) {
	if duration > threshold {
		sentry.WithScope(func(scope *sentry.Scope) {
			scope.SetLevel(sentry.LevelWarning)
			scope.SetTag("issue_type", "performance")
			scope.SetTag("operation", operation)
			scope.SetContext("performance", map[string]interface{}{
				"duration":    duration.String(),
				"threshold":   threshold.String(),
				"exceeded_by": (duration - threshold).String(),
			})

			for key, value := range context {
				scope.SetContext(key, map[string]interface{}{"value": value})
			}

			sentry.CaptureMessage(fmt.Sprintf("Performance issue: %s took %v (threshold: %v)",
				operation, duration, threshold))
		})
	}
}

// CaptureBusinessEvent отправляет бизнес-события
func (se *SentryEnterprise) CaptureBusinessEvent(
	eventType string,
	userID string,
	properties map[string]interface{},
) {
	sentry.WithScope(func(scope *sentry.Scope) {
		scope.SetLevel(sentry.LevelInfo)
		scope.SetTag("event_type", "business")
		scope.SetTag("business_event", eventType)
		scope.SetUser(sentry.User{ID: userID})

		scope.SetContext("business_event", map[string]interface{}{
			"type":       eventType,
			"properties": properties,
		})

		sentry.CaptureMessage(fmt.Sprintf("Business event: %s", eventType))
	})
}

// StartTransactionWithContext создает транзакцию с контекстом
func (se *SentryEnterprise) StartTransactionWithContext(
	ctx context.Context,
	name string,
	operation string,
	tags map[string]string,
) *sentry.Span {
	transaction := sentry.StartTransaction(ctx, name)

	// Добавляем теги к транзакции
	for key, value := range tags {
		transaction.SetTag(key, value)
	}

	return transaction
}

// CreateSpan создает span с контекстом
func (se *SentryEnterprise) CreateSpan(
	parent *sentry.Span,
	operation string,
	description string,
	tags map[string]string,
) *sentry.Span {
	span := parent.StartChild(operation)

	// Добавляем теги к span
	for key, value := range tags {
		span.SetTag(key, value)
	}

	return span
}

// Flush принудительно отправляет все события
func (se *SentryEnterprise) Flush() bool {
	return sentry.Flush(se.config.FlushTimeout)
}

// getContextValue извлекает значение из контекста
func getContextValue(ctx context.Context, key string) interface{} {
	if value := ctx.Value(key); value != nil {
		return value
	}
	return nil
}

// DefaultEnterpriseConfig возвращает конфигурацию по умолчанию для enterprise
func DefaultEnterpriseConfig(dsn, environment, release string) *SentryEnterpriseConfig {
	return &SentryEnterpriseConfig{
		DSN:              dsn,
		Environment:      environment,
		Release:          release,
		Debug:            false,
		SampleRate:       1.0,
		TracesSampleRate: 0.1,
		MaxBreadcrumbs:   100,
		MaxSpans:         1000,
		MaxTraceFileSize: 10 * 1024 * 1024, // 10MB
		FlushTimeout:     2 * time.Second,
		Tags: map[string]string{
			"service": "auth-service",
			"version": release,
		},
		BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
			// Фильтрация событий для больших проектов
			if event.Level == sentry.LevelDebug {
				return nil // Игнорируем debug события
			}

			// Добавляем дополнительную информацию
			event.ServerName = getServerName()
			event.Tags["processed_at"] = time.Now().Format(time.RFC3339)

			return event
		},
		BeforeBreadcrumb: func(breadcrumb *sentry.Breadcrumb, hint *sentry.BreadcrumbHint) *sentry.Breadcrumb {
			// Фильтрация breadcrumbs
			if breadcrumb.Level == sentry.LevelDebug {
				return nil
			}
			return breadcrumb
		},
	}
}

// getServerName возвращает имя сервера
func getServerName() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}
