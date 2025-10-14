package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/MercerMorning/go_example/auth/internal/config"
	"github.com/MercerMorning/go_example/auth/internal/monitoring"
	"github.com/getsentry/sentry-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func main() {
	fmt.Println("=== Enterprise Sentry Integration Example ===")

	// Загружаем конфигурацию
	err := config.Load(".env")
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		return
	}

	// Создаем enterprise конфигурацию Sentry
	sentryConfig := monitoring.LoadSentryConfigFromEnv()
	if sentryConfig == nil {
		fmt.Println("Sentry is disabled (no DSN provided)")
		return
	}

	// Инициализируем enterprise Sentry
	sentryEnterprise := monitoring.NewSentryEnterprise(sentryConfig)
	err = sentryEnterprise.Init()
	if err != nil {
		fmt.Printf("Failed to init Sentry: %v\n", err)
		return
	}
	defer sentryEnterprise.Flush()

	// Создаем middleware
	middleware := monitoring.NewSentryMiddleware(sentryEnterprise)

	// Демонстрируем различные сценарии
	demonstrateEnterpriseFeatures(sentryEnterprise, middleware)

	fmt.Println("\n=== Enterprise example completed ===")
}

func demonstrateEnterpriseFeatures(enterprise *monitoring.SentryEnterprise, middleware *monitoring.SentryMiddleware) {
	ctx := context.Background()

	// 1. Бизнес-события
	fmt.Println("\n1. Business Events:")
	enterprise.CaptureBusinessEvent(
		"user_registration",
		"user_123",
		map[string]interface{}{
			"plan":        "premium",
			"source":      "website",
			"campaign_id": "summer2024",
		},
	)

	// 2. Производительность
	fmt.Println("\n2. Performance Monitoring:")
	// Симуляция медленной операции
	start := time.Now()
	time.Sleep(150 * time.Millisecond) // Медленная операция
	duration := time.Since(start)

	enterprise.CapturePerformanceIssue(
		"user_creation",
		duration,
		100*time.Millisecond, // threshold
		map[string]interface{}{
			"user_id":   "user_123",
			"operation": "create_user",
		},
	)

	// 3. Ошибки с контекстом
	fmt.Println("\n3. Errors with Context:")
	user := &sentry.User{
		ID:    "user_123",
		Email: "user@example.com",
	}

	enterprise.CaptureErrorWithContext(
		ctx,
		errors.New("payment processing failed"),
		map[string]string{
			"component": "payment_service",
			"operation": "process_payment",
		},
		map[string]interface{}{
			"amount":     99.99,
			"currency":   "USD",
			"payment_id": "pay_123",
		},
		user,
	)

	// 4. Транзакции
	fmt.Println("\n4. Transactions:")
	transaction := enterprise.StartTransactionWithContext(
		ctx,
		"user_onboarding",
		"user.create",
		map[string]string{
			"user_id": "user_123",
		},
	)
	defer transaction.Finish()

	// Создаем spans внутри транзакции
	span1 := enterprise.CreateSpan(
		transaction,
		"database.insert",
		"Insert user into database",
		map[string]string{
			"table": "users",
		},
	)
	time.Sleep(50 * time.Millisecond) // Симуляция работы
	span1.Finish()

	span2 := enterprise.CreateSpan(
		transaction,
		"email.send",
		"Send welcome email",
		map[string]string{
			"template": "welcome",
		},
	)
	time.Sleep(30 * time.Millisecond) // Симуляция работы
	span2.Finish()

	// 5. Database middleware
	fmt.Println("\n5. Database Middleware:")
	dbMiddleware := middleware.DatabaseMiddleware()

	// Симуляция быстрого запроса
	dbMiddleware(func() error {
		time.Sleep(50 * time.Millisecond)
		return nil
	})

	// Симуляция медленного запроса
	dbMiddleware(func() error {
		time.Sleep(1500 * time.Millisecond) // Медленный запрос
		return nil
	})

	// Симуляция ошибки БД
	dbMiddleware(func() error {
		return errors.New("database connection timeout")
	})

	// 6. Business logic middleware
	fmt.Println("\n6. Business Logic Middleware:")
	businessMiddleware := middleware.BusinessLogicMiddleware("user_validation")

	businessMiddleware(func() error {
		time.Sleep(100 * time.Millisecond)
		return nil
	})

	businessMiddleware(func() error {
		return errors.New("user validation failed: invalid email")
	})

	// 7. gRPC interceptor simulation
	fmt.Println("\n7. gRPC Interceptor Simulation:")
	grpcInterceptor := middleware.GRPCUnaryInterceptor()

	// Симуляция успешного gRPC запроса
	grpcInterceptor(ctx, "request", &grpc.UnaryServerInfo{
		FullMethod: "/user.UserService/CreateUser",
	}, func(ctx context.Context, req interface{}) (interface{}, error) {
		time.Sleep(80 * time.Millisecond)
		return "success", nil
	})

	// Симуляция ошибки gRPC
	grpcInterceptor(ctx, "request", &grpc.UnaryServerInfo{
		FullMethod: "/user.UserService/GetUser",
	}, func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, status.Error(codes.Internal, "internal server error")
	})

	// 8. HTTP middleware simulation
	fmt.Println("\n8. HTTP Middleware Simulation:")
	httpMiddleware := middleware.HTTPMiddleware()

	// Создаем тестовый HTTP handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Оборачиваем в Sentry middleware
	_ = httpMiddleware(testHandler)

	// Симуляция HTTP запроса
	req, _ := http.NewRequest("GET", "/api/users", nil)
	req = req.WithContext(ctx)

	// В реальном приложении это будет обработано HTTP сервером
	fmt.Printf("HTTP middleware configured for: %s\n", req.URL.Path)

	// 9. Множественные ошибки (тест rate limiting)
	fmt.Println("\n9. Multiple Errors (Rate Limiting Test):")
	for i := 0; i < 5; i++ {
		enterprise.CaptureErrorWithContext(
			ctx,
			errors.New("rate limited error"),
			map[string]string{
				"component": "rate_limiter",
				"iteration": fmt.Sprintf("%d", i),
			},
			map[string]interface{}{
				"timestamp": time.Now(),
			},
			nil,
		)
		time.Sleep(10 * time.Millisecond)
	}

	fmt.Println("\nAll enterprise features demonstrated!")
}
