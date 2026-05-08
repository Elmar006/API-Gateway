package application

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"prodlich/internal/domain/entity"
	"prodlich/internal/domain/repository"
)

type fakeTokenRepo struct {
	store  map[int]entity.APIToken
	nextID int
}

func newFakeTokenRepo() *fakeTokenRepo {
	return &fakeTokenRepo{store: map[int]entity.APIToken{}}
}

func (f *fakeTokenRepo) Create(_ context.Context, t entity.APIToken) (int, error) {
	f.nextID++
	t.ID = f.nextID
	t.CreatedAt = time.Now().UTC()
	f.store[t.ID] = t
	return t.ID, nil
}
func (f *fakeTokenRepo) GetByID(_ context.Context, id int) (entity.APIToken, error) {
	if v, ok := f.store[id]; ok {
		return v, nil
	}
	return entity.APIToken{}, repository.ErrNotFound
}
func (f *fakeTokenRepo) GetByPrefix(_ context.Context, prefix string) ([]entity.APIToken, error) {
	out := []entity.APIToken{}
	for _, v := range f.store {
		if v.Prefix == prefix {
			out = append(out, v)
		}
	}
	return out, nil
}
func (f *fakeTokenRepo) ListByUser(_ context.Context, userID int) ([]entity.APIToken, error) {
	out := []entity.APIToken{}
	for _, v := range f.store {
		if v.UserID == userID {
			out = append(out, v)
		}
	}
	return out, nil
}
func (f *fakeTokenRepo) Revoke(_ context.Context, id int) error {
	v, ok := f.store[id]
	if !ok {
		return repository.ErrNotFound
	}
	now := time.Now().UTC()
	v.RevokedAt = &now
	f.store[id] = v
	return nil
}
func (f *fakeTokenRepo) TouchLastUsed(_ context.Context, id int) error {
	v, ok := f.store[id]
	if !ok {
		return repository.ErrNotFound
	}
	now := time.Now().UTC()
	v.LastUsedAt = &now
	f.store[id] = v
	return nil
}

func TestAPITokenCreateAuthenticate(t *testing.T) {
	uc := NewAPITokenUseCase(newFakeTokenRepo())
	created, err := uc.Create(context.Background(), 42, "ci-bot", []string{entity.ScopeRoutesRead, entity.ScopeMetricsRead}, nil)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if !strings.HasPrefix(created.Secret, created.Token.Prefix+".") {
		t.Fatalf("expected secret to start with prefix; got %q (prefix %q)", created.Secret, created.Token.Prefix)
	}
	resolved, err := uc.Authenticate(context.Background(), created.Secret)
	if err != nil {
		t.Fatalf("authenticate: %v", err)
	}
	if resolved.UserID != 42 {
		t.Fatalf("expected user 42, got %d", resolved.UserID)
	}
	if !resolved.HasScope(entity.ScopeRoutesRead) {
		t.Fatalf("expected scope routes:read, got %q", resolved.Scopes)
	}
	if resolved.HasScope(entity.ScopeRoutesWrite) {
		t.Fatalf("did not expect scope routes:write")
	}
}

func TestAPITokenInvalid(t *testing.T) {
	uc := NewAPITokenUseCase(newFakeTokenRepo())
	if _, err := uc.Create(context.Background(), 1, "x", []string{"unknown:scope"}, nil); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected invalid for bad scope, got %v", err)
	}
	if _, err := uc.Authenticate(context.Background(), "no-dot-secret"); !errors.Is(err, ErrTokenNotFound) {
		t.Fatalf("expected not-found for malformed secret, got %v", err)
	}
}

func TestAPITokenRevoke(t *testing.T) {
	uc := NewAPITokenUseCase(newFakeTokenRepo())
	created, err := uc.Create(context.Background(), 1, "x", []string{entity.ScopeWildcard}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := uc.Revoke(context.Background(), created.Token.ID); err != nil {
		t.Fatalf("revoke: %v", err)
	}
	if _, err := uc.Authenticate(context.Background(), created.Secret); !errors.Is(err, ErrTokenNotFound) {
		t.Fatalf("expected revoked token to fail authentication, got %v", err)
	}
}
