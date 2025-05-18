package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/user-management/config"
	pb "github.com/user-management/proto/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Command line flags
	createCmd := flag.NewFlagSet("create", flag.ExitOnError)
	createUsername := createCmd.String("username", "", "Username for the new user")
	createEmail := createCmd.String("email", "", "Email for the new user")
	createPassword := createCmd.String("password", "", "Password for the new user")
	createFullName := createCmd.String("fullname", "", "Full name for the new user")

	getCmd := flag.NewFlagSet("get", flag.ExitOnError)
	getUserID := getCmd.Int64("id", 0, "ID of the user to retrieve")

	updateCmd := flag.NewFlagSet("update", flag.ExitOnError)
	updateUserID := updateCmd.Int64("id", 0, "ID of the user to update")
	updateUsername := updateCmd.String("username", "", "New username (optional)")
	updateEmail := updateCmd.String("email", "", "New email (optional)")
	updatePassword := updateCmd.String("password", "", "New password (optional)")
	updateFullName := updateCmd.String("fullname", "", "New full name (optional)")

	deleteCmd := flag.NewFlagSet("delete", flag.ExitOnError)
	deleteUserID := deleteCmd.Int64("id", 0, "ID of the user to delete")

	listCmd := flag.NewFlagSet("list", flag.ExitOnError)
	listPage := listCmd.Int("page", 1, "Page number")
	listPageSize := listCmd.Int("pagesize", 10, "Page size")

	// Check if a command was provided
	if len(os.Args) < 2 {
		fmt.Println("Expected 'create', 'get', 'update', 'delete', or 'list' subcommand")
		os.Exit(1)
	}

	// Load configuration
	cfg := config.New()

	// Set up a connection to the server
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewUserServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Parse and execute the appropriate command
	switch os.Args[1] {
	case "create":
		err = createCmd.Parse(os.Args[2:])
		if err != nil {
			log.Fatalf("Failed to parse create command: %v", err)
		}
		if createCmd.Parsed() {
			if *createUsername == "" || *createEmail == "" || *createPassword == "" || *createFullName == "" {
				createCmd.PrintDefaults()
				os.Exit(1)
			}
			createUser(ctx, client, *createUsername, *createEmail, *createPassword, *createFullName)
		}

	case "get":
		err = getCmd.Parse(os.Args[2:])
		if err != nil {
			log.Fatalf("Failed to parse get command: %v", err)
		}
		if getCmd.Parsed() {
			if *getUserID <= 0 {
				getCmd.PrintDefaults()
				os.Exit(1)
			}
			getUser(ctx, client, *getUserID)
		}

	case "update":
		err = updateCmd.Parse(os.Args[2:])
		if err != nil {
			log.Fatalf("Failed to parse update command: %v", err)
		}
		if updateCmd.Parsed() {
			if *updateUserID <= 0 {
				updateCmd.PrintDefaults()
				os.Exit(1)
			}
			updateUser(ctx, client, *updateUserID, *updateUsername, *updateEmail, *updatePassword, *updateFullName)
		}

	case "delete":
		err = deleteCmd.Parse(os.Args[2:])
		if err != nil {
			log.Fatalf("Failed to parse delete command: %v", err)
		}
		if deleteCmd.Parsed() {
			if *deleteUserID <= 0 {
				deleteCmd.PrintDefaults()
				os.Exit(1)
			}
			deleteUser(ctx, client, *deleteUserID)
		}

	case "list":
		err = listCmd.Parse(os.Args[2:])
		if err != nil {
			log.Fatalf("Failed to parse list command: %v", err)
		}
		if listCmd.Parsed() {
			listUsers(ctx, client, *listPage, *listPageSize)
		}

	default:
		fmt.Println("Expected 'create', 'get', 'update', 'delete', or 'list' subcommand")
		os.Exit(1)
	}
}

func createUser(ctx context.Context, client pb.UserServiceClient, username, email, password, fullName string) {
	resp, err := client.CreateUser(ctx, &pb.CreateUserRequest{
		Username: username,
		Email:    email,
		Password: password,
		FullName: fullName,
	})
	if err != nil {
		log.Fatalf("Could not create user: %v", err)
	}
	log.Printf("User created: %v", resp.User)
}

func getUser(ctx context.Context, client pb.UserServiceClient, id int64) {
	resp, err := client.GetUser(ctx, &pb.GetUserRequest{Id: id})
	if err != nil {
		log.Fatalf("Could not get user: %v", err)
	}
	log.Printf("User: %v", resp.User)
}

func updateUser(ctx context.Context, client pb.UserServiceClient, id int64, username, email, password, fullName string) {
	resp, err := client.UpdateUser(ctx, &pb.UpdateUserRequest{
		Id:       id,
		Username: username,
		Email:    email,
		Password: password,
		FullName: fullName,
	})
	if err != nil {
		log.Fatalf("Could not update user: %v", err)
	}
	log.Printf("User updated: %v", resp.User)
}

func deleteUser(ctx context.Context, client pb.UserServiceClient, id int64) {
	resp, err := client.DeleteUser(ctx, &pb.DeleteUserRequest{Id: id})
	if err != nil {
		log.Fatalf("Could not delete user: %v", err)
	}
	log.Printf("User deleted, success: %v", resp.Success)
}

func listUsers(ctx context.Context, client pb.UserServiceClient, page, pageSize int) {
	resp, err := client.ListUsers(ctx, &pb.ListUsersRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
	})
	if err != nil {
		log.Fatalf("Could not list users: %v", err)
	}
	log.Printf("Total users: %d", resp.TotalCount)
	for i, user := range resp.Users {
		log.Printf("[%d] %v", i+1, user)
	}
}
