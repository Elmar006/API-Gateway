package application

import (
	"context"
	"errors"
	"testing"

	"prodlich/internal/domain/entity"
	"prodlich/internal/domain/repository"
	authpkg "prodlich/internal/infrastructure/auth"

	"golang.org/x/crypto/bcrypt"
)

type fakeAdminRepo struct {
	byName map[string]entity.AdminUser
	byID   map[int]entity.AdminUser
	last   chan int
}

func (f *fakeAdminRepo) GetByUsername(_ context.Context, u string) (entity.AdminUser, error) {
	if v, ok := f.byName[u]; ok {
		return v, nil
	}
	return entity.AdminUser{}, repository.ErrNotFound
}
func (f *fakeAdminRepo) GetByID(_ context.Context, id int) (entity.AdminUser, error) {
	if v, ok := f.byID[id]; ok {
		return v, nil
	}
	return entity.AdminUser{}, repository.ErrNotFound
}
func (f *fakeAdminRepo) CreateAdmin(_ context.Context, a entity.AdminUser) (int, error) {
	a.ID = len(f.byID) + 1
	if f.byName == nil {
		f.byName = map[string]entity.AdminUser{}
	}
	if f.byID == nil {
		f.byID = map[int]entity.AdminUser{}
	}
	f.byName[a.Username] = a
	f.byID[a.ID] = a
	return a.ID, nil
}
func (f *fakeAdminRepo) UpdateLastLogin(_ context.Context, id int) error {
	if f.last != nil {
		f.last <- id
	}
	return nil
}
func (f *fakeAdminRepo) GetAll(_ context.Context) ([]entity.AdminUser, error) {
	out := make([]entity.AdminUser, 0, len(f.byID))
	for _, v := range f.byID {
		out = append(out, v)
	}
	return out, nil
}
func (f *fakeAdminRepo) UpdateAdmin(_ context.Context, a entity.AdminUser) error {
	if _, ok := f.byID[a.ID]; !ok {
		return repository.ErrNotFound
	}
	f.byID[a.ID] = a
	f.byName[a.Username] = a
	return nil
}
func (f *fakeAdminRepo) DeleteAdmin(_ context.Context, id int) error {
	if _, ok := f.byID[id]; !ok {
		return repository.ErrNotFound
	}
	delete(f.byID, id)
	return nil
}

func TestAdminAuthLoginValidate(t *testing.T) {
	repo := &fakeAdminRepo{byName: map[string]entity.AdminUser{}, byID: map[int]entity.AdminUser{}, last: make(chan int, 1)}
	hash, _ := bcrypt.GenerateFromPassword([]byte("admin-pass"), bcrypt.MinCost)
	repo.byName["admin"] = entity.AdminUser{ID: 42, Username: "admin", PasswordHash: string(hash), Role: "admin"}
	repo.byID[42] = repo.byName["admin"]

	tokSvc := authpkg.NewJWTService("super-test-secret-1234567890")
	uc := NewAdminAuthUseCase(repo, tokSvc)

	tok, err := uc.Login(context.Background(), "admin", "admin-pass")
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if tok == "" {
		t.Fatal("empty token")
	}

	a, err := uc.ValidateToken(context.Background(), tok)
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if a == nil || a.Username != "admin" {
		t.Fatalf("admin mismatch: %+v", a)
	}
}

func TestAdminAuthLoginWrongPassword(t *testing.T) {
	repo := &fakeAdminRepo{byName: map[string]entity.AdminUser{}, byID: map[int]entity.AdminUser{}}
	hash, _ := bcrypt.GenerateFromPassword([]byte("admin-pass"), bcrypt.MinCost)
	repo.byName["admin"] = entity.AdminUser{ID: 1, Username: "admin", PasswordHash: string(hash)}
	repo.byID[1] = repo.byName["admin"]
	uc := NewAdminAuthUseCase(repo, authpkg.NewJWTService("test-secret-test-secret-1234"))

	if _, err := uc.Login(context.Background(), "admin", "wrong"); err == nil || !errors.Is(err, errors.New("invalid credentials")) && err.Error() != "invalid credentials" {
		t.Fatalf("expected invalid credentials, got %v", err)
	}
	if _, err := uc.Login(context.Background(), "missing", "x"); err == nil || err.Error() != "invalid credentials" {
		t.Fatalf("expected invalid credentials for missing user, got %v", err)
	}
}

func TestAdminAuthValidateGarbageToken(t *testing.T) {
	repo := &fakeAdminRepo{byName: map[string]entity.AdminUser{}, byID: map[int]entity.AdminUser{}}
	uc := NewAdminAuthUseCase(repo, authpkg.NewJWTService("test-secret-test-secret-1234"))
	if _, err := uc.ValidateToken(context.Background(), "garbage"); err == nil {
		t.Fatal("expected error")
	}
}
