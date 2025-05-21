package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/truongtu268/project_maker/config"
	"github.com/truongtu268/project_maker/internal/repository"
	"github.com/truongtu268/project_maker/internal/service"
	pb "github.com/truongtu268/project_maker/proto/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

// Custom logger middleware for HTTP requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf(
			"[%s] %s %s %s",
			r.Method,
			r.RequestURI,
			r.RemoteAddr,
			time.Since(start),
		)
	})
}

// CORS middleware
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
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

func startGRPCServer(cfg *config.Config, userService *service.UserService) (*grpc.Server, net.Listener, error) {
	// Start gRPC server
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.GRPCPort))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterUserServiceServer(grpcServer, &server{userService: userService})

	// Register reflection service on gRPC server for easier testing with grpcurl
	reflection.Register(grpcServer)

	go func() {
		log.Printf("Starting gRPC server on %s:%d", cfg.Server.Host, cfg.Server.GRPCPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	return grpcServer, lis, nil
}

func startHTTPServer(ctx context.Context, cfg *config.Config) (*http.Server, error) {
	mux := runtime.NewServeMux()

	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err := pb.RegisterUserServiceHandlerFromEndpoint(
		ctx,
		mux,
		fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.GRPCPort),
		opts,
	)
	if err != nil {
		return nil, err
	}

	// Create a custom HTTP router
	httpMux := http.NewServeMux()

	// Add the gRPC-Gateway mux to handle API requests
	httpMux.Handle("/api/", loggingMiddleware(corsMiddleware(mux)))

	// Add Swagger UI handler for API documentation
	httpMux.Handle("/swagger/", http.StripPrefix("/swagger/", http.FileServer(http.Dir("./third_party/OpenAPI/proto"))))

	// Add Swagger UI
	httpMux.Handle("/swagger-ui/", http.StripPrefix("/swagger-ui/", http.FileServer(http.Dir("./third_party/swagger-ui"))))

	// Create a new HTTP server
	httpServer := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.HTTPPort),
		Handler: httpMux,
	}

	// Start the HTTP server
	go func() {
		log.Printf("Starting HTTP server on %s:%d", cfg.Server.Host, cfg.Server.HTTPPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	return httpServer, nil
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

	// Create a context that can be canceled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start gRPC server
	grpcServer, _, err := startGRPCServer(cfg, userService)
	if err != nil {
		log.Fatalf("Failed to start gRPC server: %v", err)
	}
	defer grpcServer.Stop()

	// Start HTTP server with gRPC-Gateway
	httpServer, err := startHTTPServer(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}

	// Handle graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	<-sigCh
	log.Println("Received shutdown signal")

	// Gracefully stop the HTTP server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	// Stop the gRPC server
	grpcServer.GracefulStop()
	log.Println("Servers shutdown completed")
}
