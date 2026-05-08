package application

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"errors"
	"fmt"
	"strings"
	"time"

	"prodlich/internal/domain/entity"
	repo "prodlich/internal/domain/repository"

	"golang.org/x/crypto/bcrypt"
)

// ErrInvalidToken is returned for token validation failures.
var ErrInvalidToken = errors.New("invalid api token")

// ErrTokenNotFound is returned when no active token matches the supplied secret.
var ErrTokenNotFound = errors.New("api token not found")

// validScopes is the closed set of permissible scopes; used to validate input
// at creation time.
var validScopes = func() map[string]struct{} {
	out := map[string]struct{}{entity.ScopeWildcard: {}}
	for _, s := range entity.AllScopes {
		out[s] = struct{}{}
	}
	return out
}()

// CreatedToken bundles the persisted row with the one-time-visible plaintext
// secret. The plaintext is never persisted.
type CreatedToken struct {
	Token  entity.APIToken
	Secret string
}

// APIToken is the use case for managing programmatic API tokens.
type APIToken struct {
	repo repo.APITokenRepo
}

// NewAPITokenUseCase wires an APIToken use case.
func NewAPITokenUseCase(r repo.APITokenRepo) *APIToken {
	return &APIToken{repo: r}
}

// List returns every token belonging to a user.
func (a *APIToken) List(ctx context.Context, userID int) ([]entity.APIToken, error) {
	return a.repo.ListByUser(ctx, userID)
}

// Create generates a new secret, validates scopes, and persists the row.
//
// The returned CreatedToken.Secret is the only time the plaintext value is
// available; callers must surface it to the user immediately.
func (a *APIToken) Create(ctx context.Context, userID int, name string, scopes []string, expiresAt *time.Time) (CreatedToken, error) {
	if userID <= 0 {
		return CreatedToken{}, fmt.Errorf("%w: user_id is required", ErrInvalidToken)
	}
	if strings.TrimSpace(name) == "" {
		return CreatedToken{}, fmt.Errorf("%w: name is required", ErrInvalidToken)
	}
	for _, s := range scopes {
		if _, ok := validScopes[s]; !ok {
			return CreatedToken{}, fmt.Errorf("%w: unknown scope %q", ErrInvalidToken, s)
		}
	}
	prefix, err := randomBase32(10)
	if err != nil {
		return CreatedToken{}, fmt.Errorf("generate prefix: %w", err)
	}
	secretPart, err := randomBase32(32)
	if err != nil {
		return CreatedToken{}, fmt.Errorf("generate secret: %w", err)
	}
	plaintext := prefix + "." + secretPart
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), bcrypt.DefaultCost)
	if err != nil {
		return CreatedToken{}, fmt.Errorf("hash token: %w", err)
	}
	row := entity.APIToken{
		UserID:    userID,
		Name:      name,
		Prefix:    prefix,
		TokenHash: string(hash),
		Scopes:    strings.Join(scopes, ","),
		ExpiresAt: expiresAt,
	}
	id, err := a.repo.Create(ctx, row)
	if err != nil {
		return CreatedToken{}, err
	}
	row.ID = id
	row.CreatedAt = time.Now().UTC()
	return CreatedToken{Token: row, Secret: plaintext}, nil
}

// Revoke marks a token as revoked (idempotent).
func (a *APIToken) Revoke(ctx context.Context, id int) error {
	return a.repo.Revoke(ctx, id)
}

// Authenticate looks up an active token whose plaintext secret matches the
// supplied value, then best-effort updates last_used_at and returns the row.
//
// Returns ErrTokenNotFound when no candidate matches.
func (a *APIToken) Authenticate(ctx context.Context, secret string) (entity.APIToken, error) {
	prefix, _, ok := strings.Cut(secret, ".")
	if !ok || len(prefix) == 0 {
		return entity.APIToken{}, ErrTokenNotFound
	}
	candidates, err := a.repo.GetByPrefix(ctx, prefix)
	if err != nil {
		return entity.APIToken{}, err
	}
	now := time.Now().UTC()
	for _, c := range candidates {
		if !c.IsActive(now) {
			continue
		}
		if bcrypt.CompareHashAndPassword([]byte(c.TokenHash), []byte(secret)) == nil {
			// Best-effort update; ignore errors so a transient DB blip
			// doesn't reject an otherwise valid token.
			_ = a.repo.TouchLastUsed(ctx, c.ID)
			return c, nil
		}
	}
	return entity.APIToken{}, ErrTokenNotFound
}

// randomBase32 returns a lowercased un-padded base32 string with `nBytes` of
// entropy. Useful for short, copy-pasteable secrets that survive newlines.
func randomBase32(nBytes int) (string, error) {
	buf := make([]byte, nBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return strings.ToLower(strings.TrimRight(base32.StdEncoding.EncodeToString(buf), "=")), nil
}
