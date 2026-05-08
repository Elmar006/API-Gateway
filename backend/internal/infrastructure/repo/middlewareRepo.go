package repos

import (
	"context"
	"errors"
	"time"

	"prodlich/internal/domain/entity"
	"prodlich/internal/domain/repository"

	"github.com/bytedance/sonic"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// MiddlewareRepo is the PostgreSQL implementation of repository.MiddlewareRepo.
type MiddlewareRepo struct {
	pool *pgxpool.Pool
}

// NewMiddlewareRepo constructs a MiddlewareRepo against the given pool.
func NewMiddlewareRepo(pool *pgxpool.Pool) *MiddlewareRepo {
	return &MiddlewareRepo{pool: pool}
}

const middlewareColumns = `id, name, kind, config, is_active, description, created_at, updated_at`

func scanMiddleware(row pgx.Row) (entity.Middleware, error) {
	var m entity.Middleware
	if err := row.Scan(
		&m.ID, &m.Name, &m.Kind, &m.Config, &m.IsActive, &m.Description,
		&m.CreatedAt, &m.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.Middleware{}, repository.ErrNotFound
		}
		return entity.Middleware{}, err
	}
	if len(m.Config) > 0 {
		if err := sonic.Unmarshal(m.Config, &m.ConfigJSON); err != nil {
			// Surface unparseable config but keep the raw bytes so callers
			// can decide what to do (e.g. fail validation, fall back).
			m.ConfigJSON = nil
		}
	}
	return m, nil
}

// Create inserts a new middleware and returns its id.
func (r *MiddlewareRepo) Create(ctx context.Context, mw entity.Middleware) (int, error) {
	const query = `INSERT INTO middlewares (name, kind, config, is_active, description)
                   VALUES ($1, $2, $3, $4, $5) RETURNING id`
	cfg := mw.Config
	if len(cfg) == 0 {
		cfg = []byte(`{}`)
	}
	var id int
	if err := r.pool.QueryRow(ctx, query, mw.Name, mw.Kind, cfg, mw.IsActive, mw.Description).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

// GetByID returns a middleware by id.
func (r *MiddlewareRepo) GetByID(ctx context.Context, id int) (entity.Middleware, error) {
	return scanMiddleware(r.pool.QueryRow(ctx, `SELECT `+middlewareColumns+` FROM middlewares WHERE id = $1`, id))
}

// GetAll returns every middleware in stable creation order.
func (r *MiddlewareRepo) GetAll(ctx context.Context) ([]entity.Middleware, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+middlewareColumns+` FROM middlewares ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []entity.Middleware{}
	for rows.Next() {
		m, err := scanMiddleware(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// Update mutates the middleware matching mw.ID.
func (r *MiddlewareRepo) Update(ctx context.Context, mw entity.Middleware) error {
	const query = `UPDATE middlewares
                   SET name = $1, kind = $2, config = $3, is_active = $4, description = $5, updated_at = $6
                   WHERE id = $7`
	cfg := mw.Config
	if len(cfg) == 0 {
		cfg = []byte(`{}`)
	}
	res, err := r.pool.Exec(ctx, query,
		mw.Name, mw.Kind, cfg, mw.IsActive, mw.Description, time.Now().UTC(), mw.ID,
	)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return repository.ErrNotFound
	}
	return nil
}

// Delete removes a middleware (its route attachments cascade by FK).
func (r *MiddlewareRepo) Delete(ctx context.Context, id int) error {
	res, err := r.pool.Exec(ctx, `DELETE FROM middlewares WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return repository.ErrNotFound
	}
	return nil
}

// Attach links middleware to a route at the given sort_order; if the link
// already exists, sort_order is updated in place.
func (r *MiddlewareRepo) Attach(ctx context.Context, routeID, middlewareID, sortOrder int) error {
	const query = `INSERT INTO route_middlewares (route_id, middleware_id, sort_order)
                   VALUES ($1, $2, $3)
                   ON CONFLICT (route_id, middleware_id) DO UPDATE SET sort_order = EXCLUDED.sort_order`
	_, err := r.pool.Exec(ctx, query, routeID, middlewareID, sortOrder)
	return err
}

// Detach removes a single (route, middleware) link.
func (r *MiddlewareRepo) Detach(ctx context.Context, routeID, middlewareID int) error {
	const query = `DELETE FROM route_middlewares WHERE route_id = $1 AND middleware_id = $2`
	res, err := r.pool.Exec(ctx, query, routeID, middlewareID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return repository.ErrNotFound
	}
	return nil
}

// ListByRoute returns the middlewares attached to a route in their declared order.
func (r *MiddlewareRepo) ListByRoute(ctx context.Context, routeID int) ([]entity.Middleware, error) {
	const query = `SELECT m.id, m.name, m.kind, m.config, m.is_active, m.description, m.created_at, m.updated_at
                   FROM middlewares m
                   INNER JOIN route_middlewares rm ON rm.middleware_id = m.id
                   WHERE rm.route_id = $1
                   ORDER BY rm.sort_order, m.id`
	rows, err := r.pool.Query(ctx, query, routeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []entity.Middleware{}
	for rows.Next() {
		m, err := scanMiddleware(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
