package application

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"prodlich/internal/domain/entity"
	repo "prodlich/internal/domain/repository"

	"golang.org/x/crypto/bcrypt"
)

// ErrInvalidUser is returned for user validation failures.
var ErrInvalidUser = errors.New("invalid user")

// validRoles enumerates the roles enforced by the role check constraint on
// admin_users.role.
var validRoles = map[string]struct{}{
	entity.RoleAdmin:  {},
	entity.RoleEditor: {},
	entity.RoleViewer: {},
}

// User is the use case for managing administrator accounts.
type User struct {
	repo repo.AdminRepo
}

// NewUserUseCase wires a User use case against the admin repo.
func NewUserUseCase(r repo.AdminRepo) *User {
	return &User{repo: r}
}

// List returns every administrator.
func (u *User) List(ctx context.Context) ([]entity.AdminUser, error) {
	return u.repo.GetAll(ctx)
}

// Get returns a single administrator by id.
func (u *User) Get(ctx context.Context, id int) (entity.AdminUser, error) {
	return u.repo.GetByID(ctx, id)
}

// Create hashes the password, validates inputs, and inserts a new admin.
func (u *User) Create(ctx context.Context, user entity.AdminUser, password string) (int, error) {
	if err := validateUser(user); err != nil {
		return 0, err
	}
	if len(password) < 8 {
		return 0, fmt.Errorf("%w: password must be at least 8 chars", ErrInvalidUser)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, fmt.Errorf("hash password: %w", err)
	}
	user.PasswordHash = string(hash)
	return u.repo.CreateAdmin(ctx, user)
}

// Update mutates username/role and, optionally, the password (empty leaves the
// stored hash intact).
func (u *User) Update(ctx context.Context, user entity.AdminUser, password string) error {
	if err := validateUser(user); err != nil {
		return err
	}
	if password != "" {
		if len(password) < 8 {
			return fmt.Errorf("%w: password must be at least 8 chars", ErrInvalidUser)
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("hash password: %w", err)
		}
		user.PasswordHash = string(hash)
	}
	return u.repo.UpdateAdmin(ctx, user)
}

// Delete removes an administrator by id.
func (u *User) Delete(ctx context.Context, id int) error {
	return u.repo.DeleteAdmin(ctx, id)
}

func validateUser(user entity.AdminUser) error {
	if strings.TrimSpace(user.Username) == "" {
		return fmt.Errorf("%w: username is required", ErrInvalidUser)
	}
	if _, ok := validRoles[user.Role]; !ok {
		return fmt.Errorf("%w: role must be admin|editor|viewer", ErrInvalidUser)
	}
	return nil
}
