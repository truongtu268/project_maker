package integration

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/truongtu268/project_maker/internal/repository"
	"github.com/truongtu268/project_maker/internal/service"
	pb "github.com/truongtu268/project_maker/proto/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

// server is the gRPC server implementation for tests
type server struct {
	pb.UnimplementedUserServiceServer
	userService *service.UserService
}

// CreateUser implements the CreateUser RPC method
func (s *server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.UserResponse, error) {
	user, err := s.userService.CreateUser(ctx, req.Username, req.Email, req.Password, req.FullName)
	if err != nil {
		return nil, err
	}

	return &pb.UserResponse{
		User: &pb.User{
			Id:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			FullName:  user.FullName,
			CreatedAt: user.CreatedAt.Format(time.RFC3339),
			UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}

// GetUser implements the GetUser RPC method
func (s *server) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.UserResponse, error) {
	user, err := s.userService.GetUser(ctx, req.Id)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, fmt.Errorf("user not found with ID %d", req.Id)
		}
		return nil, err
	}

	return &pb.UserResponse{
		User: &pb.User{
			Id:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			FullName:  user.FullName,
			CreatedAt: user.CreatedAt.Format(time.RFC3339),
			UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}

// UpdateUser implements the UpdateUser RPC method
func (s *server) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UserResponse, error) {
	var username, email, password, fullName *string

	if req.Username != nil && *req.Username != "" {
		username = req.Username
	}
	if req.Email != nil && *req.Email != "" {
		email = req.Email
	}
	if req.Password != nil && *req.Password != "" {
		password = req.Password
	}
	if req.FullName != nil && *req.FullName != "" {
		fullName = req.FullName
	}

	user, err := s.userService.UpdateUser(ctx, req.Id, username, email, password, fullName)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, fmt.Errorf("user not found with ID %d", req.Id)
		}
		return nil, err
	}

	return &pb.UserResponse{
		User: &pb.User{
			Id:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			FullName:  user.FullName,
			CreatedAt: user.CreatedAt.Format(time.RFC3339),
			UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}

// DeleteUser implements the DeleteUser RPC method
func (s *server) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	err := s.userService.DeleteUser(ctx, req.Id)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, fmt.Errorf("user not found with ID %d", req.Id)
		}
		return nil, err
	}

	return &pb.DeleteUserResponse{
		Success: true,
	}, nil
}

// ListUsers implements the ListUsers RPC method
func (s *server) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	users, total, err := s.userService.ListUsers(ctx, int(req.Page), int(req.PageSize))
	if err != nil {
		return nil, err
	}

	var pbUsers []*pb.User
	for _, user := range users {
		pbUsers = append(pbUsers, &pb.User{
			Id:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			FullName:  user.FullName,
			CreatedAt: user.CreatedAt.Format(time.RFC3339),
			UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
		})
	}

	return &pb.ListUsersResponse{
		Users:      pbUsers,
		TotalCount: int32(total),
	}, nil
}

const bufSize = 1024 * 1024

var lis *bufconn.Listener

// Setup for testing
type TestSetup struct {
	Pool        *dockertest.Pool
	Resource    *dockertest.Resource
	DB          *sqlx.DB
	GrpcClient  pb.UserServiceClient
	GrpcServer  *grpc.Server
	Conn        *grpc.ClientConn
	UserService *service.UserService
	Cleanup     func()
}

// bufDialer is a helper function for bufconn
func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

// SetupIntegrationTest sets up everything needed for testing
func SetupIntegrationTest(t *testing.T) *TestSetup {
	t.Helper()

	// Create a buffer connection for grpc
	lis = bufconn.Listen(bufSize)

	var (
		db       *sql.DB
		pool     *dockertest.Pool
		resource *dockertest.Resource
		err      error
	)

	// Check if we're running in CI (GitHub Actions)
	if os.Getenv("CI") == "true" {
		// In CI, use the PostgreSQL service defined in GitHub Actions
		db, err = sql.Open("postgres", "postgres://postgres:postgres@localhost:5432/user_management_test?sslmode=disable")
		if err != nil {
			t.Fatalf("Could not connect to postgres: %s", err)
		}

		// Retry connection to handle potential delays in service startup
		for i := 0; i < 10; i++ {
			err = db.Ping()
			if err == nil {
				break
			}
			time.Sleep(time.Second)
		}
		if err != nil {
			t.Fatalf("Could not connect to postgres after retries: %s", err)
		}
	} else {
		// Not in CI, use dockertest to create a container
		pool, err = dockertest.NewPool("")
		if err != nil {
			t.Fatalf("Could not connect to docker: %s", err)
		}

		// Set up postgres container
		resource, err = pool.RunWithOptions(&dockertest.RunOptions{
			Repository: "postgres",
			Tag:        "15-alpine",
			Env: []string{
				"POSTGRES_PASSWORD=postgres",
				"POSTGRES_USER=postgres",
				"POSTGRES_DB=user_management_test",
			},
		}, func(config *docker.HostConfig) {
			// Set AutoRemove to true to remove container on stop
			config.AutoRemove = true
			config.RestartPolicy = docker.RestartPolicy{Name: "no"}
		})
		if err != nil {
			t.Fatalf("Could not start resource: %s", err)
		}

		// Expire the container after 5 minutes
		_ = resource.Expire(300)

		// Set up test database connection with retry
		if err = pool.Retry(func() error {
			var err error
			db, err = sql.Open("postgres", fmt.Sprintf("postgres://postgres:postgres@localhost:%s/user_management_test?sslmode=disable", resource.GetPort("5432/tcp")))
			if err != nil {
				return err
			}
			return db.Ping()
		}); err != nil {
			t.Fatalf("Could not connect to docker: %s", err)
		}
	}

	// Run migrations on test database
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		t.Fatalf("Could not create migration driver: %s", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://../../db/migrations",
		"user_management_test",
		driver,
	)
	if err != nil {
		t.Fatalf("Could not create migration instance: %s", err)
	}

	// Apply all up migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("Failed to run migrations: %s", err)
	}

	// Create sqlx DB
	dbx := sqlx.NewDb(db, "postgres")

	// Set up repository and service layers
	userRepo := repository.NewPostgresUserRepository(dbx)
	userService := service.NewUserService(userRepo)

	// Setup gRPC server
	grpcServer := grpc.NewServer()
	pb.RegisterUserServiceServer(grpcServer, &server{userService: userService})
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Connect to the server
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}

	client := pb.NewUserServiceClient(conn)

	// Cleanup function to be called at the end of the test
	cleanup := func() {
		conn.Close()
		grpcServer.Stop()
		dbx.Close()
		db.Close()

		// If not in CI and we have a docker resource, purge it
		if os.Getenv("CI") != "true" && pool != nil && resource != nil {
			if err := pool.Purge(resource); err != nil {
				log.Fatalf("Could not purge resource: %s", err)
			}
		}
	}

	setup := &TestSetup{
		DB:          dbx,
		GrpcClient:  client,
		GrpcServer:  grpcServer,
		Conn:        conn,
		UserService: userService,
		Cleanup:     cleanup,
	}

	// Only set docker resources if we're not in CI
	if os.Getenv("CI") != "true" {
		setup.Pool = pool
		setup.Resource = resource
	}

	return setup
}

// CleanupDatabase cleans up the database between tests
func CleanupDatabase(t *testing.T, db *sqlx.DB) {
	t.Helper()

	// Clear all users from the database
	_, err := db.Exec("DELETE FROM users")
	if err != nil {
		t.Fatalf("Failed to clean up database: %v", err)
	}
}
