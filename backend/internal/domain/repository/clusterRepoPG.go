package repository

import (
	"context"

	"prodlich/internal/domain/entity"
)

// ClusterRepo is the persistence contract for clusters and their targets.
//
// Cluster CRUD and target CRUD are exposed separately so the application
// layer can tweak individual targets without rewriting the whole cluster
// (and so a misbehaving target can be marked unhealthy by the health-poll
// goroutine cheaply).
type ClusterRepo interface {
	Create(ctx context.Context, cluster entity.Cluster) (int, error)
	GetByID(ctx context.Context, id int) (entity.Cluster, error)
	GetAll(ctx context.Context) ([]entity.Cluster, error)
	Update(ctx context.Context, cluster entity.Cluster) error
	Delete(ctx context.Context, id int) error

	AddTarget(ctx context.Context, target entity.ClusterTarget) (int, error)
	UpdateTarget(ctx context.Context, target entity.ClusterTarget) error
	DeleteTarget(ctx context.Context, targetID int) error
	ListTargets(ctx context.Context, clusterID int) ([]entity.ClusterTarget, error)
	SetTargetHealth(ctx context.Context, targetID int, healthy bool) error
}
