package application

import (
	"context"
	"errors"
	"testing"

	"prodlich/internal/domain/entity"
	"prodlich/internal/domain/repository"
)

type fakeUserRepo struct {
	users  map[string]*entity.User
	byID   map[int64]*entity.User
	nextID int64
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{users: map[string]*entity.User{}, byID: map[int64]*entity.User{}}
}

func (f *fakeUserRepo) Create(_ context.Context, u entity.User) (int64, error) {
	f.nextID++
	u.ID = f.nextID
	clone := u
	f.users[u.Email] = &clone
	f.byID[u.ID] = &clone
	return u.ID, nil
}

func (f *fakeUserRepo) GetByEmail(_ context.Context, email string) (*entity.User, error) {
	if u, ok := f.users[email]; ok {
		return u, nil
	}
	return nil, repository.ErrNotFound
}

func (f *fakeUserRepo) GetByID(_ context.Context, id int64) (*entity.User, error) {
	if u, ok := f.byID[id]; ok {
		return u, nil
	}
	return nil, repository.ErrNotFound
}

func TestRegisterValidation(t *testing.T) {
	uc := NewAuthUseCase(newFakeUserRepo(), "this-is-a-very-secret-key-1234")
	tests := []struct {
		name, n, e, p string
		wantErr       error
	}{
		{name: "valid", n: "Alice", e: "alice@example.com", p: "secret-pass"},
		{name: "empty name", n: "", e: "alice@example.com", p: "secret-pass", wantErr: ErrInvalidName},
		{name: "bad email", n: "Alice", e: "not-an-email", p: "secret-pass", wantErr: ErrInvalidEmail},
		{name: "weak pwd", n: "Alice", e: "alice@example.com", p: "short", wantErr: ErrWeakPassword},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tok, err := uc.Register(context.Background(), tt.n, tt.e, tt.p)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("err=%v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if tok == "" {
				t.Fatal("empty token")
			}
		})
	}
}

func TestRegisterDuplicateEmail(t *testing.T) {
	repo := newFakeUserRepo()
	uc := NewAuthUseCase(repo, "this-is-a-very-secret-key-1234")
	if _, err := uc.Register(context.Background(), "A", "a@example.com", "abcdefgh"); err != nil {
		t.Fatalf("register first: %v", err)
	}
	if _, err := uc.Register(context.Background(), "A", "a@example.com", "abcdefgh"); !errors.Is(err, ErrEmailExists) {
		t.Fatalf("expected ErrEmailExists, got %v", err)
	}
}

func TestLoginAndValidate(t *testing.T) {
	repo := newFakeUserRepo()
	uc := NewAuthUseCase(repo, "this-is-a-very-secret-key-1234")
	if _, err := uc.Register(context.Background(), "Bob", "b@example.com", "abcdefgh"); err != nil {
		t.Fatalf("register: %v", err)
	}
	tok, err := uc.Login(context.Background(), "b@example.com", "abcdefgh")
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	user, err := uc.ValidateToken(context.Background(), tok)
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if user.Email != "b@example.com" {
		t.Errorf("user email %q", user.Email)
	}

	if _, err := uc.Login(context.Background(), "b@example.com", "wrong"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
	if _, err := uc.Login(context.Background(), "missing@example.com", "abcdefgh"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}
