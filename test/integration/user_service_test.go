package integration

import (
	"context"
	"fmt"
	"strings"
	"testing"

	pb "github.com/truongtu268/project_maker/proto/user"
)

func TestUserService_CreateUser(t *testing.T) {
	// Setup test environment
	testSetup := SetupIntegrationTest(t)
	defer testSetup.Cleanup()

	ctx := context.Background()

	// Test cases
	tests := []struct {
		name        string
		req         *pb.CreateUserRequest
		wantErr     bool
		errorMsg    string
		checkFields bool
	}{
		{
			name: "Success",
			req: &pb.CreateUserRequest{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password123",
				FullName: "Test User",
			},
			wantErr:     false,
			checkFields: true,
		},
		{
			name: "DuplicateUsername",
			req: &pb.CreateUserRequest{
				Username: "testuser",
				Email:    "different@example.com",
				Password: "password123",
				FullName: "Test User 2",
			},
			wantErr:  true,
			errorMsg: "username already taken",
		},
		{
			name: "DuplicateEmail",
			req: &pb.CreateUserRequest{
				Username: "differentuser",
				Email:    "test@example.com",
				Password: "password123",
				FullName: "Test User 3",
			},
			wantErr:  true,
			errorMsg: "email already registered",
		},
		{
			name: "EmptyUsername",
			req: &pb.CreateUserRequest{
				Username: "",
				Email:    "new@example.com",
				Password: "password123",
				FullName: "Test User 4",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := testSetup.GrpcClient.CreateUser(ctx, tt.req)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing %q but got %q", tt.errorMsg, err.Error())
				}
				return
			}
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if tt.checkFields {
				if resp.User.Id == 0 {
					t.Error("Expected user ID to be set")
				}
				if resp.User.Username != tt.req.Username {
					t.Errorf("Expected username %q but got %q", tt.req.Username, resp.User.Username)
				}
				if resp.User.Email != tt.req.Email {
					t.Errorf("Expected email %q but got %q", tt.req.Email, resp.User.Email)
				}
				if resp.User.FullName != tt.req.FullName {
					t.Errorf("Expected full name %q but got %q", tt.req.FullName, resp.User.FullName)
				}
				if resp.User.CreatedAt == "" {
					t.Error("Expected created_at to be set")
				}
				if resp.User.UpdatedAt == "" {
					t.Error("Expected updated_at to be set")
				}
			}
		})
	}
}

func TestUserService_GetUser(t *testing.T) {
	// Setup test environment
	testSetup := SetupIntegrationTest(t)
	defer testSetup.Cleanup()

	ctx := context.Background()

	// Create a test user first
	createResp, err := testSetup.GrpcClient.CreateUser(ctx, &pb.CreateUserRequest{
		Username: "getuser",
		Email:    "getuser@example.com",
		Password: "password123",
		FullName: "Get User",
	})
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	userID := createResp.User.Id

	// Test cases
	tests := []struct {
		name     string
		req      *pb.GetUserRequest
		wantErr  bool
		errorMsg string
	}{
		{
			name: "Success",
			req: &pb.GetUserRequest{
				Id: userID,
			},
			wantErr: false,
		},
		{
			name: "NotFound",
			req: &pb.GetUserRequest{
				Id: 999999, // Non-existent ID
			},
			wantErr:  true,
			errorMsg: "user not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := testSetup.GrpcClient.GetUser(ctx, tt.req)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing %q but got %q", tt.errorMsg, err.Error())
				}
				return
			}
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if resp.User.Id != userID {
				t.Errorf("Expected user ID %d but got %d", userID, resp.User.Id)
			}
			if resp.User.Username != "getuser" {
				t.Errorf("Expected username %q but got %q", "getuser", resp.User.Username)
			}
		})
	}
}

func TestUserService_UpdateUser(t *testing.T) {
	// Setup test environment
	testSetup := SetupIntegrationTest(t)
	defer testSetup.Cleanup()

	ctx := context.Background()

	// Create a test user first
	createResp, err := testSetup.GrpcClient.CreateUser(ctx, &pb.CreateUserRequest{
		Username: "updateuser",
		Email:    "updateuser@example.com",
		Password: "password123",
		FullName: "Update User",
	})
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	userID := createResp.User.Id

	// Create another user for testing duplicate checks
	_, err = testSetup.GrpcClient.CreateUser(ctx, &pb.CreateUserRequest{
		Username: "anotheruser",
		Email:    "another@example.com",
		Password: "password123",
		FullName: "Another User",
	})
	if err != nil {
		t.Fatalf("Failed to create another test user: %v", err)
	}

	// Helper function to create string pointer
	strPtr := func(s string) *string {
		return &s
	}

	// Test cases
	tests := []struct {
		name     string
		req      *pb.UpdateUserRequest
		wantErr  bool
		errorMsg string
		checkFn  func(t *testing.T, user *pb.User)
	}{
		{
			name: "UpdateFullName",
			req: &pb.UpdateUserRequest{
				Id:       userID,
				FullName: strPtr("Updated Name"),
			},
			wantErr: false,
			checkFn: func(t *testing.T, user *pb.User) {
				if user.FullName != "Updated Name" {
					t.Errorf("Expected full name %q but got %q", "Updated Name", user.FullName)
				}
			},
		},
		{
			name: "UpdateUsername",
			req: &pb.UpdateUserRequest{
				Id:       userID,
				Username: strPtr("updateduser"),
			},
			wantErr: false,
			checkFn: func(t *testing.T, user *pb.User) {
				if user.Username != "updateduser" {
					t.Errorf("Expected username %q but got %q", "updateduser", user.Username)
				}
			},
		},
		{
			name: "UpdateEmail",
			req: &pb.UpdateUserRequest{
				Id:    userID,
				Email: strPtr("updated@example.com"),
			},
			wantErr: false,
			checkFn: func(t *testing.T, user *pb.User) {
				if user.Email != "updated@example.com" {
					t.Errorf("Expected email %q but got %q", "updated@example.com", user.Email)
				}
			},
		},
		{
			name: "DuplicateUsername",
			req: &pb.UpdateUserRequest{
				Id:       userID,
				Username: strPtr("anotheruser"),
			},
			wantErr:  true,
			errorMsg: "username already taken",
		},
		{
			name: "DuplicateEmail",
			req: &pb.UpdateUserRequest{
				Id:    userID,
				Email: strPtr("another@example.com"),
			},
			wantErr:  true,
			errorMsg: "email already registered",
		},
		{
			name: "UserNotFound",
			req: &pb.UpdateUserRequest{
				Id:       999999, // Non-existent ID
				Username: strPtr("nonexistent"),
			},
			wantErr:  true,
			errorMsg: "user not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := testSetup.GrpcClient.UpdateUser(ctx, tt.req)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing %q but got %q", tt.errorMsg, err.Error())
				}
				return
			}
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if resp.User.Id != userID {
				t.Errorf("Expected user ID %d but got %d", userID, resp.User.Id)
			}

			if tt.checkFn != nil {
				tt.checkFn(t, resp.User)
			}
		})
	}
}

func TestUserService_DeleteUser(t *testing.T) {
	// Setup test environment
	testSetup := SetupIntegrationTest(t)
	defer testSetup.Cleanup()

	ctx := context.Background()

	// Create a test user first
	createResp, err := testSetup.GrpcClient.CreateUser(ctx, &pb.CreateUserRequest{
		Username: "deleteuser",
		Email:    "deleteuser@example.com",
		Password: "password123",
		FullName: "Delete User",
	})
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	userID := createResp.User.Id

	// Test cases
	tests := []struct {
		name     string
		req      *pb.DeleteUserRequest
		wantErr  bool
		errorMsg string
	}{
		{
			name: "Success",
			req: &pb.DeleteUserRequest{
				Id: userID,
			},
			wantErr: false,
		},
		{
			name: "AlreadyDeleted",
			req: &pb.DeleteUserRequest{
				Id: userID, // Should be deleted by now
			},
			wantErr:  true,
			errorMsg: "user not found",
		},
		{
			name: "NotFound",
			req: &pb.DeleteUserRequest{
				Id: 999999, // Non-existent ID
			},
			wantErr:  true,
			errorMsg: "user not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := testSetup.GrpcClient.DeleteUser(ctx, tt.req)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing %q but got %q", tt.errorMsg, err.Error())
				}
				return
			}
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if !resp.Success {
				t.Error("Expected success to be true")
			}

			// Verify user is deleted by trying to get it
			_, err = testSetup.GrpcClient.GetUser(ctx, &pb.GetUserRequest{Id: tt.req.Id})
			if err == nil {
				t.Errorf("User was not deleted as expected")
			}
		})
	}
}

func TestUserService_ListUsers(t *testing.T) {
	// Setup test environment
	testSetup := SetupIntegrationTest(t)
	defer testSetup.Cleanup()

	ctx := context.Background()

	// Clean up database first to ensure consistent state
	CleanupDatabase(t, testSetup.DB)

	// Create multiple test users
	for i := 1; i <= 5; i++ {
		username := fmt.Sprintf("listuser%d", i)
		email := fmt.Sprintf("listuser%d@example.com", i)
		_, err := testSetup.GrpcClient.CreateUser(ctx, &pb.CreateUserRequest{
			Username: username,
			Email:    email,
			Password: "password123",
			FullName: fmt.Sprintf("List User %d", i),
		})
		if err != nil {
			t.Fatalf("Failed to create test user %d: %v", i, err)
		}
	}

	// Test cases
	tests := []struct {
		name        string
		req         *pb.ListUsersRequest
		wantErr     bool
		checkCount  int32
		checkFields bool
	}{
		{
			name: "ListAll",
			req: &pb.ListUsersRequest{
				Page:     1,
				PageSize: 10,
			},
			wantErr:     false,
			checkCount:  5,
			checkFields: true,
		},
		{
			name: "Pagination",
			req: &pb.ListUsersRequest{
				Page:     1,
				PageSize: 3,
			},
			wantErr:     false,
			checkCount:  3, // Should return only 3 users
			checkFields: true,
		},
		{
			name: "SecondPage",
			req: &pb.ListUsersRequest{
				Page:     2,
				PageSize: 3,
			},
			wantErr:     false,
			checkCount:  2, // 5 total users, 3 on first page, 2 on second page
			checkFields: true,
		},
		{
			name: "EmptyPage",
			req: &pb.ListUsersRequest{
				Page:     3,
				PageSize: 3,
			},
			wantErr:    false,
			checkCount: 0, // Should have no users on the third page
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := testSetup.GrpcClient.ListUsers(ctx, tt.req)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if int32(len(resp.Users)) != tt.checkCount {
				t.Errorf("Expected %d users but got %d", tt.checkCount, len(resp.Users))
			}

			if resp.TotalCount != 5 {
				t.Errorf("Expected total count 5 but got %d", resp.TotalCount)
			}

			if tt.checkFields && len(resp.Users) > 0 {
				// Check that each user has the expected fields
				for _, user := range resp.Users {
					if user.Id == 0 {
						t.Error("Expected user ID to be set")
					}
					if !strings.HasPrefix(user.Username, "listuser") {
						t.Errorf("Unexpected username: %s", user.Username)
					}
					if !strings.HasSuffix(user.Email, "@example.com") {
						t.Errorf("Unexpected email: %s", user.Email)
					}
					if !strings.HasPrefix(user.FullName, "List User") {
						t.Errorf("Unexpected full name: %s", user.FullName)
					}
					if user.CreatedAt == "" {
						t.Error("Expected created_at to be set")
					}
					if user.UpdatedAt == "" {
						t.Error("Expected updated_at to be set")
					}
				}
			}
		})
	}
}
