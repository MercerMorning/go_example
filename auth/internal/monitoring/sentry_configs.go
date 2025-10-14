package monitoring

import (
	"os"
	"strconv"
	"time"

	"github.com/getsentry/sentry-go"
)

// EnvironmentConfig - конфигурация для разных окружений
type EnvironmentConfig struct {
	Environment      string
	SampleRate       float64
	TracesSampleRate float64
	Debug            bool
	MaxBreadcrumbs   int
	FlushTimeout     time.Duration
}

// GetEnvironmentConfig возвращает конфигурацию для окружения
func GetEnvironmentConfig(env string) *EnvironmentConfig {
	switch env {
	case "development":
		return &EnvironmentConfig{
			Environment:      "development",
			SampleRate:       1.0, // 100% в dev
			TracesSampleRate: 1.0, // 100% в dev
			Debug:            true,
			MaxBreadcrumbs:   50,
			FlushTimeout:     5 * time.Second,
		}
	case "staging":
		return &EnvironmentConfig{
			Environment:      "staging",
			SampleRate:       0.5, // 50% в staging
			TracesSampleRate: 0.3, // 30% в staging
			Debug:            false,
			MaxBreadcrumbs:   100,
			FlushTimeout:     3 * time.Second,
		}
	case "production":
		return &EnvironmentConfig{
			Environment:      "production",
			SampleRate:       0.1,  // 10% в production
			TracesSampleRate: 0.05, // 5% в production
			Debug:            false,
			MaxBreadcrumbs:   100,
			FlushTimeout:     2 * time.Second,
		}
	default:
		return &EnvironmentConfig{
			Environment:      "unknown",
			SampleRate:       0.1,
			TracesSampleRate: 0.05,
			Debug:            false,
			MaxBreadcrumbs:   100,
			FlushTimeout:     2 * time.Second,
		}
	}
}

// LoadSentryConfigFromEnv загружает конфигурацию Sentry из переменных окружения
func LoadSentryConfigFromEnv() *SentryEnterpriseConfig {
	env := getEnv("SENTRY_ENVIRONMENT", "development")
	envConfig := GetEnvironmentConfig(env)

	dsn := getEnv("SENTRY_DSN", "")
	if dsn == "" {
		return nil // Sentry отключен
	}

	release := getEnv("SENTRY_RELEASE", "unknown")
	serverName := getEnv("SENTRY_SERVER_NAME", getHostname())

	// Загружаем дополнительные настройки из env
	maxBreadcrumbs := getEnvInt("SENTRY_MAX_BREADCRUMBS", envConfig.MaxBreadcrumbs)
	maxSpans := getEnvInt("SENTRY_MAX_SPANS", 1000)
	maxTraceFileSize := getEnvInt64("SENTRY_MAX_TRACE_FILE_SIZE", 10*1024*1024) // 10MB

	config := &SentryEnterpriseConfig{
		DSN:              dsn,
		Environment:      envConfig.Environment,
		Release:          release,
		Debug:            envConfig.Debug,
		SampleRate:       envConfig.SampleRate,
		TracesSampleRate: envConfig.TracesSampleRate,
		MaxBreadcrumbs:   maxBreadcrumbs,
		MaxSpans:         maxSpans,
		MaxTraceFileSize: maxTraceFileSize,
		FlushTimeout:     envConfig.FlushTimeout,
		ServerName:       serverName,
		Dist:             getEnv("SENTRY_DIST", ""),
		Tags: map[string]string{
			"service":     getEnv("SENTRY_SERVICE", "auth-service"),
			"version":     release,
			"environment": envConfig.Environment,
			"server":      serverName,
		},
		BeforeSend:       createBeforeSendHandler(envConfig.Environment),
		BeforeBreadcrumb: createBeforeBreadcrumbHandler(envConfig.Environment),
	}

	// Добавляем дополнительные теги из env
	additionalTags := getEnv("SENTRY_ADDITIONAL_TAGS", "")
	if additionalTags != "" {
		// Формат: "key1:value1,key2:value2"
		// Парсим дополнительные теги
		// TODO: реализовать парсинг
	}

	return config
}

// createBeforeSendHandler создает обработчик BeforeSend для окружения
func createBeforeSendHandler(environment string) func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
	return func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
		// Добавляем информацию о сервере
		event.ServerName = getHostname()
		event.Tags["processed_at"] = time.Now().Format(time.RFC3339)

		// Фильтрация в зависимости от окружения
		switch environment {
		case "development":
			// В dev логируем все
			return event
		case "staging":
			// В staging фильтруем debug события
			if event.Level == sentry.LevelDebug {
				return nil
			}
			return event
		case "production":
			// В production агрессивная фильтрация
			if event.Level == sentry.LevelDebug {
				return nil
			}
			// Фильтруем частые ошибки
			if isFrequentError(event) {
				return nil
			}
			return event
		default:
			return event
		}
	}
}

// createBeforeBreadcrumbHandler создает обработчик BeforeBreadcrumb для окружения
func createBeforeBreadcrumbHandler(environment string) func(breadcrumb *sentry.Breadcrumb, hint *sentry.BreadcrumbHint) *sentry.Breadcrumb {
	return func(breadcrumb *sentry.Breadcrumb, hint *sentry.BreadcrumbHint) *sentry.Breadcrumb {
		// Фильтрация breadcrumbs в зависимости от окружения
		switch environment {
		case "development":
			return breadcrumb
		case "staging", "production":
			// Фильтруем debug breadcrumbs
			if breadcrumb.Level == sentry.LevelDebug {
				return nil
			}
			return breadcrumb
		default:
			return breadcrumb
		}
	}
}

// isFrequentError проверяет, является ли ошибка частой
func isFrequentError(event *sentry.Event) bool {
	// Простая логика для определения частых ошибок
	// В реальном проекте это может быть более сложная логика
	if event.Exception != nil && len(event.Exception) > 0 {
		exception := event.Exception[0]
		if exception.Type == "context.DeadlineExceeded" {
			return true // Игнорируем таймауты
		}
	}
	return false
}

// getEnv получает переменную окружения с значением по умолчанию
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt получает int переменную окружения
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvInt64 получает int64 переменную окружения
func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getHostname возвращает имя хоста
func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}
