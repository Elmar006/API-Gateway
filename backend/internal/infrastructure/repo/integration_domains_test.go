//go:build integration

package repos

import (
	"context"
	"testing"
	"time"

	"prodlich/internal/domain/entity"
	"prodlich/internal/domain/repository"
)

func TestEnvironmentRepoCRUD(t *testing.T) {
	pool, cleanup := startPostgres(t)
	defer cleanup()

	repo := NewEnvironmentRepo(pool)
	ctx := context.Background()

	// "production" auto-seeded by migration 000005.
	all, err := repo.GetAll(ctx)
	if err != nil {
		t.Fatalf("getall: %v", err)
	}
	if len(all) != 1 || all[0].Name != "production" {
		t.Fatalf("expected default production env, got %+v", all)
	}

	id, err := repo.Create(ctx, entity.Environment{Name: "staging", BaseDomain: "staging.example.com", Color: "#f59e0b"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	got, err := repo.GetByID(ctx, id)
	if err != nil || got.Name != "staging" {
		t.Fatalf("get: %v %+v", err, got)
	}
	got.Description = "test"
	if err := repo.Update(ctx, got); err != nil {
		t.Fatalf("update: %v", err)
	}
	if err := repo.Delete(ctx, id); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := repo.GetByID(ctx, id); err != repository.ErrNotFound {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestClusterRepoCRUDWithTargets(t *testing.T) {
	pool, cleanup := startPostgres(t)
	defer cleanup()

	repo := NewClusterRepo(pool)
	ctx := context.Background()

	cid, err := repo.Create(ctx, entity.Cluster{Name: "primary", Strategy: entity.ClusterStrategyWeighted})
	if err != nil {
		t.Fatalf("create cluster: %v", err)
	}
	t1, err := repo.AddTarget(ctx, entity.ClusterTarget{ClusterID: cid, URL: "http://up-a:80", Weight: 1, IsHealthy: true})
	if err != nil {
		t.Fatalf("add target a: %v", err)
	}
	if _, err := repo.AddTarget(ctx, entity.ClusterTarget{ClusterID: cid, URL: "http://up-b:80", Weight: 2, IsHealthy: true}); err != nil {
		t.Fatalf("add target b: %v", err)
	}

	got, err := repo.GetByID(ctx, cid)
	if err != nil {
		t.Fatalf("get cluster: %v", err)
	}
	if len(got.Targets) != 2 {
		t.Fatalf("expected 2 targets, got %d", len(got.Targets))
	}

	if err := repo.SetTargetHealth(ctx, t1, false); err != nil {
		t.Fatalf("set target health: %v", err)
	}
	if err := repo.DeleteTarget(ctx, t1); err != nil {
		t.Fatalf("delete target: %v", err)
	}
	if err := repo.Delete(ctx, cid); err != nil {
		t.Fatalf("delete cluster: %v", err)
	}
}

func TestMiddlewareRepoCRUDAndAttach(t *testing.T) {
	pool, cleanup := startPostgres(t)
	defer cleanup()

	mwRepo := NewMiddlewareRepo(pool)
	routeRepo := NewRouteRepo(pool)
	ctx := context.Background()

	rid, err := routeRepo.Create(ctx, entity.Route{
		Method: "GET", PathPattern: "/x", TargetURL: "http://up:80", IsActive: true,
	})
	if err != nil {
		t.Fatalf("create route: %v", err)
	}
	mwID, err := mwRepo.Create(ctx, entity.Middleware{
		Name: "cors-default", Kind: entity.MiddlewareKindCORS, Config: []byte(`{"origins":["*"]}`), IsActive: true,
	})
	if err != nil {
		t.Fatalf("create mw: %v", err)
	}
	if err := mwRepo.Attach(ctx, rid, mwID, 0); err != nil {
		t.Fatalf("attach: %v", err)
	}
	got, err := mwRepo.ListByRoute(ctx, rid)
	if err != nil {
		t.Fatalf("list by route: %v", err)
	}
	if len(got) != 1 || got[0].Kind != entity.MiddlewareKindCORS {
		t.Fatalf("unexpected list: %+v", got)
	}
	if err := mwRepo.Detach(ctx, rid, mwID); err != nil {
		t.Fatalf("detach: %v", err)
	}
	if err := mwRepo.Delete(ctx, mwID); err != nil {
		t.Fatalf("delete mw: %v", err)
	}
	if err := routeRepo.Delete(ctx, rid); err != nil {
		t.Fatalf("delete route: %v", err)
	}
}

func TestAPITokenRepoLifecycle(t *testing.T) {
	pool, cleanup := startPostgres(t)
	defer cleanup()

	adminRepo := NewAdminRepo(pool)
	tokRepo := NewAPITokenRepo(pool)
	ctx := context.Background()

	uid, err := adminRepo.CreateAdmin(ctx, entity.AdminUser{
		Username: "owner", PasswordHash: "h", Role: "admin", CreatedAt: time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("create admin: %v", err)
	}
	id, err := tokRepo.Create(ctx, entity.APIToken{
		UserID: uid, Name: "ci", Prefix: "abc", TokenHash: "h", Scopes: entity.ScopeRoutesRead,
	})
	if err != nil {
		t.Fatalf("create token: %v", err)
	}
	cands, err := tokRepo.GetByPrefix(ctx, "abc")
	if err != nil || len(cands) != 1 {
		t.Fatalf("by prefix: %v / %d", err, len(cands))
	}
	if err := tokRepo.TouchLastUsed(ctx, id); err != nil {
		t.Fatalf("touch: %v", err)
	}
	if err := tokRepo.Revoke(ctx, id); err != nil {
		t.Fatalf("revoke: %v", err)
	}
	got, err := tokRepo.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.RevokedAt == nil {
		t.Fatalf("expected revoked_at set")
	}
}
