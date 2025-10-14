package logger

import (
	"context"
	"time"

	"github.com/getsentry/sentry-go"
	"go.uber.org/zap/zapcore"
)

// SentryCore - кастомный core для интеграции с Sentry
type SentryCore struct {
	zapcore.Core
	level zapcore.Level
}

// NewSentryCore создает новый Sentry core
func NewSentryCore(core zapcore.Core, level zapcore.Level) *SentryCore {
	return &SentryCore{
		Core:  core,
		level: level,
	}
}

// Write отправляет логи в Sentry
func (c *SentryCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	// Сначала записываем в оригинальный core
	if err := c.Core.Write(entry, fields); err != nil {
		return err
	}

	// Если уровень лога подходящий, отправляем в Sentry
	if entry.Level >= c.level {
		c.sendToSentry(entry, fields)
	}

	return nil
}

// sendToSentry отправляет лог в Sentry
func (c *SentryCore) sendToSentry(entry zapcore.Entry, fields []zapcore.Field) {
	// Создаем scope для Sentry
	scope := sentry.NewScope()

	// Добавляем поля как теги и контекст
	for _, field := range fields {
		switch field.Type {
		case zapcore.StringType:
			scope.SetTag(field.Key, field.String)
		case zapcore.Int64Type:
			scope.SetTag(field.Key, string(rune(field.Integer)))
		case zapcore.BoolType:
			scope.SetTag(field.Key, string(rune(field.Integer)))
		default:
			scope.SetContext(field.Key, map[string]interface{}{"value": field.Interface})
		}
	}

	// Устанавливаем уровень
	level := sentry.LevelInfo
	switch entry.Level {
	case zapcore.DebugLevel:
		level = sentry.LevelDebug
	case zapcore.InfoLevel:
		level = sentry.LevelInfo
	case zapcore.WarnLevel:
		level = sentry.LevelWarning
	case zapcore.ErrorLevel:
		level = sentry.LevelError
	case zapcore.FatalLevel, zapcore.PanicLevel:
		level = sentry.LevelFatal
	}

	// Отправляем сообщение в Sentry
	if entry.Level >= zapcore.ErrorLevel {
		// Для ошибок создаем exception
		sentry.WithScope(func(scope *sentry.Scope) {
			scope.SetLevel(level)
			scope.SetTag("logger", "zap")
			scope.SetContext("log_entry", map[string]interface{}{
				"message": entry.Message,
				"time":    entry.Time,
				"caller":  entry.Caller.String(),
			})

			// Добавляем поля
			for _, field := range fields {
				scope.SetContext(field.Key, map[string]interface{}{"value": field.Interface})
			}

			sentry.CaptureException(&sentryException{
				message: entry.Message,
				time:    entry.Time,
			})
		})
	} else {
		// Для остальных уровней отправляем как breadcrumb
		sentry.AddBreadcrumb(&sentry.Breadcrumb{
			Message:   entry.Message,
			Level:     level,
			Timestamp: entry.Time,
			Data: map[string]interface{}{
				"logger": entry.LoggerName,
				"caller": entry.Caller.String(),
			},
		})
	}
}

// sentryException - кастомный тип исключения для Sentry
type sentryException struct {
	message string
	time    time.Time
}

func (e *sentryException) Error() string {
	return e.message
}

// InitSentry инициализирует Sentry
func InitSentry(dsn, environment, release string, debug bool, sampleRate, tracesSampleRate float64) error {
	return sentry.Init(sentry.ClientOptions{
		Dsn:              dsn,
		Environment:      environment,
		Release:          release,
		Debug:            debug,
		SampleRate:       sampleRate,
		TracesSampleRate: tracesSampleRate,
		BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
			// Можно добавить фильтрацию событий здесь
			return event
		},
	})
}

// FlushSentry принудительно отправляет все события в Sentry
func FlushSentry(timeout time.Duration) bool {
	return sentry.Flush(timeout)
}

// CaptureError отправляет ошибку в Sentry
func CaptureError(err error, tags map[string]string, context map[string]interface{}) {
	sentry.WithScope(func(scope *sentry.Scope) {
		// Добавляем теги
		for key, value := range tags {
			scope.SetTag(key, value)
		}

		// Добавляем контекст
		for key, value := range context {
			scope.SetContext(key, map[string]interface{}{"value": value})
		}

		sentry.CaptureException(err)
	})
}

// CaptureMessage отправляет сообщение в Sentry
func CaptureMessage(message string, level sentry.Level, tags map[string]string, context map[string]interface{}) {
	sentry.WithScope(func(scope *sentry.Scope) {
		scope.SetLevel(level)

		// Добавляем теги
		for key, value := range tags {
			scope.SetTag(key, value)
		}

		// Добавляем контекст
		for key, value := range context {
			scope.SetContext(key, map[string]interface{}{"value": value})
		}

		sentry.CaptureMessage(message)
	})
}

// StartTransaction создает новую транзакцию для Sentry
func StartTransaction(ctx context.Context, name, operation string) *sentry.Span {
	transaction := sentry.StartTransaction(ctx, name)
	return transaction
}
