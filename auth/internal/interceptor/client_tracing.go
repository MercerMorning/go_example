package interceptor

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/uber/jaeger-client-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func ClientTracingInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	// Создаем span для исходящего вызова
	span, ctx := opentracing.StartSpanFromContext(ctx, method)
	defer span.Finish()

	// Устанавливаем теги для клиентского span
	ext.SpanKindRPCClient.Set(span)
	ext.Component.Set(span, "grpc-client")
	ext.PeerService.Set(span, "other_service")

	// Получаем span context и добавляем trace ID в метаданные
	spanContext, ok := span.Context().(jaeger.SpanContext)
	if ok {
		// Добавляем trace ID в исходящие метаданные
		ctx = metadata.AppendToOutgoingContext(ctx, "x-trace-id", spanContext.TraceID().String())
	}

	// Выполняем вызов
	err := invoker(ctx, method, req, reply, cc, opts...)
	
	if err != nil {
		ext.Error.Set(span, true)
		span.SetTag("error", err.Error())
	}

	return err
}
