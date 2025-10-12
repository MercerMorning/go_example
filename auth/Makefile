LOCAL_BIN:=$(CURDIR)/bin

install-golangci-lint:
	GOBIN=$(LOCAL_BIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.61.0

lint:
	$(LOCAL_BIN)/golangci-lint run ./... --config .golangci.pipeline.yaml
	
install-deps:
	GOBIN=$(LOCAL_BIN) go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.34.2
	GOBIN=$(LOCAL_BIN) go install -mod=mod google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.4
	GOBIN=$(LOCAL_BIN) go install github.com/pressly/goose/v3/cmd/goose@v3.20.0
	GOBIN=$(LOCAL_BIN) go install github.com/envoyproxy/protoc-gen-validate@v1.0.4
	GOBIN=$(LOCAL_BIN) go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.20.0
	GOBIN=$(LOCAL_BIN) go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.20.0
	GOBIN=$(LOCAL_BIN) go install github.com/rakyll/statik@v0.1.7

generate-user-api:
	mkdir -p pkg/user_v1 pkg/swagger
	protoc --proto_path api/user_v1 --proto_path vendor.protogen \
	--go_out=pkg/user_v1 --go_opt=paths=source_relative \
	--plugin=protoc-gen-go=bin/protoc-gen-go \
	--go-grpc_out=pkg/user_v1 --go-grpc_opt=paths=source_relative \
	--plugin=protoc-gen-go-grpc=bin/protoc-gen-go-grpc \
	--grpc-gateway_out=pkg/user_v1 --grpc-gateway_opt=paths=source_relative \
	--plugin=protoc-gen-grpc-gateway=bin/protoc-gen-grpc-gateway \
	--openapiv2_out=allow_merge=true,merge_file_name=api:pkg/swagger \
	--plugin=protoc-gen-openapiv2=bin/protoc-gen-openapiv2 \
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

vendor-proto:
		@if [ ! -d vendor.protogen/validate ]; then \
			mkdir -p vendor.protogen/validate &&\
			git clone https://github.com/envoyproxy/protoc-gen-validate vendor.protogen/protoc-gen-validate &&\
			mv vendor.protogen/protoc-gen-validate/validate/*.proto vendor.protogen/validate &&\
			rm -rf vendor.protogen/protoc-gen-validate ;\
		fi
		@if [ ! -d vendor.protogen/google ]; then \
			git clone https://github.com/googleapis/googleapis vendor.protogen/googleapis &&\
			mkdir -p  vendor.protogen/google/ &&\
			mv vendor.protogen/googleapis/google/api vendor.protogen/google &&\
			rm -rf vendor.protogen/googleapis ;\
		fi
		@if [ ! -d vendor.protogen/protoc-gen-openapiv2 ]; then \
			mkdir -p vendor.protogen/protoc-gen-openapiv2/options &&\
			git clone https://github.com/grpc-ecosystem/grpc-gateway vendor.protogen/openapiv2 &&\
			mv vendor.protogen/openapiv2/protoc-gen-openapiv2/options/*.proto vendor.protogen/protoc-gen-openapiv2/options &&\
			rm -rf vendor.protogen/openapiv2 ;\
		fi