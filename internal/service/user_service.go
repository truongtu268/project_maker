package service

import (
	"context"
	"errors"

	"github.com/user-management/internal/domain/user"
	"github.com/user-management/internal/repository"
)

// UserService is responsible for user-related business logic
type UserService struct {
	repo repository.UserRepository
}

// NewUserService creates a new user service
func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

// CreateUser creates a new user
func (s *UserService) CreateUser(ctx context.Context, username, email, password, fullName string) (*user.User, error) {
	// Check if user with same username or email already exists
	if _, err := s.repo.GetByUsername(ctx, username); err == nil {
		return nil, errors.New("username already taken")
	}

	if _, err := s.repo.GetByEmail(ctx, email); err == nil {
		return nil, errors.New("email already registered")
	}

	// Create new user
	newUser, err := user.NewUser(username, email, password, fullName)
	if err != nil {
		return nil, err
	}

	// Save to repository
	if err := s.repo.Create(ctx, newUser); err != nil {
		return nil, err
	}

	return newUser, nil
}

// GetUser retrieves a user by ID
func (s *UserService) GetUser(ctx context.Context, id int64) (*user.User, error) {
	return s.repo.GetByID(ctx, id)
}

// UpdateUser updates user details
func (s *UserService) UpdateUser(ctx context.Context, id int64, username, email, password, fullName *string) (*user.User, error) {
	// Get existing user
	existingUser, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if username is being changed and already exists
	if username != nil && *username != existingUser.Username {
		if _, err := s.repo.GetByUsername(ctx, *username); err == nil {
			return nil, errors.New("username already taken")
		}
		existingUser.Username = *username
	}

	// Check if email is being changed and already exists
	if email != nil && *email != existingUser.Email {
		if _, err := s.repo.GetByEmail(ctx, *email); err == nil {
			return nil, errors.New("email already registered")
		}
		existingUser.Email = *email
	}

	// Update password if provided
	if password != nil {
		hashedPassword, err := user.HashPassword(*password)
		if err != nil {
			return nil, err
		}
		existingUser.PasswordHash = hashedPassword
	}

	// Update full name if provided
	if fullName != nil {
		existingUser.FullName = *fullName
	}

	// Save changes
	if err := s.repo.Update(ctx, existingUser); err != nil {
		return nil, err
	}

	return existingUser, nil
}

// DeleteUser deletes a user by ID
func (s *UserService) DeleteUser(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

// ListUsers retrieves a paginated list of users
func (s *UserService) ListUsers(ctx context.Context, page, pageSize int) ([]*user.User, int, error) {
	// Calculate offset
	offset := (page - 1) * pageSize
	if offset < 0 {
		offset = 0
	}

	// Ensure page size is reasonable
	if pageSize <= 0 {
		pageSize = 10
	} else if pageSize > 100 {
		pageSize = 100
	}

	return s.repo.List(ctx, offset, pageSize)
}
