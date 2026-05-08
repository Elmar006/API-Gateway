// Package repository defines interfaces for route persistence.
package repository

import (
	"context"
	"prodlich/internal/domain/entity"
)

type RouteRepo interface {
	Create(ctx context.Context, route entity.Route) (int, error)
	GetByID(ctx context.Context, id int) (entity.Route, error)
	GetAll(ctx context.Context) ([]entity.Route, error)
	Update(ctx context.Context, route entity.Route) error
	Delete(ctx context.Context, id int) error
	ToggleActive(ctx context.Context, id int) error
	GetActiveRoutes(ctx context.Context) ([]entity.Route, error)
}
