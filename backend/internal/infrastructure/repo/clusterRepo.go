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

// ClusterRepo is the PostgreSQL implementation of repository.ClusterRepo.
type ClusterRepo struct {
	pool *pgxpool.Pool
}

// NewClusterRepo constructs a ClusterRepo against the given pool.
func NewClusterRepo(pool *pgxpool.Pool) *ClusterRepo {
	return &ClusterRepo{pool: pool}
}

const clusterColumns = `id, name, strategy, description, created_at, updated_at`
const targetColumns = `id, cluster_id, url, weight, is_healthy, last_check, created_at, updated_at`

func scanCluster(row pgx.Row) (entity.Cluster, error) {
	var c entity.Cluster
	if err := row.Scan(&c.ID, &c.Name, &c.Strategy, &c.Description, &c.CreatedAt, &c.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.Cluster{}, repository.ErrNotFound
		}
		return entity.Cluster{}, err
	}
	return c, nil
}

func scanTarget(row pgx.Row) (entity.ClusterTarget, error) {
	var t entity.ClusterTarget
	if err := row.Scan(
		&t.ID, &t.ClusterID, &t.URL, &t.Weight, &t.IsHealthy,
		&t.LastCheck, &t.CreatedAt, &t.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.ClusterTarget{}, repository.ErrNotFound
		}
		return entity.ClusterTarget{}, err
	}
	return t, nil
}

// Create inserts a new cluster (without targets) and returns its id.
func (r *ClusterRepo) Create(ctx context.Context, cluster entity.Cluster) (int, error) {
	const query = `INSERT INTO clusters (name, strategy, description)
                   VALUES ($1, $2, $3) RETURNING id`
	var id int
	if err := r.pool.QueryRow(ctx, query, cluster.Name, cluster.Strategy, cluster.Description).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

// GetByID returns a cluster with its targets eagerly loaded.
func (r *ClusterRepo) GetByID(ctx context.Context, id int) (entity.Cluster, error) {
	c, err := scanCluster(r.pool.QueryRow(ctx, `SELECT `+clusterColumns+` FROM clusters WHERE id = $1`, id))
	if err != nil {
		return entity.Cluster{}, err
	}
	c.Targets, err = r.ListTargets(ctx, id)
	if err != nil {
		return entity.Cluster{}, err
	}
	return c, nil
}

// GetAll returns every cluster, each with its targets eagerly loaded.
func (r *ClusterRepo) GetAll(ctx context.Context) ([]entity.Cluster, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+clusterColumns+` FROM clusters ORDER BY id`)
	if err != nil {
		return nil, err
	}
	clusters := []entity.Cluster{}
	for rows.Next() {
		c, err := scanCluster(rows)
		if err != nil {
			rows.Close()
			return nil, err
		}
		clusters = append(clusters, c)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i := range clusters {
		ts, err := r.ListTargets(ctx, clusters[i].ID)
		if err != nil {
			return nil, err
		}
		clusters[i].Targets = ts
	}
	return clusters, nil
}

// Update mutates name/strategy/description for the cluster.
func (r *ClusterRepo) Update(ctx context.Context, cluster entity.Cluster) error {
	const query = `UPDATE clusters SET name = $1, strategy = $2, description = $3, updated_at = $4 WHERE id = $5`
	res, err := r.pool.Exec(ctx, query, cluster.Name, cluster.Strategy, cluster.Description, time.Now().UTC(), cluster.ID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return repository.ErrNotFound
	}
	return nil
}

// Delete removes a cluster (its targets are cascade-deleted by FK).
func (r *ClusterRepo) Delete(ctx context.Context, id int) error {
	res, err := r.pool.Exec(ctx, `DELETE FROM clusters WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return repository.ErrNotFound
	}
	return nil
}

// AddTarget appends a target to a cluster and returns its id.
func (r *ClusterRepo) AddTarget(ctx context.Context, target entity.ClusterTarget) (int, error) {
	const query = `INSERT INTO cluster_targets (cluster_id, url, weight, is_healthy)
                   VALUES ($1, $2, $3, $4) RETURNING id`
	var id int
	if err := r.pool.QueryRow(ctx, query, target.ClusterID, target.URL, target.Weight, target.IsHealthy).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

// UpdateTarget mutates URL/weight for an existing target. Health is updated
// by SetTargetHealth and not via this method.
func (r *ClusterRepo) UpdateTarget(ctx context.Context, target entity.ClusterTarget) error {
	const query = `UPDATE cluster_targets SET url = $1, weight = $2, updated_at = $3 WHERE id = $4`
	res, err := r.pool.Exec(ctx, query, target.URL, target.Weight, time.Now().UTC(), target.ID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return repository.ErrNotFound
	}
	return nil
}

// DeleteTarget removes a single target from a cluster.
func (r *ClusterRepo) DeleteTarget(ctx context.Context, targetID int) error {
	res, err := r.pool.Exec(ctx, `DELETE FROM cluster_targets WHERE id = $1`, targetID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return repository.ErrNotFound
	}
	return nil
}

// ListTargets returns all targets for a cluster.
func (r *ClusterRepo) ListTargets(ctx context.Context, clusterID int) ([]entity.ClusterTarget, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+targetColumns+` FROM cluster_targets WHERE cluster_id = $1 ORDER BY id`, clusterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []entity.ClusterTarget{}
	for rows.Next() {
		t, err := scanTarget(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// SetTargetHealth flips the is_healthy flag and updates last_check. Used by
// the periodic health-poll goroutine.
func (r *ClusterRepo) SetTargetHealth(ctx context.Context, targetID int, healthy bool) error {
	const query = `UPDATE cluster_targets SET is_healthy = $1, last_check = $2, updated_at = $2 WHERE id = $3`
	res, err := r.pool.Exec(ctx, query, healthy, time.Now().UTC(), targetID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return repository.ErrNotFound
	}
	return nil
}
