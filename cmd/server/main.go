package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/truongtu268/project_maker/config"
	"github.com/truongtu268/project_maker/internal/repository"
	"github.com/truongtu268/project_maker/internal/service"
	pb "github.com/truongtu268/project_maker/proto/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// server is the gRPC server implementation
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

	if req.Username != "" {
		username = &req.Username
	}
	if req.Email != "" {
		email = &req.Email
	}
	if req.Password != "" {
		password = &req.Password
	}
	if req.FullName != "" {
		fullName = &req.FullName
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

func runMigrations(db *sql.DB, cfg *config.Config) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://db/migrations",
		cfg.Database.DBName,
		driver,
	)
	if err != nil {
		return err
	}

	// Apply all up migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}

func main() {
	// Load configuration
	cfg := config.New()

	// Set up database connection
	db, err := sql.Open("postgres", cfg.Database.DSN())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := runMigrations(db, cfg); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Create sqlx DB
	dbx := sqlx.NewDb(db, "postgres")

	// Set up repository and service
	userRepo := repository.NewPostgresUserRepository(dbx)
	userService := service.NewUserService(userRepo)

	// Start gRPC server
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterUserServiceServer(grpcServer, &server{userService: userService})

	// Register reflection service on gRPC server for easier testing with grpcurl
	reflection.Register(grpcServer)

	log.Printf("Starting gRPC server on %s:%d", cfg.Server.Host, cfg.Server.Port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
