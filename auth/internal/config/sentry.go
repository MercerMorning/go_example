package config

import (
	"os"
)

type SentryConfig struct {
	DSN              string
	Environment      string
	Release          string
	Debug            bool
	SampleRate       float64
	TracesSampleRate float64
}

func NewSentryConfig() *SentryConfig {
	return &SentryConfig{
		DSN:              getEnv("SENTRY_DSN", ""),
		Environment:      getEnv("SENTRY_ENVIRONMENT", "development"),
		Release:          getEnv("SENTRY_RELEASE", ""),
		Debug:            getEnv("SENTRY_DEBUG", "false") == "true",
		SampleRate:       1.0,
		TracesSampleRate: 0.1,
	}
}

func (c *SentryConfig) IsEnabled() bool {
	return c.DSN != ""
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
