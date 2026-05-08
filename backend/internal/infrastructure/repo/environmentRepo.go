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

// EnvironmentRepo is the PostgreSQL implementation of repository.EnvironmentRepo.
type EnvironmentRepo struct {
	pool *pgxpool.Pool
}

// NewEnvironmentRepo constructs an EnvironmentRepo against the given pool.
func NewEnvironmentRepo(pool *pgxpool.Pool) *EnvironmentRepo {
	return &EnvironmentRepo{pool: pool}
}

const envColumns = `id, name, base_domain, color, description, created_at, updated_at`

func scanEnv(row pgx.Row) (entity.Environment, error) {
	var e entity.Environment
	if err := row.Scan(
		&e.ID, &e.Name, &e.BaseDomain, &e.Color, &e.Description,
		&e.CreatedAt, &e.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.Environment{}, repository.ErrNotFound
		}
		return entity.Environment{}, err
	}
	return e, nil
}

// Create inserts a new environment and returns its id.
func (r *EnvironmentRepo) Create(ctx context.Context, env entity.Environment) (int, error) {
	const query = `INSERT INTO environments (name, base_domain, color, description)
                   VALUES ($1, $2, $3, $4) RETURNING id`
	var id int
	if err := r.pool.QueryRow(ctx, query,
		env.Name, env.BaseDomain, env.Color, env.Description,
	).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

// GetByID returns an environment by id.
func (r *EnvironmentRepo) GetByID(ctx context.Context, id int) (entity.Environment, error) {
	return scanEnv(r.pool.QueryRow(ctx, `SELECT `+envColumns+` FROM environments WHERE id = $1`, id))
}

// GetByName returns an environment by its unique name.
func (r *EnvironmentRepo) GetByName(ctx context.Context, name string) (entity.Environment, error) {
	return scanEnv(r.pool.QueryRow(ctx, `SELECT `+envColumns+` FROM environments WHERE name = $1`, name))
}

// GetAll returns every environment in stable creation order.
func (r *EnvironmentRepo) GetAll(ctx context.Context) ([]entity.Environment, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+envColumns+` FROM environments ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []entity.Environment{}
	for rows.Next() {
		e, err := scanEnv(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// Update mutates the environment matching env.ID.
func (r *EnvironmentRepo) Update(ctx context.Context, env entity.Environment) error {
	const query = `UPDATE environments
                   SET name = $1, base_domain = $2, color = $3, description = $4, updated_at = $5
                   WHERE id = $6`
	res, err := r.pool.Exec(ctx, query,
		env.Name, env.BaseDomain, env.Color, env.Description, time.Now().UTC(), env.ID,
	)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return repository.ErrNotFound
	}
	return nil
}

// Delete removes an environment.
func (r *EnvironmentRepo) Delete(ctx context.Context, id int) error {
	res, err := r.pool.Exec(ctx, `DELETE FROM environments WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return repository.ErrNotFound
	}
	return nil
}
