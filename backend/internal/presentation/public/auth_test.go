package public

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"prodlich/internal/application"
	"prodlich/internal/domain/entity"
	"prodlich/internal/domain/repository"
)

type stubUserRepo struct {
	users  map[string]*entity.User
	byID   map[int64]*entity.User
	nextID int64
}

func newStubRepo() *stubUserRepo {
	return &stubUserRepo{users: map[string]*entity.User{}, byID: map[int64]*entity.User{}}
}

func (s *stubUserRepo) Create(_ context.Context, u entity.User) (int64, error) {
	s.nextID++
	u.ID = s.nextID
	clone := u
	s.users[u.Email] = &clone
	s.byID[u.ID] = &clone
	return u.ID, nil
}

func (s *stubUserRepo) GetByEmail(_ context.Context, email string) (*entity.User, error) {
	if u, ok := s.users[email]; ok {
		return u, nil
	}
	return nil, repository.ErrNotFound
}

func (s *stubUserRepo) GetByID(_ context.Context, id int64) (*entity.User, error) {
	if u, ok := s.byID[id]; ok {
		return u, nil
	}
	return nil, repository.ErrNotFound
}

func TestUserRegisterHappyPath(t *testing.T) {
	uc := application.NewAuthUseCase(newStubRepo(), "this-is-a-secure-test-secret-12")
	body := strings.NewReader(`{"name":"Alice","email":"alice@example.com","password":"abcdefgh"}`)
	rec := httptest.NewRecorder()
	UserRegisterHandler(uc).ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/auth/register", body))
	if rec.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"token":`)) {
		t.Errorf("missing token in response: %s", rec.Body.String())
	}
}

func TestUserRegisterValidationError(t *testing.T) {
	uc := application.NewAuthUseCase(newStubRepo(), "this-is-a-secure-test-secret-12")
	body := strings.NewReader(`{"name":"","email":"alice@example.com","password":"abcdefgh"}`)
	rec := httptest.NewRecorder()
	UserRegisterHandler(uc).ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/auth/register", body))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d", rec.Code)
	}
}

func TestUserLogin(t *testing.T) {
	repo := newStubRepo()
	uc := application.NewAuthUseCase(repo, "this-is-a-secure-test-secret-12")
	if _, err := uc.Register(context.Background(), "Alice", "a@example.com", "abcdefgh"); err != nil {
		t.Fatal(err)
	}
	body := strings.NewReader(`{"email":"a@example.com","password":"abcdefgh"}`)
	rec := httptest.NewRecorder()
	UserLoginHandler(uc).ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/auth/login", body))
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestUserLoginInvalid(t *testing.T) {
	uc := application.NewAuthUseCase(newStubRepo(), "this-is-a-secure-test-secret-12")
	body := strings.NewReader(`{"email":"none@example.com","password":"abcdefgh"}`)
	rec := httptest.NewRecorder()
	UserLoginHandler(uc).ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/auth/login", body))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d", rec.Code)
	}
}
