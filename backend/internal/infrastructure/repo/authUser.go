// Package repos contains PostgreSQL implementations of domain repositories for user management.
package repos

import (
	"context"
	"errors"
	"prodlich/internal/domain/entity"
	"prodlich/internal/domain/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserRepoPG implements the UserRepo interface for PostgreSQL.
type UserRepoPG struct {
	pool *pgxpool.Pool
}

func NewUserRepoPG(pool *pgxpool.Pool) *UserRepoPG {
	return &UserRepoPG{pool: pool}
}

func (r *UserRepoPG) Create(ctx context.Context, user entity.User) (int64, error) {
	query := `INSERT INTO users (name, email, password_hash, created_at)
              VALUES ($1, $2, $3, $4) RETURNING id`
	var id int64
	err := r.pool.QueryRow(ctx, query, user.Name, user.Email, user.Password, user.CreatedAt).Scan(&id)
	return id, err
}

func (r *UserRepoPG) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	query := `SELECT id, name, email, password_hash, created_at FROM users WHERE email = $1`
	var user entity.User
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Name, &user.Email, &user.Password, &user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepoPG) GetByID(ctx context.Context, id int64) (*entity.User, error) {
	query := `SELECT id, name, email, password_hash, created_at FROM users WHERE id = $1`
	var user entity.User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Name, &user.Email, &user.Password, &user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}
