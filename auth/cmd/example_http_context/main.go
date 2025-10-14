package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/MercerMorning/go_example/auth/internal/config"
	"github.com/MercerMorning/go_example/auth/internal/monitoring"
	"github.com/getsentry/sentry-go"
)

func main() {
	fmt.Println("=== HTTP Context Auto-Population Example ===")

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

	// Создаем HTTP сервер с Sentry middleware
	mux := http.NewServeMux()

	// Оборачиваем все обработчики в Sentry middleware
	httpMiddleware := middleware.HTTPMiddleware()

	// Обработчики для демонстрации
	mux.Handle("/api/users", httpMiddleware(http.HandlerFunc(usersHandler)))
	mux.Handle("/api/orders", httpMiddleware(http.HandlerFunc(ordersHandler)))
	mux.Handle("/api/error", httpMiddleware(http.HandlerFunc(errorHandler)))
	mux.Handle("/api/panic", httpMiddleware(http.HandlerFunc(panicHandler)))

	// Запускаем сервер
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	fmt.Println("HTTP server started on :8080")
	fmt.Println("Test endpoints:")
	fmt.Println("  GET  /api/users  - Success response")
	fmt.Println("  GET  /api/orders - Success with user context")
	fmt.Println("  GET  /api/error  - Error response")
	fmt.Println("  GET  /api/panic  - Panic (will be caught)")
	fmt.Println("\nTry these URLs:")
	fmt.Println("  curl http://localhost:8080/api/users")
	fmt.Println("  curl -H 'X-User-ID: user123' http://localhost:8080/api/orders")
	fmt.Println("  curl http://localhost:8080/api/error")
	fmt.Println("  curl http://localhost:8080/api/panic")

	// Запускаем сервер в горутине
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Server error: %v\n", err)
		}
	}()

	// Ждем немного для демонстрации
	time.Sleep(30 * time.Second)

	fmt.Println("\nShutting down server...")
	server.Close()
}

// usersHandler - обработчик для демонстрации успешного ответа
func usersHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Processing request: %s %s\n", r.Method, r.URL.Path)

	// Логируем информацию о запросе (будет отправлено в Sentry как breadcrumb)
	sentry.AddBreadcrumb(&sentry.Breadcrumb{
		Message:   "Processing users request",
		Level:     sentry.LevelInfo,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"endpoint": "/api/users",
			"method":   r.Method,
		},
	})

	// Симуляция работы
	time.Sleep(100 * time.Millisecond)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"users": [{"id": 1, "name": "John"}]}`))
}

// ordersHandler - обработчик с пользовательским контекстом
func ordersHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Processing request: %s %s\n", r.Method, r.URL.Path)

	// Получаем пользователя из заголовка
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = "anonymous"
	}

	// Логируем бизнес-событие
	sentry.AddBreadcrumb(&sentry.Breadcrumb{
		Message:   "User accessed orders",
		Level:     sentry.LevelInfo,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"user_id":  userID,
			"endpoint": "/api/orders",
		},
	})

	// Симуляция работы
	time.Sleep(150 * time.Millisecond)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"orders": [{"id": 1, "user_id": "` + userID + `"}]}`))
}

// errorHandler - обработчик для демонстрации ошибки
func errorHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Processing request: %s %s\n", r.Method, r.URL.Path)

	// Симуляция ошибки
	err := errors.New("database connection failed")

	// Отправляем ошибку в Sentry с контекстом HTTP запроса
	monitoring.CaptureHTTPError(r, err, map[string]string{
		"component": "error_handler",
		"operation": "process_request",
	}, map[string]interface{}{
		"request_id": r.Header.Get("X-Request-ID"),
		"timestamp":  time.Now(),
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(`{"error": "Internal server error"}`))
}

// panicHandler - обработчик для демонстрации паники
func panicHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Processing request: %s %s\n", r.Method, r.URL.Path)

	// Симуляция паники (будет перехвачена Sentry middleware)
	panic("Something went wrong in the handler!")
}
