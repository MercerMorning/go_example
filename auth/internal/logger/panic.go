package logger

import (
	"runtime/debug"

	"github.com/getsentry/sentry-go"
	"go.uber.org/zap"
)

// RecoverPanic восстанавливается от паники и отправляет её в Sentry
func RecoverPanic() {
	if r := recover(); r != nil {
		// Логируем панику
		Error("Panic recovered",
			zap.Any("panic", r),
			zap.String("stack", string(debug.Stack())),
		)

		// Отправляем в Sentry
		sentry.WithScope(func(scope *sentry.Scope) {
			scope.SetLevel(sentry.LevelFatal)
			scope.SetTag("error_type", "panic")
			scope.SetContext("panic_info", map[string]interface{}{
				"panic_value": r,
				"stack":       string(debug.Stack()),
			})
			sentry.CaptureException(&panicError{value: r})
		})

		// Re-panic чтобы приложение завершилось
		panic(r)
	}
}

// RecoverPanicSilent восстанавливается от паники без re-panic
func RecoverPanicSilent() {
	if r := recover(); r != nil {
		// Логируем панику
		Error("Panic recovered (silent)",
			zap.Any("panic", r),
			zap.String("stack", string(debug.Stack())),
		)

		// Отправляем в Sentry
		sentry.WithScope(func(scope *sentry.Scope) {
			scope.SetLevel(sentry.LevelFatal)
			scope.SetTag("error_type", "panic")
			scope.SetContext("panic_info", map[string]interface{}{
				"panic_value": r,
				"stack":       string(debug.Stack()),
			})
			sentry.CaptureException(&panicError{value: r})
		})
	}
}

// panicError - кастомный тип ошибки для паник
type panicError struct {
	value interface{}
}

func (e *panicError) Error() string {
	return "panic occurred"
}

// WithPanicRecovery оборачивает функцию с перехватом паник
func WithPanicRecovery(fn func()) {
	defer RecoverPanic()
	fn()
}

// WithPanicRecoverySilent оборачивает функцию с перехватом паник без re-panic
func WithPanicRecoverySilent(fn func()) {
	defer RecoverPanicSilent()
	fn()
}
