// Package repos contains PostgreSQL implementations of domain repositories for admin management.
package repos

import (
	"context"
	"errors"
	"time"

	"prodlich/internal/domain/entity"
	"prodlich/internal/domain/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AdminRepo persists administrator users.
type AdminRepo struct {
	pool *pgxpool.Pool
}

// NewAdminRepo wires an AdminRepo against the given pool.
func NewAdminRepo(pool *pgxpool.Pool) *AdminRepo {
	return &AdminRepo{pool: pool}
}

const adminColumns = `id, username, password_hash, role, created_at, updated_at, last_login`

func scanAdmin(row pgx.Row) (entity.AdminUser, error) {
	var a entity.AdminUser
	if err := row.Scan(
		&a.ID, &a.Username, &a.PasswordHash, &a.Role,
		&a.CreatedAt, &a.UpdatedAt, &a.LastLogin,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.AdminUser{}, repository.ErrNotFound
		}
		return entity.AdminUser{}, err
	}
	return a, nil
}

// GetByUsername returns an admin by username. Returns repository.ErrNotFound when absent.
func (a *AdminRepo) GetByUsername(ctx context.Context, username string) (entity.AdminUser, error) {
	return scanAdmin(a.pool.QueryRow(ctx, `SELECT `+adminColumns+` FROM admin_users WHERE username = $1`, username))
}

// GetByID returns an admin by ID. Returns repository.ErrNotFound when absent.
func (a *AdminRepo) GetByID(ctx context.Context, id int) (entity.AdminUser, error) {
	return scanAdmin(a.pool.QueryRow(ctx, `SELECT `+adminColumns+` FROM admin_users WHERE id = $1`, id))
}

// GetAll returns every admin user in stable creation order.
func (a *AdminRepo) GetAll(ctx context.Context) ([]entity.AdminUser, error) {
	rows, err := a.pool.Query(ctx, `SELECT `+adminColumns+` FROM admin_users ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []entity.AdminUser{}
	for rows.Next() {
		u, err := scanAdmin(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

// CreateAdmin inserts a new admin and returns the new id.
//
// Zero CreatedAt / UpdatedAt fall through to the column DEFAULT NOW() so
// callers can omit timestamps and rely on the database clock.
func (a *AdminRepo) CreateAdmin(ctx context.Context, admin entity.AdminUser) (int, error) {
	if admin.CreatedAt.IsZero() {
		admin.CreatedAt = time.Now().UTC()
	}
	if admin.UpdatedAt.IsZero() {
		admin.UpdatedAt = admin.CreatedAt
	}
	const query = `INSERT INTO admin_users (username, password_hash, role, created_at, updated_at, last_login)
                   VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	var id int
	if err := a.pool.QueryRow(ctx, query,
		admin.Username,
		admin.PasswordHash,
		admin.Role,
		admin.CreatedAt,
		admin.UpdatedAt,
		admin.LastLogin,
	).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

// UpdateAdmin updates username/role/password (password optional). Empty
// PasswordHash leaves the existing hash untouched.
func (a *AdminRepo) UpdateAdmin(ctx context.Context, admin entity.AdminUser) error {
	const query = `UPDATE admin_users
                   SET username = $1,
                       role = $2,
                       password_hash = COALESCE(NULLIF($3, ''), password_hash),
                       updated_at = NOW()
                   WHERE id = $4`
	res, err := a.pool.Exec(ctx, query, admin.Username, admin.Role, admin.PasswordHash, admin.ID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return repository.ErrNotFound
	}
	return nil
}

// DeleteAdmin removes an admin by id.
func (a *AdminRepo) DeleteAdmin(ctx context.Context, id int) error {
	res, err := a.pool.Exec(ctx, `DELETE FROM admin_users WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return repository.ErrNotFound
	}
	return nil
}

// UpdateLastLogin sets last_login = NOW() for the given admin id.
func (a *AdminRepo) UpdateLastLogin(ctx context.Context, id int) error {
	const query = `UPDATE admin_users SET last_login = NOW() WHERE id = $1`
	res, err := a.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return repository.ErrNotFound
	}
	return nil
}


