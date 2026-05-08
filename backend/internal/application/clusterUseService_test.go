package application

import (
	"context"
	"errors"
	"testing"

	"prodlich/internal/domain/entity"
	"prodlich/internal/domain/repository"
)

type fakeClusterRepo struct {
	clusters map[int]entity.Cluster
	targets  map[int]entity.ClusterTarget
	nextC    int
	nextT    int
}

func newFakeClusterRepo() *fakeClusterRepo {
	return &fakeClusterRepo{clusters: map[int]entity.Cluster{}, targets: map[int]entity.ClusterTarget{}}
}

func (f *fakeClusterRepo) Create(_ context.Context, c entity.Cluster) (int, error) {
	f.nextC++
	c.ID = f.nextC
	f.clusters[c.ID] = c
	return c.ID, nil
}
func (f *fakeClusterRepo) GetByID(_ context.Context, id int) (entity.Cluster, error) {
	if v, ok := f.clusters[id]; ok {
		return v, nil
	}
	return entity.Cluster{}, repository.ErrNotFound
}
func (f *fakeClusterRepo) GetAll(_ context.Context) ([]entity.Cluster, error) {
	out := make([]entity.Cluster, 0, len(f.clusters))
	for _, v := range f.clusters {
		out = append(out, v)
	}
	return out, nil
}
func (f *fakeClusterRepo) Update(_ context.Context, c entity.Cluster) error {
	if _, ok := f.clusters[c.ID]; !ok {
		return repository.ErrNotFound
	}
	f.clusters[c.ID] = c
	return nil
}
func (f *fakeClusterRepo) Delete(_ context.Context, id int) error {
	if _, ok := f.clusters[id]; !ok {
		return repository.ErrNotFound
	}
	delete(f.clusters, id)
	return nil
}
func (f *fakeClusterRepo) AddTarget(_ context.Context, t entity.ClusterTarget) (int, error) {
	f.nextT++
	t.ID = f.nextT
	f.targets[t.ID] = t
	return t.ID, nil
}
func (f *fakeClusterRepo) UpdateTarget(_ context.Context, t entity.ClusterTarget) error {
	if _, ok := f.targets[t.ID]; !ok {
		return repository.ErrNotFound
	}
	f.targets[t.ID] = t
	return nil
}
func (f *fakeClusterRepo) DeleteTarget(_ context.Context, id int) error {
	if _, ok := f.targets[id]; !ok {
		return repository.ErrNotFound
	}
	delete(f.targets, id)
	return nil
}
func (f *fakeClusterRepo) ListTargets(_ context.Context, clusterID int) ([]entity.ClusterTarget, error) {
	out := []entity.ClusterTarget{}
	for _, t := range f.targets {
		if t.ClusterID == clusterID {
			out = append(out, t)
		}
	}
	return out, nil
}
func (f *fakeClusterRepo) SetTargetHealth(_ context.Context, id int, healthy bool) error {
	t, ok := f.targets[id]
	if !ok {
		return repository.ErrNotFound
	}
	t.IsHealthy = healthy
	f.targets[id] = t
	return nil
}

func TestClusterValidate(t *testing.T) {
	uc := NewClusterUseCase(newFakeClusterRepo())
	if _, err := uc.Create(context.Background(), entity.Cluster{Name: "", Strategy: entity.ClusterStrategyRoundRobin}); !errors.Is(err, ErrInvalidCluster) {
		t.Fatalf("expected invalid for empty name")
	}
	if _, err := uc.Create(context.Background(), entity.Cluster{Name: "x", Strategy: "bogus"}); !errors.Is(err, ErrInvalidCluster) {
		t.Fatalf("expected invalid for bad strategy")
	}
	if _, err := uc.AddTarget(context.Background(), entity.ClusterTarget{ClusterID: 1, URL: "not a url"}); !errors.Is(err, ErrInvalidCluster) {
		t.Fatalf("expected invalid for bad url")
	}
}

func TestClusterTargetCRUD(t *testing.T) {
	uc := NewClusterUseCase(newFakeClusterRepo())
	cid, err := uc.Create(context.Background(), entity.Cluster{Name: "primary", Strategy: entity.ClusterStrategyWeighted})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	tid, err := uc.AddTarget(context.Background(), entity.ClusterTarget{ClusterID: cid, URL: "http://up:80", Weight: 5})
	if err != nil {
		t.Fatalf("add target: %v", err)
	}
	if err := uc.SetTargetHealth(context.Background(), tid, false); err != nil {
		t.Fatalf("set health: %v", err)
	}
	if err := uc.RemoveTarget(context.Background(), tid); err != nil {
		t.Fatalf("remove: %v", err)
	}
}
