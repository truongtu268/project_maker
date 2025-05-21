.PHONY: proto build run clean docker-build docker-up docker-down migrate-up migrate-down

# Variables
BINARY_NAME_SERVER=server
BINARY_NAME_CLIENT=client
PROTO_DIR=proto
PROTO_OUT=proto

# Go build flags
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Docker compose commands
DOCKER_COMPOSE=docker-compose
DOCKER_COMPOSE_FILE=docker-compose.yml

# Protoc command
PROTOC=protoc

# Generate protobuf code
proto:
	$(PROTOC) -I . \
		-I third_party \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=. --grpc-gateway_opt=paths=source_relative \
		--openapiv2_out=./third_party/OpenAPI \
		$(PROTO_DIR)/user/user.proto

# Install required tools
tools:
	$(GOGET) -u google.golang.org/protobuf/cmd/protoc-gen-go
	$(GOGET) -u google.golang.org/grpc/cmd/protoc-gen-go-grpc
	$(GOGET) -u github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway
	$(GOGET) -u github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2
	$(GOGET) -u github.com/golang-migrate/migrate/v4/cmd/migrate
	$(GOGET) -u github.com/golang/protobuf/protoc-gen-go

# Download protobuf dependencies
proto-deps:
	mkdir -p third_party/OpenAPI
	mkdir -p third_party/google/api
	
	# Download necessary proto files
	curl -sSL https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/annotations.proto > third_party/google/api/annotations.proto
	curl -sSL https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/http.proto > third_party/google/api/http.proto
	curl -sSL https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/field_behavior.proto > third_party/google/api/field_behavior.proto
	
	# Download Swagger annotations
	curl -sSL https://raw.githubusercontent.com/grpc-ecosystem/grpc-gateway/main/protoc-gen-openapiv2/options/annotations.proto > third_party/protoc-gen-openapiv2/options/annotations.proto
	curl -sSL https://raw.githubusercontent.com/grpc-ecosystem/grpc-gateway/main/protoc-gen-openapiv2/options/openapiv2.proto > third_party/protoc-gen-openapiv2/options/openapiv2.proto

# Build the project
build: proto
	$(GOBUILD) -o $(BINARY_NAME_SERVER) ./cmd/server
	$(GOBUILD) -o $(BINARY_NAME_CLIENT) ./cmd/client

# Run server
run-server: build
	./$(BINARY_NAME_SERVER)

# Run client
run-client: build
	./$(BINARY_NAME_CLIENT)

# Clean build files
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME_SERVER)
	rm -f $(BINARY_NAME_CLIENT)

# Initialize Go modules
init:
	$(GOMOD) init github.com/truongtu268/project_maker
	$(GOMOD) tidy

# Update dependencies
deps:
	$(GOMOD) tidy

# Create a new migration file
migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir db/migrations -seq $$name

# Run database migrations up
migrate-up:
	migrate -path db/migrations -database "postgres://postgres:postgres@localhost:5432/user_management?sslmode=disable" up

# Run database migrations down
migrate-down:
	migrate -path db/migrations -database "postgres://postgres:postgres@localhost:5432/user_management?sslmode=disable" down

# Build Docker image
docker-build:
	docker build -t user-management .

# Start Docker containers
docker-up:
	$(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_FILE) up -d

# Stop Docker containers
docker-down:
	$(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_FILE) down

# Show Docker container logs
docker-logs:
	$(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_FILE) logs -f

# Access PostgreSQL database
db-shell:
	docker exec -it $$(docker-compose ps -q postgres) psql -U postgres -d user_management

# Check PostgreSQL status
postgres-check:
	./scripts/check_postgres.sh

# Clean up Docker resources (use with caution!)
docker-cleanup:
	./scripts/cleanup_docker.sh

# Run unit tests
test:
	$(GOCMD) test -v ./...

# Run integration tests
test-integration:
	$(GOCMD) test -v ./test/integration/...

# Default target
all: proto build 