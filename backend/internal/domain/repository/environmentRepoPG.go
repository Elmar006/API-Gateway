package repository

import (
	"context"

	"prodlich/internal/domain/entity"
)

// EnvironmentRepo is the persistence contract for environments.
type EnvironmentRepo interface {
	Create(ctx context.Context, env entity.Environment) (int, error)
	GetByID(ctx context.Context, id int) (entity.Environment, error)
	GetByName(ctx context.Context, name string) (entity.Environment, error)
	GetAll(ctx context.Context) ([]entity.Environment, error)
	Update(ctx context.Context, env entity.Environment) error
	Delete(ctx context.Context, id int) error
}
