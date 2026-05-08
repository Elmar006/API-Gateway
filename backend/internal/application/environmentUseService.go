package application

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"prodlich/internal/domain/entity"
	repo "prodlich/internal/domain/repository"
)

// ErrInvalidEnvironment is returned by validateEnvironment when input is bad.
var ErrInvalidEnvironment = errors.New("invalid environment")

// Environment is the use case for managing route environments.
type Environment struct {
	repo repo.EnvironmentRepo
}

// NewEnvironmentUseCase wires an Environment use case.
func NewEnvironmentUseCase(r repo.EnvironmentRepo) *Environment {
	return &Environment{repo: r}
}

// List returns every environment.
func (e *Environment) List(ctx context.Context) ([]entity.Environment, error) {
	return e.repo.GetAll(ctx)
}

// Get returns a single environment by id.
func (e *Environment) Get(ctx context.Context, id int) (entity.Environment, error) {
	return e.repo.GetByID(ctx, id)
}

// Create validates and inserts an environment.
func (e *Environment) Create(ctx context.Context, env entity.Environment) (int, error) {
	if err := validateEnvironment(env); err != nil {
		return 0, err
	}
	return e.repo.Create(ctx, env)
}

// Update validates and persists changes to an existing environment.
func (e *Environment) Update(ctx context.Context, env entity.Environment) error {
	if err := validateEnvironment(env); err != nil {
		return err
	}
	return e.repo.Update(ctx, env)
}

// Delete removes an environment.
func (e *Environment) Delete(ctx context.Context, id int) error {
	return e.repo.Delete(ctx, id)
}

func validateEnvironment(env entity.Environment) error {
	if strings.TrimSpace(env.Name) == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidEnvironment)
	}
	if strings.ContainsAny(env.Name, " \t\n") {
		return fmt.Errorf("%w: name must not contain whitespace", ErrInvalidEnvironment)
	}
	return nil
}
