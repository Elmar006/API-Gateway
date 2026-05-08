package application

import (
	"context"
	"errors"
	"testing"

	"prodlich/internal/domain/entity"
	"prodlich/internal/domain/repository"
)

type fakeEnvRepo struct {
	store map[int]entity.Environment
	next  int
}

func newFakeEnvRepo() *fakeEnvRepo { return &fakeEnvRepo{store: map[int]entity.Environment{}} }

func (f *fakeEnvRepo) Create(_ context.Context, e entity.Environment) (int, error) {
	f.next++
	e.ID = f.next
	f.store[e.ID] = e
	return e.ID, nil
}
func (f *fakeEnvRepo) GetByID(_ context.Context, id int) (entity.Environment, error) {
	if v, ok := f.store[id]; ok {
		return v, nil
	}
	return entity.Environment{}, repository.ErrNotFound
}
func (f *fakeEnvRepo) GetByName(_ context.Context, n string) (entity.Environment, error) {
	for _, v := range f.store {
		if v.Name == n {
			return v, nil
		}
	}
	return entity.Environment{}, repository.ErrNotFound
}
func (f *fakeEnvRepo) GetAll(_ context.Context) ([]entity.Environment, error) {
	out := make([]entity.Environment, 0, len(f.store))
	for _, v := range f.store {
		out = append(out, v)
	}
	return out, nil
}
func (f *fakeEnvRepo) Update(_ context.Context, e entity.Environment) error {
	if _, ok := f.store[e.ID]; !ok {
		return repository.ErrNotFound
	}
	f.store[e.ID] = e
	return nil
}
func (f *fakeEnvRepo) Delete(_ context.Context, id int) error {
	if _, ok := f.store[id]; !ok {
		return repository.ErrNotFound
	}
	delete(f.store, id)
	return nil
}

func TestEnvironmentValidate(t *testing.T) {
	uc := NewEnvironmentUseCase(newFakeEnvRepo())
	if _, err := uc.Create(context.Background(), entity.Environment{Name: ""}); !errors.Is(err, ErrInvalidEnvironment) {
		t.Fatalf("expected invalid for empty name, got %v", err)
	}
	if _, err := uc.Create(context.Background(), entity.Environment{Name: "with space"}); !errors.Is(err, ErrInvalidEnvironment) {
		t.Fatalf("expected invalid for whitespace, got %v", err)
	}
}

func TestEnvironmentCRUD(t *testing.T) {
	uc := NewEnvironmentUseCase(newFakeEnvRepo())
	id, err := uc.Create(context.Background(), entity.Environment{Name: "staging"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := uc.Update(context.Background(), entity.Environment{ID: id, Name: "stg"}); err != nil {
		t.Fatalf("update: %v", err)
	}
	got, err := uc.Get(context.Background(), id)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Name != "stg" {
		t.Fatalf("expected name=stg, got %q", got.Name)
	}
	if err := uc.Delete(context.Background(), id); err != nil {
		t.Fatalf("delete: %v", err)
	}
}
