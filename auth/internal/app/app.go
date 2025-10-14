package app

import (
	"context"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	"github.com/MercerMorning/go_example/auth/internal/closer"
	"github.com/MercerMorning/go_example/auth/internal/config"
	"github.com/MercerMorning/go_example/auth/internal/interceptor"
	"github.com/MercerMorning/go_example/auth/internal/metric"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"

	desc "github.com/MercerMorning/go_example/auth/pkg/user_v1"

	"github.com/MercerMorning/go_example/auth/internal/logger"
	grpcMiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
)

type App struct {
	serviceProvider *serviceProvider
	grpcServer      *grpc.Server
	httpServer      *http.Server
}

func NewApp(ctx context.Context) (*App, error) {
	a := &App{}

	err := a.initDeps(ctx)
	if err != nil {
		return nil, err
	}

	return a, nil
}

func (a *App) Run() error {
	defer func() {
		// Flush Sentry перед завершением
		logger.FlushSentry(2 * time.Second)
		closer.CloseAll()
		closer.Wait()
	}()

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		defer logger.RecoverPanicSilent() // Перехватываем паники в горутинах

		err := a.runGRPCServer()
		if err != nil {
			logger.Error("Failed to run GRPC server",
				zap.String("error", err.Error()))
			log.Fatalf("failed to run GRPC server: %v", err)
		}
	}()

	go func() {
		defer wg.Done()
		defer logger.RecoverPanicSilent() // Перехватываем паники в горутинах

		err := a.runHTTPServer()
		if err != nil {
			logger.Error("Failed to run HTTP server",
				zap.String("error", err.Error()))
			log.Fatalf("failed to run HTTP server: %v", err)
		}
	}()

	go func() {
		defer wg.Done()
		defer logger.RecoverPanicSilent() // Перехватываем паники в горутинах

		err := a.runMetric()
		if err != nil {
			logger.Error("Failed to run METRIC server",
				zap.String("error", err.Error()))
			log.Fatalf("failed to run METRIC server: %v", err)
		}
	}()

	// go func() {
	// 	defer wg.Done()

	// 	err := a.runSwaggerServer()
	// 	if err != nil {
	// 		log.Fatalf("failed to run Swagger server: %v", err)
	// 	}
	// }()

	wg.Wait()

	return nil
}

func (a *App) runMetric() error {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	prometheusServer := &http.Server{
		Addr:    "localhost:2112",
		Handler: mux,
	}

	log.Printf("Prometheus server is running on %s", "localhost:2112")

	err := prometheusServer.ListenAndServe()
	if err != nil {
		return err
	}

	return nil
}

func (a *App) runHTTPServer() error {
	log.Printf("HTTP server is running on %s", a.serviceProvider.HTTPConfig().Address())

	err := a.httpServer.ListenAndServe()
	if err != nil {
		return err
	}

	return nil
}

func (a *App) initDeps(ctx context.Context) error {
	inits := []func(context.Context) error{
		a.initConfig,
		a.initSentry,
		a.initMonitoring,
		a.initLogger,
		a.initServiceProvider,
		a.initGRPCServer,
		a.initHTTPServer,
		a.initMetric,
	}

	for _, f := range inits {
		err := f(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *App) initMetric(ctx context.Context) error {
	err := metric.Init(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (a *App) initHTTPServer(ctx context.Context) error {
	mux := runtime.NewServeMux()

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	err := desc.RegisterUserV1HandlerFromEndpoint(ctx, mux, a.serviceProvider.GRPCConfig().Address(), opts)
	if err != nil {
		return err
	}

	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type", "Content-Length", "Authorization"},
		AllowCredentials: true,
	})

	a.httpServer = &http.Server{
		Addr:    a.serviceProvider.HTTPConfig().Address(),
		Handler: corsMiddleware.Handler(mux),
	}

	return nil
}

func (a *App) initMonitoring(_ context.Context) error {
	sentryConfig := config.NewSentryConfig()

	if sentryConfig.IsEnabled() {
		err := logger.InitSentry(
			sentryConfig.DSN,
			sentryConfig.Environment,
			sentryConfig.Release,
			sentryConfig.Debug,
			sentryConfig.SampleRate,
			sentryConfig.TracesSampleRate,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *App) initSentry(_ context.Context) error {
	sentryConfig := config.NewSentryConfig()

	if sentryConfig.IsEnabled() {
		err := logger.InitSentry(
			sentryConfig.DSN,
			sentryConfig.Environment,
			sentryConfig.Release,
			sentryConfig.Debug,
			sentryConfig.SampleRate,
			sentryConfig.TracesSampleRate,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *App) initLogger(_ context.Context) error {
	zapLog, err := zap.NewProduction()
	if err != nil {
		return err
	}

	// Создаем Sentry core если Sentry включен
	sentryConfig := config.NewSentryConfig()
	var core zapcore.Core = zapLog.Core()

	if sentryConfig.IsEnabled() {
		sentryCore := logger.NewSentryCore(zapLog.Core(), zapcore.ErrorLevel)
		core = sentryCore
	}

	logger.Init(core)

	return nil
}

func (a *App) initConfig(_ context.Context) error {
	err := config.Load(".env")
	if err != nil {
		return err
	}

	return nil
}

func (a *App) initServiceProvider(_ context.Context) error {
	a.serviceProvider = newServiceProvider()
	return nil
}

func (a *App) initGRPCServer(ctx context.Context) error {
	// a.grpcServer = grpc.NewServer(grpc.Creds(insecure.NewCredentials()))
	logger.Info("init grpc server")

	a.grpcServer = grpc.NewServer(
		grpc.Creds(insecure.NewCredentials()),
		grpc.UnaryInterceptor(
			grpcMiddleware.ChainUnaryServer(
				interceptor.LogInterceptor,
				interceptor.MetricsInterceptor,
			),
		),
	)

	reflection.Register(a.grpcServer)

	desc.RegisterUserV1Server(a.grpcServer, a.serviceProvider.UserImpl(ctx))

	return nil
}

func (a *App) runGRPCServer() error {
	log.Printf("GRPC server is running on %s", a.serviceProvider.GRPCConfig().Address())

	list, err := net.Listen("tcp", a.serviceProvider.GRPCConfig().Address())
	if err != nil {
		return err
	}

	err = a.grpcServer.Serve(list)
	if err != nil {
		return err
	}

	return nil
}
