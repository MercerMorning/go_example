package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/MercerMorning/go_example/auth/internal/config"
	"github.com/MercerMorning/go_example/auth/internal/logger"
	"github.com/getsentry/sentry-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	// Загружаем конфигурацию
	err := config.Load(".env")
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		return
	}

	// Инициализируем Sentry
	sentryConfig := config.NewSentryConfig()
	if sentryConfig.IsEnabled() {
		err = logger.InitSentry(
			sentryConfig.DSN,
			sentryConfig.Environment,
			sentryConfig.Release,
			sentryConfig.Debug,
			sentryConfig.SampleRate,
			sentryConfig.TracesSampleRate,
		)
		if err != nil {
			fmt.Printf("Failed to init Sentry: %v\n", err)
			return
		}
		defer logger.FlushSentry(2 * time.Second)
	}

	// Инициализируем логгер с Sentry интеграцией
	zapLog, err := zap.NewProduction()
	if err != nil {
		fmt.Printf("Failed to create logger: %v\n", err)
		return
	}

	var core zapcore.Core = zapLog.Core()
	if sentryConfig.IsEnabled() {
		sentryCore := logger.NewSentryCore(zapLog.Core(), zapcore.ErrorLevel)
		core = sentryCore
	}

	logger.Init(core)

	// Примеры использования
	demonstrateSentryFeatures()
}

func demonstrateSentryFeatures() {
	// 1. Обычное логирование (будет отправлено в Sentry как breadcrumb)
	logger.Info("Application started",
		zap.String("version", "1.0.0"),
		zap.String("environment", "development"),
	)

	// 2. Логирование с контекстом
	logger.Info("User action",
		zap.String("user_id", "12345"),
		zap.String("action", "login"),
		zap.String("ip", "192.168.1.1"),
	)

	// 3. Предупреждение
	logger.Warn("Rate limit approaching",
		zap.String("user_id", "12345"),
		zap.Int("requests", 95),
		zap.Int("limit", 100),
	)

	// 4. Ошибка (будет отправлена в Sentry как exception)
	logger.Error("Database connection failed",
		zap.String("database", "users"),
		zap.String("error", "connection timeout"),
		zap.Duration("timeout", 30*time.Second),
	)

	// 5. Прямая отправка ошибки в Sentry
	customError := errors.New("custom business logic error")
	logger.CaptureError(customError, map[string]string{
		"component": "user_service",
		"operation": "create_user",
	}, map[string]interface{}{
		"user_id": "12345",
		"email":   "user@example.com",
	})

	// 6. Прямая отправка сообщения в Sentry
	logger.CaptureMessage("Important business event occurred",
		sentry.LevelInfo,
		map[string]string{
			"event_type": "user_registration",
		}, map[string]interface{}{
			"user_id": "67890",
			"plan":    "premium",
		})

	// 7. Создание транзакции для отслеживания производительности
	ctx := context.Background()
	transaction := logger.StartTransaction(ctx, "user_creation", "user.create")
	defer transaction.Finish()

	// Симуляция работы
	time.Sleep(100 * time.Millisecond)

	// Создание span внутри транзакции
	span := transaction.StartChild("database.insert")
	time.Sleep(50 * time.Millisecond)
	span.Finish()

	// 8. Фатальная ошибка (будет отправлена в Sentry)
	logger.Fatal("Critical system failure",
		zap.String("component", "auth_service"),
		zap.String("reason", "certificate expired"),
	)
}
