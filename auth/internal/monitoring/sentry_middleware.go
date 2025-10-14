package monitoring

import (
	"context"
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// SentryMiddleware - middleware для интеграции Sentry с HTTP и gRPC
type SentryMiddleware struct {
	enterprise *SentryEnterprise
}

// NewSentryMiddleware создает новый middleware
func NewSentryMiddleware(enterprise *SentryEnterprise) *SentryMiddleware {
	return &SentryMiddleware{
		enterprise: enterprise,
	}
}

// HTTPMiddleware создает HTTP middleware для Sentry
func (sm *SentryMiddleware) HTTPMiddleware() func(http.Handler) http.Handler {
	// Используем официальный HTTP middleware от Sentry
	sentryHandler := sentryhttp.New(sentryhttp.Options{
		Repanic:         true,
		WaitForDelivery: true,
		Timeout:         2 * time.Second,
	})

	// Оборачиваем в дополнительный middleware для расширенного контекста
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Создаем hub для этого запроса
			hub := sentry.GetHubFromContext(r.Context())
			if hub == nil {
				hub = sentry.CurrentHub().Clone()
			}

			// Устанавливаем контекст запроса в Sentry
			hub.Scope().SetRequest(r)

			// Добавляем дополнительную информацию о запросе
			hub.Scope().SetTag("http.method", r.Method)
			hub.Scope().SetTag("http.url", r.URL.String())
			hub.Scope().SetTag("http.user_agent", r.UserAgent())
			hub.Scope().SetTag("http.remote_addr", r.RemoteAddr)

			// Добавляем заголовки как контекст (фильтруем чувствительные)
			headers := make(map[string]string)
			for key, values := range r.Header {
				// Пропускаем чувствительные заголовки
				if key == "Authorization" || key == "Cookie" || key == "X-API-Key" {
					continue
				}
				if len(values) > 0 {
					headers[key] = values[0]
				}
			}
			hub.Scope().SetContext("http.headers", headers)

			// Добавляем информацию о пользователе если есть
			if userID := r.Header.Get("X-User-ID"); userID != "" {
				hub.Scope().SetUser(sentry.User{
					ID: userID,
				})
			}

			// Добавляем информацию о сессии если есть
			if sessionID := r.Header.Get("X-Session-ID"); sessionID != "" {
				hub.Scope().SetTag("session.id", sessionID)
			}

			// Добавляем информацию о запросе в контекст
			ctx := sentry.SetHubOnContext(r.Context(), hub)
			r = r.WithContext(ctx)

			// Вызываем оригинальный Sentry middleware
			sentryHandler.Handle(next).ServeHTTP(w, r)
		})
	}
}

// GRPCUnaryInterceptor создает gRPC unary interceptor для Sentry
func (sm *SentryMiddleware) GRPCUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Создаем транзакцию для gRPC запроса
		transaction := sm.enterprise.StartTransactionWithContext(
			ctx,
			info.FullMethod,
			"grpc",
			map[string]string{
				"method": info.FullMethod,
				"type":   "unary",
			},
		)
		defer transaction.Finish()

		// Добавляем контекст к транзакции
		transaction.SetContext("grpc", map[string]interface{}{
			"method": info.FullMethod,
			"type":   "unary",
		})

		// Выполняем запрос
		start := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(start)

		// Логируем производительность
		sm.enterprise.CapturePerformanceIssue(
			info.FullMethod,
			duration,
			100*time.Millisecond, // threshold
			map[string]interface{}{
				"grpc_method": info.FullMethod,
				"duration":    duration.String(),
			},
		)

		// Обрабатываем ошибки
		if err != nil {
			grpcStatus := status.Code(err)

			// Отправляем в Sentry только серьезные ошибки
			if grpcStatus == codes.Internal || grpcStatus == codes.Unknown {
				sm.enterprise.CaptureErrorWithContext(
					ctx,
					err,
					map[string]string{
						"grpc_method": info.FullMethod,
						"grpc_code":   grpcStatus.String(),
					},
					map[string]interface{}{
						"request": req,
					},
					nil,
				)
			}
		}

		return resp, err
	}
}

// GRPCStreamInterceptor создает gRPC stream interceptor для Sentry
func (sm *SentryMiddleware) GRPCStreamInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		// Создаем транзакцию для gRPC stream
		transaction := sm.enterprise.StartTransactionWithContext(
			ss.Context(),
			info.FullMethod,
			"grpc",
			map[string]string{
				"method": info.FullMethod,
				"type":   "stream",
			},
		)
		defer transaction.Finish()

		// Выполняем stream
		start := time.Now()
		err := handler(srv, ss)
		duration := time.Since(start)

		// Логируем производительность
		sm.enterprise.CapturePerformanceIssue(
			info.FullMethod,
			duration,
			500*time.Millisecond, // threshold для stream
			map[string]interface{}{
				"grpc_method": info.FullMethod,
				"duration":    duration.String(),
			},
		)

		// Обрабатываем ошибки
		if err != nil {
			grpcStatus := status.Code(err)

			if grpcStatus == codes.Internal || grpcStatus == codes.Unknown {
				sm.enterprise.CaptureErrorWithContext(
					ss.Context(),
					err,
					map[string]string{
						"grpc_method": info.FullMethod,
						"grpc_code":   grpcStatus.String(),
					},
					map[string]interface{}{
						"stream_type": "server",
					},
					nil,
				)
			}
		}

		return err
	}
}

// AddHTTPContextToSentry добавляет контекст HTTP запроса в Sentry вручную
func AddHTTPContextToSentry(r *http.Request) {
	hub := sentry.GetHubFromContext(r.Context())
	if hub == nil {
		hub = sentry.CurrentHub()
	}

	// Устанавливаем контекст запроса
	hub.Scope().SetRequest(r)

	// Добавляем теги
	hub.Scope().SetTag("http.method", r.Method)
	hub.Scope().SetTag("http.url", r.URL.String())
	hub.Scope().SetTag("http.user_agent", r.UserAgent())
	hub.Scope().SetTag("http.remote_addr", r.RemoteAddr)

	// Добавляем заголовки (фильтруем чувствительные)
	headers := make(map[string]string)
	for key, values := range r.Header {
		if key == "Authorization" || key == "Cookie" || key == "X-API-Key" {
			continue
		}
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}
	hub.Scope().SetContext("http.headers", headers)

	// Добавляем информацию о пользователе
	if userID := r.Header.Get("X-User-ID"); userID != "" {
		hub.Scope().SetUser(sentry.User{ID: userID})
	}

	// Добавляем информацию о сессии
	if sessionID := r.Header.Get("X-Session-ID"); sessionID != "" {
		hub.Scope().SetTag("session.id", sessionID)
	}
}

// CaptureHTTPError отправляет ошибку с контекстом HTTP запроса
func CaptureHTTPError(r *http.Request, err error, tags map[string]string, extra map[string]interface{}) {
	// Добавляем контекст HTTP запроса
	AddHTTPContextToSentry(r)

	// Добавляем дополнительные теги
	hub := sentry.GetHubFromContext(r.Context())
	if hub == nil {
		hub = sentry.CurrentHub()
	}

	for key, value := range tags {
		hub.Scope().SetTag(key, value)
	}

	// Добавляем дополнительную информацию
	for key, value := range extra {
		hub.Scope().SetContext(key, map[string]interface{}{"value": value})
	}

	// Отправляем ошибку
	sentry.CaptureException(err)
}

// DatabaseMiddleware создает middleware для отслеживания запросов к БД
func (sm *SentryMiddleware) DatabaseMiddleware() func(func() error) error {
	return func(dbOperation func() error) error {
		start := time.Now()
		err := dbOperation()
		duration := time.Since(start)

		// Логируем медленные запросы
		if duration > 1*time.Second {
			sm.enterprise.CapturePerformanceIssue(
				"database_query",
				duration,
				1*time.Second,
				map[string]interface{}{
					"operation": "database",
					"duration":  duration.String(),
				},
			)
		}

		// Логируем ошибки БД
		if err != nil {
			sm.enterprise.CaptureErrorWithContext(
				context.Background(),
				err,
				map[string]string{
					"component": "database",
				},
				map[string]interface{}{
					"duration": duration.String(),
				},
				nil,
			)
		}

		return err
	}
}

// AddHTTPContextToSentry добавляет контекст HTTP запроса в Sentry вручную
func AddHTTPContextToSentry(r *http.Request) {
	hub := sentry.GetHubFromContext(r.Context())
	if hub == nil {
		hub = sentry.CurrentHub()
	}

	// Устанавливаем контекст запроса
	hub.Scope().SetRequest(r)

	// Добавляем теги
	hub.Scope().SetTag("http.method", r.Method)
	hub.Scope().SetTag("http.url", r.URL.String())
	hub.Scope().SetTag("http.user_agent", r.UserAgent())
	hub.Scope().SetTag("http.remote_addr", r.RemoteAddr)

	// Добавляем заголовки (фильтруем чувствительные)
	headers := make(map[string]string)
	for key, values := range r.Header {
		if key == "Authorization" || key == "Cookie" || key == "X-API-Key" {
			continue
		}
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}
	hub.Scope().SetContext("http.headers", headers)

	// Добавляем информацию о пользователе
	if userID := r.Header.Get("X-User-ID"); userID != "" {
		hub.Scope().SetUser(sentry.User{ID: userID})
	}

	// Добавляем информацию о сессии
	if sessionID := r.Header.Get("X-Session-ID"); sessionID != "" {
		hub.Scope().SetTag("session.id", sessionID)
	}
}

// CaptureHTTPError отправляет ошибку с контекстом HTTP запроса
func CaptureHTTPError(r *http.Request, err error, tags map[string]string, extra map[string]interface{}) {
	// Добавляем контекст HTTP запроса
	AddHTTPContextToSentry(r)

	// Добавляем дополнительные теги
	hub := sentry.GetHubFromContext(r.Context())
	if hub == nil {
		hub = sentry.CurrentHub()
	}

	for key, value := range tags {
		hub.Scope().SetTag(key, value)
	}

	// Добавляем дополнительную информацию
	for key, value := range extra {
		hub.Scope().SetContext(key, map[string]interface{}{"value": value})
	}

	// Отправляем ошибку
	sentry.CaptureException(err)
}

// BusinessLogicMiddleware создает middleware для бизнес-логики
func (sm *SentryMiddleware) BusinessLogicMiddleware(operation string) func(func() error) error {
	return func(businessOperation func() error) error {
		start := time.Now()
		err := businessOperation()
		duration := time.Since(start)

		// Логируем производительность бизнес-логики
		sm.enterprise.CapturePerformanceIssue(
			operation,
			duration,
			500*time.Millisecond,
			map[string]interface{}{
				"operation": operation,
				"duration":  duration.String(),
			},
		)

		// Логируем ошибки бизнес-логики
		if err != nil {
			sm.enterprise.CaptureErrorWithContext(
				context.Background(),
				err,
				map[string]string{
					"component": "business_logic",
					"operation": operation,
				},
				map[string]interface{}{
					"duration": duration.String(),
				},
				nil,
			)
		}

		return err
	}
}

// AddHTTPContextToSentry добавляет контекст HTTP запроса в Sentry вручную
func AddHTTPContextToSentry(r *http.Request) {
	hub := sentry.GetHubFromContext(r.Context())
	if hub == nil {
		hub = sentry.CurrentHub()
	}

	// Устанавливаем контекст запроса
	hub.Scope().SetRequest(r)

	// Добавляем теги
	hub.Scope().SetTag("http.method", r.Method)
	hub.Scope().SetTag("http.url", r.URL.String())
	hub.Scope().SetTag("http.user_agent", r.UserAgent())
	hub.Scope().SetTag("http.remote_addr", r.RemoteAddr)

	// Добавляем заголовки (фильтруем чувствительные)
	headers := make(map[string]string)
	for key, values := range r.Header {
		if key == "Authorization" || key == "Cookie" || key == "X-API-Key" {
			continue
		}
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}
	hub.Scope().SetContext("http.headers", headers)

	// Добавляем информацию о пользователе
	if userID := r.Header.Get("X-User-ID"); userID != "" {
		hub.Scope().SetUser(sentry.User{ID: userID})
	}

	// Добавляем информацию о сессии
	if sessionID := r.Header.Get("X-Session-ID"); sessionID != "" {
		hub.Scope().SetTag("session.id", sessionID)
	}
}

// CaptureHTTPError отправляет ошибку с контекстом HTTP запроса
func CaptureHTTPError(r *http.Request, err error, tags map[string]string, extra map[string]interface{}) {
	// Добавляем контекст HTTP запроса
	AddHTTPContextToSentry(r)

	// Добавляем дополнительные теги
	hub := sentry.GetHubFromContext(r.Context())
	if hub == nil {
		hub = sentry.CurrentHub()
	}

	for key, value := range tags {
		hub.Scope().SetTag(key, value)
	}

	// Добавляем дополнительную информацию
	for key, value := range extra {
		hub.Scope().SetContext(key, map[string]interface{}{"value": value})
	}

	// Отправляем ошибку
	sentry.CaptureException(err)
}
