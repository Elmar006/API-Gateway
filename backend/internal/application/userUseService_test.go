package application

import (
	"context"
	"errors"
	"testing"

	"prodlich/internal/domain/entity"
)

func TestUserValidate(t *testing.T) {
	repo := &fakeAdminRepo{byName: map[string]entity.AdminUser{}, byID: map[int]entity.AdminUser{}}
	uc := NewUserUseCase(repo)
	if _, err := uc.Create(context.Background(), entity.AdminUser{Username: "", Role: entity.RoleAdmin}, "supersecret"); !errors.Is(err, ErrInvalidUser) {
		t.Fatalf("expected invalid for empty username")
	}
	if _, err := uc.Create(context.Background(), entity.AdminUser{Username: "u", Role: "bogus"}, "supersecret"); !errors.Is(err, ErrInvalidUser) {
		t.Fatalf("expected invalid for bad role")
	}
	if _, err := uc.Create(context.Background(), entity.AdminUser{Username: "u", Role: entity.RoleViewer}, "short"); !errors.Is(err, ErrInvalidUser) {
		t.Fatalf("expected invalid for short password")
	}
}

func TestUserCreateUpdateDelete(t *testing.T) {
	repo := &fakeAdminRepo{byName: map[string]entity.AdminUser{}, byID: map[int]entity.AdminUser{}}
	uc := NewUserUseCase(repo)
	id, err := uc.Create(context.Background(), entity.AdminUser{Username: "alice", Role: entity.RoleEditor}, "supersecret")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := uc.Update(context.Background(), entity.AdminUser{ID: id, Username: "alice", Role: entity.RoleAdmin}, ""); err != nil {
		t.Fatalf("update: %v", err)
	}
	if got, err := uc.Get(context.Background(), id); err != nil || got.Role != entity.RoleAdmin {
		t.Fatalf("get: %v %+v", err, got)
	}
	if err := uc.Delete(context.Background(), id); err != nil {
		t.Fatalf("delete: %v", err)
	}
}
