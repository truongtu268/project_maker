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
	$(PROTOC) --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		$(PROTO_DIR)/user/user.proto

# Install required tools
tools:
	$(GOGET) -u google.golang.org/protobuf/cmd/protoc-gen-go
	$(GOGET) -u google.golang.org/grpc/cmd/protoc-gen-go-grpc
	$(GOGET) -u github.com/golang-migrate/migrate/v4/cmd/migrate
	$(GOGET) -u github.com/golang/protobuf/protoc-gen-go

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
	$(GOMOD) init github.com/user-management
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

# Run unit tests
test:
	$(GOCMD) test -v ./...

# Run integration tests
test-integration:
	$(GOCMD) test -v ./test/integration/...

# Default target
all: proto build 