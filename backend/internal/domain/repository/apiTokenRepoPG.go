package repository

import (
	"context"

	"prodlich/internal/domain/entity"
)

// APITokenRepo is the persistence contract for programmatic API tokens.
type APITokenRepo interface {
	Create(ctx context.Context, t entity.APIToken) (int, error)
	GetByID(ctx context.Context, id int) (entity.APIToken, error)
	// GetByPrefix returns all candidate tokens whose prefix matches; the
	// caller bcrypt-compares the supplied raw secret against each row's
	// token_hash. Prefix is short and may collide, so this can return
	// multiple rows; in practice it almost never does (16-char base32 prefix).
	GetByPrefix(ctx context.Context, prefix string) ([]entity.APIToken, error)
	ListByUser(ctx context.Context, userID int) ([]entity.APIToken, error)
	Revoke(ctx context.Context, id int) error
	TouchLastUsed(ctx context.Context, id int) error
}
