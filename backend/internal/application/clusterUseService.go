package application

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"prodlich/internal/domain/entity"
	repo "prodlich/internal/domain/repository"
)

// ErrInvalidCluster is returned for cluster validation failures.
var ErrInvalidCluster = errors.New("invalid cluster")

// validClusterStrategies enumerates accepted strategy values; the migration
// also enforces this with a CHECK constraint, but we validate up-front to give
// callers a clean error message.
var validClusterStrategies = map[string]struct{}{
	entity.ClusterStrategyRoundRobin: {},
	entity.ClusterStrategyWeighted:   {},
	entity.ClusterStrategyLeastConn:  {},
}

// Cluster is the use case for managing logical clusters of upstream targets.
type Cluster struct {
	repo repo.ClusterRepo
}

// NewClusterUseCase wires a Cluster use case.
func NewClusterUseCase(r repo.ClusterRepo) *Cluster {
	return &Cluster{repo: r}
}

// List returns every cluster (with targets).
func (c *Cluster) List(ctx context.Context) ([]entity.Cluster, error) {
	return c.repo.GetAll(ctx)
}

// Get returns a single cluster (with targets).
func (c *Cluster) Get(ctx context.Context, id int) (entity.Cluster, error) {
	return c.repo.GetByID(ctx, id)
}

// Create validates and inserts a new cluster.
func (c *Cluster) Create(ctx context.Context, cl entity.Cluster) (int, error) {
	if err := validateCluster(cl); err != nil {
		return 0, err
	}
	return c.repo.Create(ctx, cl)
}

// Update validates and persists changes to an existing cluster.
func (c *Cluster) Update(ctx context.Context, cl entity.Cluster) error {
	if err := validateCluster(cl); err != nil {
		return err
	}
	return c.repo.Update(ctx, cl)
}

// Delete removes a cluster (its targets cascade via FK).
func (c *Cluster) Delete(ctx context.Context, id int) error {
	return c.repo.Delete(ctx, id)
}

// AddTarget validates and appends a target to a cluster.
func (c *Cluster) AddTarget(ctx context.Context, target entity.ClusterTarget) (int, error) {
	if err := validateClusterTarget(target); err != nil {
		return 0, err
	}
	return c.repo.AddTarget(ctx, target)
}

// UpdateTarget mutates URL/weight of an existing target.
func (c *Cluster) UpdateTarget(ctx context.Context, target entity.ClusterTarget) error {
	if err := validateClusterTarget(target); err != nil {
		return err
	}
	return c.repo.UpdateTarget(ctx, target)
}

// RemoveTarget deletes a target from a cluster.
func (c *Cluster) RemoveTarget(ctx context.Context, targetID int) error {
	return c.repo.DeleteTarget(ctx, targetID)
}

// SetTargetHealth flips the is_healthy flag on a target.
func (c *Cluster) SetTargetHealth(ctx context.Context, targetID int, healthy bool) error {
	return c.repo.SetTargetHealth(ctx, targetID, healthy)
}

func validateCluster(cl entity.Cluster) error {
	if strings.TrimSpace(cl.Name) == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidCluster)
	}
	if _, ok := validClusterStrategies[cl.Strategy]; !ok {
		return fmt.Errorf("%w: strategy must be one of round_robin|weighted|least_conn", ErrInvalidCluster)
	}
	return nil
}

func validateClusterTarget(t entity.ClusterTarget) error {
	if strings.TrimSpace(t.URL) == "" {
		return fmt.Errorf("%w: target url is required", ErrInvalidCluster)
	}
	u, err := url.Parse(t.URL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("%w: target url must be absolute", ErrInvalidCluster)
	}
	if t.Weight < 0 {
		return fmt.Errorf("%w: weight must be >= 0", ErrInvalidCluster)
	}
	return nil
}
