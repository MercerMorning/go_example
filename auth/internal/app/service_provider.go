package app

import (
	"context"
	"log"

	"github.com/MercerMorning/go_example/auth/internal/api/user"
	"github.com/MercerMorning/go_example/auth/internal/client/db"
	"github.com/MercerMorning/go_example/auth/internal/client/db/pg"
	"github.com/MercerMorning/go_example/auth/internal/client/db/transaction"
	"github.com/MercerMorning/go_example/auth/internal/closer"
	"github.com/MercerMorning/go_example/auth/internal/config"
	"github.com/MercerMorning/go_example/auth/internal/interceptor"
	"github.com/MercerMorning/go_example/auth/internal/repository"
	"github.com/MercerMorning/go_example/auth/internal/service"
	desc "github.com/MercerMorning/go_example/auth/pkg/user_v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	userRepository "github.com/MercerMorning/go_example/auth/internal/repository/user"
	userService "github.com/MercerMorning/go_example/auth/internal/service/user"
)

type serviceProvider struct {
	pgConfig   config.PGConfig
	grpcConfig config.GRPCConfig
	httpConfig config.HTTPConfig

	dbClient       db.Client
	txManager      db.TxManager
	userRepository repository.UserRepository

	userService service.UserService
	userClient  desc.UserV1Client

	userImpl *user.Implementation
}

func newServiceProvider() *serviceProvider {
	return &serviceProvider{}
}

func (s *serviceProvider) HTTPConfig() config.HTTPConfig {
	if s.httpConfig == nil {
		cfg, err := config.NewHTTPConfig()
		if err != nil {
			log.Fatalf("failed to get http config: %s", err.Error())
		}

		s.httpConfig = cfg
	}

	return s.httpConfig
}

func (s *serviceProvider) PGConfig() config.PGConfig {
	if s.pgConfig == nil {
		cfg, err := config.NewPGConfig()
		if err != nil {
			log.Fatalf("failed to get pg config: %s", err.Error())
		}

		s.pgConfig = cfg
	}

	return s.pgConfig
}

func (s *serviceProvider) GRPCConfig() config.GRPCConfig {
	if s.grpcConfig == nil {
		cfg, err := config.NewGRPCConfig()
		if err != nil {
			log.Fatalf("failed to get grpc config: %s", err.Error())
		}

		s.grpcConfig = cfg
	}

	return s.grpcConfig
}

func (s *serviceProvider) DBClient(ctx context.Context) db.Client {
	if s.dbClient == nil {
		cl, err := pg.New(ctx, s.PGConfig().DSN())
		if err != nil {
			log.Fatalf("failed to create db client: %v", err)
		}

		err = cl.DB().Ping(ctx)
		if err != nil {
			log.Fatalf("ping error: %s", err.Error())
		}
		closer.Add(cl.Close)

		s.dbClient = cl
	}

	return s.dbClient
}

func (s *serviceProvider) TxManager(ctx context.Context) db.TxManager {
	if s.txManager == nil {
		s.txManager = transaction.NewTransactionManager(s.DBClient(ctx).DB())
	}

	return s.txManager
}

func (s *serviceProvider) NoteRepository(ctx context.Context) repository.UserRepository {
	if s.userRepository == nil {
		s.userRepository = userRepository.NewRepository(s.DBClient(ctx))
	}

	return s.userRepository
}

func (s *serviceProvider) UserService(ctx context.Context) service.UserService {
	if s.userService == nil {
		s.userService = userService.NewService(
			s.NoteRepository(ctx),
			s.TxManager(ctx),
		)
	}

	return s.userService
}

func (s *serviceProvider) UserClient(ctx context.Context) desc.UserV1Client {
	if s.userClient == nil {
		// Подключаемся к other_service на порту 50052 с клиентским интерцептором для трейсинга
		conn, err := grpc.Dial("localhost:50052", 
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithUnaryInterceptor(interceptor.ClientTracingInterceptor),
		)
		if err != nil {
			log.Fatalf("failed to connect to other_service: %v", err)
		}

		s.userClient = desc.NewUserV1Client(conn)
		closer.Add(conn.Close)
	}

	return s.userClient
}

func (s *serviceProvider) UserImpl(ctx context.Context) *user.Implementation {
	if s.userImpl == nil {
		s.userImpl = user.NewImplementation(s.UserService(ctx), s.UserClient(ctx))
	}

	return s.userImpl
}
