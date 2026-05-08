//go:build integration

package repos

import (
	"context"
	"net/url"
	"testing"
	"time"

	"prodlich/internal/domain/entity"
	"prodlich/internal/domain/repository"
	dbpkg "prodlich/internal/infrastructure/db"
	"prodlich/internal/model"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func startPostgres(t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()
	ctx := context.Background()
	container, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("gateway_test"),
		postgres.WithUsername("gateway"),
		postgres.WithPassword("gateway-pass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		t.Skipf("postgres testcontainer unavailable: %v", err)
	}
	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("conn string: %v", err)
	}
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatalf("pool: %v", err)
	}

	mc := dsnToModelDB(t, dsn)
	if err := dbpkg.MigrateUp(&mc); err != nil {
		t.Fatalf("migrate up: %v", err)
	}
	cleanup := func() {
		pool.Close()
		_ = container.Terminate(ctx)
	}
	return pool, cleanup
}

func dsnToModelDB(t *testing.T, dsn string) model.DB {
	t.Helper()
	u, err := url.Parse(dsn)
	if err != nil {
		t.Fatalf("parse dsn: %v", err)
	}
	pwd, _ := u.User.Password()
	port := u.Port()
	if port == "" {
		port = "5432"
	}
	return model.DB{
		Host:     u.Hostname(),
		Port:     port,
		User:     u.User.Username(),
		Password: pwd,
		DBName:   u.Path[1:],
		SSLMode:  "disable",
	}
}

func TestRouteRepoCRUD(t *testing.T) {
	pool, cleanup := startPostgres(t)
	defer cleanup()

	repo := NewRouteRepo(pool)
	ctx := context.Background()
	id, err := repo.Create(ctx, entity.Route{
		Method: "GET", PathPattern: "/x", TargetURL: "http://upstream:8080", IsActive: true, Priority: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	got, err := repo.GetByID(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if got.PathPattern != "/x" {
		t.Errorf("path: %q", got.PathPattern)
	}

	got.RateLimit = 60
	if err := repo.Update(ctx, got); err != nil {
		t.Fatal(err)
	}
	if err := repo.ToggleActive(ctx, id); err != nil {
		t.Fatal(err)
	}
	all, err := repo.GetAll(ctx)
	if err != nil || len(all) != 1 {
		t.Fatalf("getall: len=%d err=%v", len(all), err)
	}
	active, err := repo.GetActiveRoutes(ctx)
	if err != nil || len(active) != 0 {
		t.Fatalf("active should be empty after toggle: %v / %d", err, len(active))
	}
	if err := repo.Delete(ctx, id); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.GetByID(ctx, id); err != repository.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestLogRepoBatchInsertAndQuery(t *testing.T) {
	pool, cleanup := startPostgres(t)
	defer cleanup()

	repo := NewLogRepo(pool)
	ctx := context.Background()
	logs := []entity.RequestLog{
		{
			RequestID: uuid.Must(uuid.NewV4()), Method: "GET", Path: "/a",
			ClientIP: "1.2.3.4", StatusCode: 200, ResponseTimeMs: 10, CreatedAt: time.Now().UTC(),
		},
		{
			RequestID: uuid.Must(uuid.NewV4()), Method: "POST", Path: "/b",
			ClientIP: "5.6.7.8", StatusCode: 500, ResponseTimeMs: 50, CreatedAt: time.Now().UTC(),
		},
	}
	if err := repo.InsertBatch(ctx, logs); err != nil {
		t.Fatal(err)
	}
	out, total, err := repo.GetWithFilters(ctx, repository.LogFilters{}, 100, 0)
	if err != nil {
		t.Fatal(err)
	}
	if total != 2 || len(out) != 2 {
		t.Fatalf("total=%d len=%d", total, len(out))
	}

	metrics, err := repo.GetMetrics(ctx, time.Now().Add(-time.Hour), time.Now().Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if metrics.TotalRequests != 2 {
		t.Errorf("total=%d", metrics.TotalRequests)
	}
}

func TestAdminRepoLifecycle(t *testing.T) {
	pool, cleanup := startPostgres(t)
	defer cleanup()

	repo := NewAdminRepo(pool)
	ctx := context.Background()
	id, err := repo.CreateAdmin(ctx, entity.AdminUser{
		Username: "alice", PasswordHash: "h", Role: "admin", CreatedAt: time.Now().UTC(),
	})
	if err != nil {
		t.Fatal(err)
	}
	got, err := repo.GetByUsername(ctx, "alice")
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != id {
		t.Errorf("id mismatch: %d vs %d", got.ID, id)
	}
	if _, err := repo.GetByUsername(ctx, "missing"); err != repository.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if err := repo.UpdateLastLogin(ctx, id); err != nil {
		t.Fatal(err)
	}
}

func TestUserRepoLifecycle(t *testing.T) {
	pool, cleanup := startPostgres(t)
	defer cleanup()

	repo := NewUserRepoPG(pool)
	ctx := context.Background()
	id, err := repo.Create(ctx, entity.User{
		Name: "Alice", Email: "alice@example.com", Password: "hash", CreatedAt: time.Now().UTC(),
	})
	if err != nil {
		t.Fatal(err)
	}
	got, err := repo.GetByEmail(ctx, "alice@example.com")
	if err != nil || got.ID != id {
		t.Fatalf("get by email: id=%d err=%v", got.ID, err)
	}
	got2, err := repo.GetByID(ctx, id)
	if err != nil || got2.ID != id {
		t.Fatalf("get by id: %v", err)
	}
	if _, err := repo.GetByEmail(ctx, "nobody"); err != repository.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
