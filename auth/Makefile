LOCAL_BIN:=$(CURDIR)/bin

install-golangci-lint:
	GOBIN=$(LOCAL_BIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.61.0

lint:
	$(LOCAL_BIN)/golangci-lint run ./... --config .golangci.pipeline.yaml

install-deps:
	GOBIN=$(LOCAL_BIN) go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.1
	GOBIN=$(LOCAL_BIN) go install -mod=mod google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2

generate-user-api:
	mkdir -p pkg/user_v1
	protoc --proto_path api/user_v1 \
	--go_out=pkg/user_v1 --go_opt=paths=source_relative \
	--plugin=protoc-gen-go=bin/protoc-gen-go \
	--go-grpc_out=pkg/user_v1 --go-grpc_opt=paths=source_relative \
	--plugin=protoc-gen-go-grpc=bin/protoc-gen-go-grpc \
	api/user_v1/user.proto

# Тестирование
test-unit:
	go test -v ./internal/api/user/tests/create_test.go

test-integration:
	@echo "Starting PostgreSQL container for integration tests..."
	docker-compose -f docker-compose.test.yaml up -d
	@echo "Waiting for database to be ready..."
	sleep 5
	@echo "Running integration tests..."
	TEST_PG_DSN="postgres://postgres:postgres@localhost:54322/auth_test?sslmode=disable" go test -v ./internal/api/user/tests/integration_test.go -run TestIntegrationSuite
	@echo "Stopping PostgreSQL container..."
	docker-compose -f docker-compose.test.yaml down

test-integration-local:
	@echo "Running integration tests with local database..."
	TEST_PG_DSN="postgres://postgres:postgres@localhost:54321/auth_test?sslmode=disable" go test -v ./internal/api/user/tests/integration_test.go -run TestIntegrationSuite

test-all:
	$(MAKE) test-unit
	$(MAKE) test-integration-local

# Установка зависимостей для тестов
install-test-deps:
	go mod download
	go install github.com/stretchr/testify@latest