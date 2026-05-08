// Package repository defines interfaces for data persistence and access patterns.
package repository

import (
	"context"
	"prodlich/internal/domain/entity"
)

// UserRepo defines persistence operations for regular user accounts.
type UserRepo interface {
	Create(ctx context.Context, user entity.User) (int64, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	GetByID(ctx context.Context, id int64) (*entity.User, error)
}
