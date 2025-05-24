# User Management System

A complete user management system built with Go, gRPC, and PostgreSQL.

## Features

- User CRUD operations via gRPC and REST API
- REST API with OpenAPI/Swagger documentation
- PostgreSQL database with migrations
- Docker and Docker Compose support
- Command-line client for testing

## Project Structure

```
.
├── cmd
│   ├── client        # gRPC client implementation
│   └── server        # gRPC server implementation
├── config            # Application configuration
├── db
│   └── migrations    # Database migration files
├── internal
│   ├── domain        # Domain models
│   ├── repository    # Data access layer
│   └── service       # Business logic layer
└── proto             # Protocol Buffer definitions
```

## Prerequisites

- Go 1.16+
- Docker and Docker Compose
- PostgreSQL (or use the Docker Compose setup)
- Protocol Buffers compiler (protoc)

## Setup

1. Clone the repository:

```
git clone https://github.com/yourusername/user-management.git
cd user-management
```

2. Initialize Go modules and download dependencies:

```
make init
make deps
```

3. Generate Protocol Buffer code:

```
make proto
```

4. Start the PostgreSQL database using Docker Compose:

```
make docker-up
```

5. Run database migrations:

```
make migrate-up
```

6. Build and run the server:

```
make run-server
```

The server will start:
- gRPC server on port 50051
- REST API server on port 8080

## API Endpoints

The service provides both gRPC and REST API interfaces:

### REST API Endpoints

| Method | Endpoint             | Description                 |
|--------|----------------------|-----------------------------|
| POST   | /api/v1/users        | Create a new user           |
| GET    | /api/v1/users/{id}   | Get a user by ID            |
| PATCH  | /api/v1/users/{id}   | Update a user               |
| DELETE | /api/v1/users/{id}   | Delete a user               |
| GET    | /api/v1/users        | List users with pagination  |

## Using the Client

The client supports several commands for interacting with the user management service:

1. Create a new user:

```
./client create --username=john --email=john@example.com --password=secret --fullname="John Doe"
```

2. Get a user by ID:

```
./client get --id=1
```

3. Update a user:

```
./client update --id=1 --username=johndoe --email=johndoe@example.com
```

4. Delete a user:

```
./client delete --id=1
```

5. List users:

```
./client list --page=1 --pagesize=10
```

## Docker Deployment

To build and run the application using Docker:

1. Build the Docker image:

```
make docker-build
```

2. Start the services using Docker Compose:

```
make docker-up
```

3. To stop the services:

```
make docker-down
```

## Development

### Creating New Migrations

```
make migrate-create
```

### Running Migrations

```
make migrate-up   # Apply pending migrations
make migrate-down # Rollback the last migration
```

### Accessing the Database Shell

```
make db-shell
```

## Testing

### Running Integration Tests

Integration tests verify the functionality of the entire system, including the gRPC API, service layer, and database interactions. They use Docker to spin up a PostgreSQL container for testing.

Prerequisites:
- Docker must be running on your system

To run the integration tests:

```
make test-integration
```

Or use the provided script:

```
./scripts/run_integration_tests.sh
```

The tests will:
1. Start a PostgreSQL container
2. Run migrations to set up the database schema
3. Start a gRPC server with an in-memory connection
4. Run tests against all user service methods
5. Clean up resources when done

## License

MIT 