package main

import (
	"context"
	"errors"
	"testing"

	"prodlich/internal/domain/entity"
	"prodlich/internal/domain/repository"
)

type stubAdminRepo struct {
	users   map[string]entity.AdminUser
	created entity.AdminUser
	getErr  error
}

func (s *stubAdminRepo) GetByUsername(_ context.Context, name string) (entity.AdminUser, error) {
	if s.getErr != nil {
		return entity.AdminUser{}, s.getErr
	}
	if u, ok := s.users[name]; ok {
		return u, nil
	}
	return entity.AdminUser{}, repository.ErrNotFound
}
func (s *stubAdminRepo) GetByID(context.Context, int) (entity.AdminUser, error) {
	return entity.AdminUser{}, repository.ErrNotFound
}
func (s *stubAdminRepo) CreateAdmin(_ context.Context, a entity.AdminUser) (int, error) {
	s.created = a
	if s.users == nil {
		s.users = map[string]entity.AdminUser{}
	}
	s.users[a.Username] = a
	return 1, nil
}
func (s *stubAdminRepo) UpdateLastLogin(context.Context, int) error { return nil }
func (s *stubAdminRepo) GetAll(context.Context) ([]entity.AdminUser, error) {
	return nil, nil
}
func (s *stubAdminRepo) UpdateAdmin(context.Context, entity.AdminUser) error { return nil }
func (s *stubAdminRepo) DeleteAdmin(context.Context, int) error              { return nil }

func TestBootstrapAdminEmptyArgsNoOp(t *testing.T) {
	repo := &stubAdminRepo{users: map[string]entity.AdminUser{}}
	if err := bootstrapAdmin(context.Background(), repo, "", ""); err != nil {
		t.Fatal(err)
	}
	if repo.created.Username != "" {
		t.Errorf("unexpected create: %+v", repo.created)
	}
}

func TestBootstrapAdminCreatesWhenMissing(t *testing.T) {
	repo := &stubAdminRepo{users: map[string]entity.AdminUser{}}
	if err := bootstrapAdmin(context.Background(), repo, "admin", "supersecret"); err != nil {
		t.Fatal(err)
	}
	if repo.created.Username != "admin" || repo.created.PasswordHash == "" || repo.created.Role != "admin" {
		t.Fatalf("missing or invalid bootstrap admin: %+v", repo.created)
	}
}

func TestBootstrapAdminSkipsWhenExists(t *testing.T) {
	repo := &stubAdminRepo{users: map[string]entity.AdminUser{"admin": {ID: 1, Username: "admin"}}}
	if err := bootstrapAdmin(context.Background(), repo, "admin", "supersecret"); err != nil {
		t.Fatal(err)
	}
	if repo.created.Username != "" {
		t.Errorf("should not have created: %+v", repo.created)
	}
}

func TestBootstrapAdminPropagatesUnexpectedError(t *testing.T) {
	want := errors.New("db down")
	repo := &stubAdminRepo{getErr: want}
	if err := bootstrapAdmin(context.Background(), repo, "admin", "supersecret"); !errors.Is(err, want) {
		t.Fatalf("expected wrapped %v, got %v", want, err)
	}
}

func TestNewServerHasTimeouts(t *testing.T) {
	srv := newServer(":0", nil)
	if srv.ReadTimeout == 0 || srv.WriteTimeout == 0 || srv.IdleTimeout == 0 || srv.ReadHeaderTimeout == 0 {
		t.Fatalf("missing timeouts: %+v", srv)
	}
}
