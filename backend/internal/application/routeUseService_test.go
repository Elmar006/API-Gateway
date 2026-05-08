package application

import (
	"context"
	"errors"
	"testing"

	"prodlich/internal/domain/entity"
	"prodlich/internal/domain/repository"
)

type fakeRouteRepo struct {
	store  map[int]entity.Route
	nextID int
}

func newFakeRouteRepo() *fakeRouteRepo {
	return &fakeRouteRepo{store: map[int]entity.Route{}}
}

func (f *fakeRouteRepo) Create(_ context.Context, r entity.Route) (int, error) {
	f.nextID++
	r.ID = f.nextID
	f.store[r.ID] = r
	return r.ID, nil
}

func (f *fakeRouteRepo) GetByID(_ context.Context, id int) (entity.Route, error) {
	if r, ok := f.store[id]; ok {
		return r, nil
	}
	return entity.Route{}, repository.ErrNotFound
}

func (f *fakeRouteRepo) GetAll(_ context.Context) ([]entity.Route, error) {
	out := make([]entity.Route, 0, len(f.store))
	for _, r := range f.store {
		out = append(out, r)
	}
	return out, nil
}

func (f *fakeRouteRepo) Update(_ context.Context, r entity.Route) error {
	if _, ok := f.store[r.ID]; !ok {
		return repository.ErrNotFound
	}
	f.store[r.ID] = r
	return nil
}

func (f *fakeRouteRepo) Delete(_ context.Context, id int) error {
	if _, ok := f.store[id]; !ok {
		return repository.ErrNotFound
	}
	delete(f.store, id)
	return nil
}

func (f *fakeRouteRepo) ToggleActive(_ context.Context, id int) error {
	r, ok := f.store[id]
	if !ok {
		return repository.ErrNotFound
	}
	r.IsActive = !r.IsActive
	f.store[id] = r
	return nil
}

func (f *fakeRouteRepo) GetActiveRoutes(_ context.Context) ([]entity.Route, error) {
	out := []entity.Route{}
	for _, r := range f.store {
		if r.IsActive {
			out = append(out, r)
		}
	}
	return out, nil
}

func TestValidateRoute(t *testing.T) {
	tests := []struct {
		name string
		r    entity.Route
		ok   bool
	}{
		{name: "valid", r: entity.Route{PathPattern: "/x", TargetURL: "http://upstream:8080"}, ok: true},
		{name: "no path", r: entity.Route{TargetURL: "http://x"}},
		{name: "missing slash", r: entity.Route{PathPattern: "x", TargetURL: "http://x"}},
		{name: "bad target", r: entity.Route{PathPattern: "/x", TargetURL: "not-a-url"}},
		{name: "negative limit", r: entity.Route{PathPattern: "/x", TargetURL: "http://x", RateLimit: -1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRoute(tt.r)
			if tt.ok && err != nil {
				t.Fatalf("ok=true but err=%v", err)
			}
			if !tt.ok && err == nil {
				t.Fatal("expected error")
			}
			if !tt.ok && !errors.Is(err, ErrInvalidRoute) {
				t.Fatalf("error not wrapped: %v", err)
			}
		})
	}
}

func TestRouteCreateReloadsAndNotifies(t *testing.T) {
	repo := newFakeRouteRepo()
	uc := NewRouterService(repo, true)
	var observed int
	uc.OnReload(func(routes []entity.Route) {
		observed = len(routes)
	})
	id, err := uc.Create(context.Background(), entity.Route{
		Method: "get", PathPattern: "/x", TargetURL: "http://up:8080", IsActive: true,
	})
	if err != nil || id == 0 {
		t.Fatalf("create: id=%d err=%v", id, err)
	}
	if observed != 1 {
		t.Fatalf("observer not notified, got %d", observed)
	}
	if got := uc.GetActiveRoutes(); len(got) != 1 {
		t.Fatalf("active routes = %d", len(got))
	}
	if got := uc.GetActiveRoutes()[0].Method; got != "GET" {
		t.Fatalf("method not normalized: %q", got)
	}
}

func TestRouteUpdateValidationFails(t *testing.T) {
	repo := newFakeRouteRepo()
	uc := NewRouterService(repo, true)
	if _, err := uc.Create(context.Background(), entity.Route{Method: "GET", PathPattern: "/x", TargetURL: "http://up", IsActive: true}); err != nil {
		t.Fatal(err)
	}
	err := uc.UpdateRoute(context.Background(), entity.Route{ID: 1, PathPattern: "", TargetURL: "http://up"})
	if !errors.Is(err, ErrInvalidRoute) {
		t.Fatalf("expected ErrInvalidRoute, got %v", err)
	}
}

func TestRouteUpdateAndGet(t *testing.T) {
	repo := newFakeRouteRepo()
	uc := NewRouterService(repo, true)
	id, err := uc.Create(context.Background(), entity.Route{Method: "GET", PathPattern: "/x", TargetURL: "http://up", IsActive: true})
	if err != nil {
		t.Fatal(err)
	}
	if err := uc.UpdateRoute(context.Background(), entity.Route{ID: id, Method: "POST", PathPattern: "/y", TargetURL: "http://up", IsActive: true}); err != nil {
		t.Fatalf("update: %v", err)
	}
	r, err := uc.GetRouteByID(context.Background(), id)
	if err != nil || r.PathPattern != "/y" || r.Method != "POST" {
		t.Fatalf("get by id: %+v err=%v", r, err)
	}
	all, err := uc.GetAllRoutes(context.Background())
	if err != nil || len(all) != 1 {
		t.Fatalf("get all: len=%d err=%v", len(all), err)
	}
}

func TestRouteUpdateMissing(t *testing.T) {
	uc := NewRouterService(newFakeRouteRepo(), true)
	err := uc.UpdateRoute(context.Background(), entity.Route{ID: 99, Method: "GET", PathPattern: "/x", TargetURL: "http://up", IsActive: true})
	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestRouteToggleMissing(t *testing.T) {
	uc := NewRouterService(newFakeRouteRepo(), true)
	if err := uc.ToggleRoute(context.Background(), 99); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestRouteDeleteAndToggle(t *testing.T) {
	repo := newFakeRouteRepo()
	uc := NewRouterService(repo, true)
	id, err := uc.Create(context.Background(), entity.Route{Method: "GET", PathPattern: "/x", TargetURL: "http://up", IsActive: true})
	if err != nil {
		t.Fatal(err)
	}
	if err := uc.ToggleRoute(context.Background(), id); err != nil {
		t.Fatal(err)
	}
	if got := len(uc.GetActiveRoutes()); got != 0 {
		t.Fatalf("expected 0 active routes after toggle, got %d", got)
	}
	if err := uc.DeleteRoute(context.Background(), id); err != nil {
		t.Fatal(err)
	}
	if err := uc.DeleteRoute(context.Background(), id); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
