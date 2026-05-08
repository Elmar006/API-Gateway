package application

import (
	"context"
	"errors"
	"testing"

	"prodlich/internal/domain/entity"
	"prodlich/internal/domain/repository"
)

type fakeMWRepo struct {
	mws    map[int]entity.Middleware
	links  map[int][]int // routeID -> middleware IDs in sort order
	nextID int
}

func newFakeMWRepo() *fakeMWRepo {
	return &fakeMWRepo{mws: map[int]entity.Middleware{}, links: map[int][]int{}}
}

func (f *fakeMWRepo) Create(_ context.Context, m entity.Middleware) (int, error) {
	f.nextID++
	m.ID = f.nextID
	f.mws[m.ID] = m
	return m.ID, nil
}
func (f *fakeMWRepo) GetByID(_ context.Context, id int) (entity.Middleware, error) {
	if v, ok := f.mws[id]; ok {
		return v, nil
	}
	return entity.Middleware{}, repository.ErrNotFound
}
func (f *fakeMWRepo) GetAll(_ context.Context) ([]entity.Middleware, error) {
	out := make([]entity.Middleware, 0, len(f.mws))
	for _, v := range f.mws {
		out = append(out, v)
	}
	return out, nil
}
func (f *fakeMWRepo) Update(_ context.Context, m entity.Middleware) error {
	if _, ok := f.mws[m.ID]; !ok {
		return repository.ErrNotFound
	}
	f.mws[m.ID] = m
	return nil
}
func (f *fakeMWRepo) Delete(_ context.Context, id int) error {
	if _, ok := f.mws[id]; !ok {
		return repository.ErrNotFound
	}
	delete(f.mws, id)
	return nil
}
func (f *fakeMWRepo) Attach(_ context.Context, routeID, mwID, _ int) error {
	f.links[routeID] = append(f.links[routeID], mwID)
	return nil
}
func (f *fakeMWRepo) Detach(_ context.Context, routeID, mwID int) error {
	rs := f.links[routeID]
	for i, id := range rs {
		if id == mwID {
			f.links[routeID] = append(rs[:i], rs[i+1:]...)
			return nil
		}
	}
	return repository.ErrNotFound
}
func (f *fakeMWRepo) ListByRoute(_ context.Context, routeID int) ([]entity.Middleware, error) {
	out := []entity.Middleware{}
	for _, id := range f.links[routeID] {
		if v, ok := f.mws[id]; ok {
			out = append(out, v)
		}
	}
	return out, nil
}

func TestMiddlewareValidate(t *testing.T) {
	uc := NewMiddlewareUseCase(newFakeMWRepo())
	if _, err := uc.Create(context.Background(), entity.Middleware{Name: "", Kind: entity.MiddlewareKindCORS}); !errors.Is(err, ErrInvalidMiddleware) {
		t.Fatalf("expected invalid for empty name")
	}
	if _, err := uc.Create(context.Background(), entity.Middleware{Name: "x", Kind: "weirdo"}); !errors.Is(err, ErrInvalidMiddleware) {
		t.Fatalf("expected invalid for unknown kind")
	}
	if _, err := uc.Create(context.Background(), entity.Middleware{Name: "x", Kind: entity.MiddlewareKindCORS, Config: []byte("not json")}); !errors.Is(err, ErrInvalidMiddleware) {
		t.Fatalf("expected invalid for bad config json")
	}
}

func TestMiddlewareAttachDetachOrdering(t *testing.T) {
	uc := NewMiddlewareUseCase(newFakeMWRepo())
	a, err := uc.Create(context.Background(), entity.Middleware{Name: "cors", Kind: entity.MiddlewareKindCORS})
	if err != nil {
		t.Fatal(err)
	}
	b, err := uc.Create(context.Background(), entity.Middleware{Name: "rid", Kind: entity.MiddlewareKindRequestID})
	if err != nil {
		t.Fatal(err)
	}
	if err := uc.Attach(context.Background(), 7, a, 0); err != nil {
		t.Fatalf("attach a: %v", err)
	}
	if err := uc.Attach(context.Background(), 7, b, 1); err != nil {
		t.Fatalf("attach b: %v", err)
	}
	got, err := uc.ListByRoute(context.Background(), 7)
	if err != nil || len(got) != 2 {
		t.Fatalf("list: got %d items err=%v", len(got), err)
	}
	if err := uc.Detach(context.Background(), 7, a); err != nil {
		t.Fatalf("detach: %v", err)
	}
}
