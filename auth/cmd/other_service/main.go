package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/MercerMorning/go_example/auth/internal/interceptor"
	"github.com/MercerMorning/go_example/auth/internal/logger"
	"github.com/MercerMorning/go_example/auth/internal/tracing"
	"github.com/MercerMorning/go_example/auth/pkg/user_v1"
)

var logLevel = flag.String("l", "info", "log level")

const (
	grpcPort    = 50052
	serviceName = "user-service"
)

type server struct {
	user_v1.UnimplementedUserV1Server
}

// Create creates a new user
func (s *server) Create(ctx context.Context, req *user_v1.CreateRequest) (*user_v1.CreateResponse, error) {
	fmt.Println("IAM HERE CREATE")
	if req.GetName() == "" {
		return nil, errors.Errorf("name is empty")
	}
	if req.GetEmail() == "" {
		return nil, errors.Errorf("email is empty")
	}
	if req.GetPassword() == "" {
		return nil, errors.Errorf("password is empty")
	}
	if req.GetPassword() != req.GetPasswordConfirm() {
		return nil, errors.Errorf("passwords do not match")
	}

	// Simulate processing time
	time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)

	// Generate a fake user ID
	userID := int64(rand.Intn(10000) + 1)

	return &user_v1.CreateResponse{
		Id: userID,
	}, nil
}

// Get retrieves a user by ID
func (s *server) Get(ctx context.Context, req *user_v1.GetRequest) (*user_v1.GetResponse, error) {
	fmt.Println("IAM HERE")
	if req.GetId() == 0 {
		return nil, errors.Errorf("id is empty")
	}

	// Simulate processing time
	time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)

	return &user_v1.GetResponse{
		Id:        req.GetId(),
		Name:      fmt.Sprintf("User %d", req.GetId()),
		Email:     fmt.Sprintf("user%d@example.com", req.GetId()),
		Role:      user_v1.Role_USER,
		CreatedAt: timestamppb.New(time.Now().Add(-time.Hour * 24 * 30)), // 30 days ago
		UpdatedAt: timestamppb.New(time.Now()),
	}, nil
}

// Update updates an existing user
func (s *server) Update(ctx context.Context, req *user_v1.UpdateRequest) (*emptypb.Empty, error) {
	if req.GetId() == 0 {
		return nil, errors.Errorf("id is empty")
	}

	// Simulate processing time
	time.Sleep(time.Duration(rand.Intn(300)) * time.Millisecond)

	return &emptypb.Empty{}, nil
}

// Delete deletes a user by ID
func (s *server) Delete(ctx context.Context, req *user_v1.DeleteRequest) (*emptypb.Empty, error) {
	if req.GetId() == 0 {
		return nil, errors.Errorf("id is empty")
	}

	// Simulate processing time
	time.Sleep(time.Duration(rand.Intn(200)) * time.Millisecond)

	return &emptypb.Empty{}, nil
}

func main() {
	fmt.Println("start other service")
	flag.Parse()

	logger.Init(getCore(getAtomicLevel()))

	// Create a simple logger for tracing
	zapLogger, _ := zap.NewProduction()
	tracing.Init(zapLogger, serviceName)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer(
		grpc.UnaryInterceptor(
			interceptor.ServerTracingInterceptor,
		),
	)
	reflection.Register(s)
	user_v1.RegisterUserV1Server(s, &server{})

	log.Printf("server listening at %v", lis.Addr())

	if err = s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func getCore(level zap.AtomicLevel) zapcore.Core {
	stdout := zapcore.AddSync(os.Stdout)

	developmentCfg := zap.NewDevelopmentEncoderConfig()
	developmentCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder

	consoleEncoder := zapcore.NewConsoleEncoder(developmentCfg)

	return zapcore.NewCore(consoleEncoder, stdout, level)
}

func getAtomicLevel() zap.AtomicLevel {
	var level zapcore.Level
	if err := level.Set(*logLevel); err != nil {
		log.Fatalf("failed to set log level: %v", err)
	}

	return zap.NewAtomicLevelAt(level)
}
