# User Management System

A complete user management system built with Go, gRPC, and PostgreSQL.

## Features

- User CRUD operations via gRPC
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

## License

MIT 