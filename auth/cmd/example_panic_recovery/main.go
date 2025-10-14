package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/MercerMorning/go_example/auth/internal/config"
	"github.com/MercerMorning/go_example/auth/internal/logger"
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

	fmt.Println("=== Демонстрация перехвата паник и ошибок ===")

	// 1. Обычная ошибка через логгер
	fmt.Println("\n1. Обычная ошибка через логгер:")
	logger.Error("Database connection failed",
		zap.String("database", "users"),
		zap.String("error", "connection timeout"),
	)

	// 2. Паника с перехватом (приложение продолжит работу)
	fmt.Println("\n2. Паника с перехватом (silent):")
	logger.WithPanicRecoverySilent(func() {
		panic("Something went wrong in business logic!")
	})
	fmt.Println("Приложение продолжает работу после паники")

	// 3. Паника с перехватом (приложение завершится)
	fmt.Println("\n3. Паника с перехватом (re-panic):")
	// Раскомментируйте следующую строку для демонстрации re-panic
	// logger.WithPanicRecovery(func() {
	// 	panic("Critical error - application will terminate!")
	// })

	// 4. Ошибка в горутине
	fmt.Println("\n4. Ошибка в горутине:")
	go func() {
		defer logger.RecoverPanicSilent()

		// Симуляция работы
		time.Sleep(100 * time.Millisecond)

		// Ошибка в горутине
		logger.Error("Error in goroutine",
			zap.String("goroutine", "worker"),
			zap.String("error", "processing failed"),
		)
	}()

	// 5. Паника в горутине
	fmt.Println("\n5. Паника в горутине:")
	go func() {
		defer logger.RecoverPanicSilent()

		// Симуляция работы
		time.Sleep(200 * time.Millisecond)

		// Паника в горутине
		panic("Panic in goroutine!")
	}()

	// 6. Необработанная ошибка (НЕ будет автоматически отправлена в Sentry)
	fmt.Println("\n6. Необработанная ошибка (требует ручного логирования):")
	unhandledErr := errors.New("unhandled error")
	// Эта ошибка НЕ будет отправлена в Sentry автоматически!
	// Нужно явно залогировать:
	logger.Error("Unhandled error occurred",
		zap.String("error", unhandledErr.Error()),
		zap.String("component", "business_logic"),
	)

	// 7. Прямая отправка ошибки в Sentry
	fmt.Println("\n7. Прямая отправка ошибки в Sentry:")
	customError := errors.New("custom business logic error")
	logger.CaptureError(customError, map[string]string{
		"component": "user_service",
		"operation": "create_user",
	}, map[string]interface{}{
		"user_id": "12345",
		"email":   "user@example.com",
	})

	// Ждем завершения горутин
	time.Sleep(1 * time.Second)

	fmt.Println("\n=== Демонстрация завершена ===")
	fmt.Println("Проверьте Sentry dashboard для просмотра отправленных событий")
}
