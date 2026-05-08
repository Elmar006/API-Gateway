// Package repos contains PostgreSQL implementations of domain repositories for routes.
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

type RouteRepo struct {
	pool *pgxpool.Pool
}

func NewRouteRepo(pool *pgxpool.Pool) *RouteRepo {
	return &RouteRepo{pool: pool}
}

const routeColumns = `id, method, path_pattern, target_url, is_active, priority, rate_limit, require_auth,
                      environment_id, cluster_id, created_at, updated_at`

func scanRoute(row pgx.Row) (entity.Route, error) {
	var route entity.Route
	if err := row.Scan(
		&route.ID, &route.Method, &route.PathPattern, &route.TargetURL,
		&route.IsActive, &route.Priority, &route.RateLimit, &route.RequireAuth,
		&route.EnvironmentID, &route.ClusterID,
		&route.CreatedAt, &route.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.Route{}, repository.ErrNotFound
		}
		return entity.Route{}, err
	}
	return route, nil
}

func (r *RouteRepo) Create(ctx context.Context, route entity.Route) (int, error) {
	const query = `INSERT INTO routes
                       (method, path_pattern, target_url, is_active, priority, rate_limit, require_auth, environment_id, cluster_id)
                   VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`
	var id int
	if err := r.pool.QueryRow(ctx, query,
		route.Method,
		route.PathPattern,
		route.TargetURL,
		route.IsActive,
		route.Priority,
		route.RateLimit,
		route.RequireAuth,
		route.EnvironmentID,
		route.ClusterID,
	).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

// GetByID fetches a route by its ID from the database.
func (r *RouteRepo) GetByID(ctx context.Context, id int) (entity.Route, error) {
	return scanRoute(r.pool.QueryRow(ctx, `SELECT `+routeColumns+` FROM routes WHERE id = $1`, id))
}

func (r *RouteRepo) GetAll(ctx context.Context) ([]entity.Route, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+routeColumns+` FROM routes ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	routes := []entity.Route{}
	for rows.Next() {
		route, err := scanRoute(rows)
		if err != nil {
			return nil, err
		}
		routes = append(routes, route)
	}
	return routes, rows.Err()
}

func (r *RouteRepo) Update(ctx context.Context, route entity.Route) error {
	const query = `UPDATE routes
                   SET method = $1, path_pattern = $2, target_url = $3, is_active = $4,
                       priority = $5, rate_limit = $6, require_auth = $7,
                       environment_id = $8, cluster_id = $9, updated_at = $10
                   WHERE id = $11`
	route.UpdatedAt = time.Now()
	res, err := r.pool.Exec(ctx, query,
		route.Method,
		route.PathPattern,
		route.TargetURL,
		route.IsActive,
		route.Priority,
		route.RateLimit,
		route.RequireAuth,
		route.EnvironmentID,
		route.ClusterID,
		route.UpdatedAt,
		route.ID,
	)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (r *RouteRepo) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM routes WHERE id = $1`
	res, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if res.RowsAffected() == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (r *RouteRepo) ToggleActive(ctx context.Context, id int) error {
	query := `UPDATE routes SET is_active = NOT is_active, updated_at = NOW() WHERE id = $1`
	res, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if res.RowsAffected() == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (r *RouteRepo) GetActiveRoutes(ctx context.Context) ([]entity.Route, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+routeColumns+` FROM routes WHERE is_active = true ORDER BY priority DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var routes []entity.Route
	for rows.Next() {
		route, err := scanRoute(rows)
		if err != nil {
			return nil, err
		}
		routes = append(routes, route)
	}
	return routes, rows.Err()
}
