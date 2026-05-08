// Package repository defines interfaces for data persistence and access patterns.
package repository

import (
	"context"
	"prodlich/internal/domain/entity"
)

// AdminRepo defines persistence operations for administrator accounts.
type AdminRepo interface {
	GetByUsername(ctx context.Context, username string) (entity.AdminUser, error)
	GetByID(ctx context.Context, id int) (entity.AdminUser, error)
	GetAll(ctx context.Context) ([]entity.AdminUser, error)
	CreateAdmin(ctx context.Context, admin entity.AdminUser) (int, error)
	UpdateAdmin(ctx context.Context, admin entity.AdminUser) error
	DeleteAdmin(ctx context.Context, id int) error
	UpdateLastLogin(ctx context.Context, id int) error
}
