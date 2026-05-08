package repos

import (
	"context"
	"errors"

	"prodlich/internal/domain/entity"
	"prodlich/internal/domain/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// APITokenRepo is the PostgreSQL implementation of repository.APITokenRepo.
type APITokenRepo struct {
	pool *pgxpool.Pool
}

// NewAPITokenRepo constructs an APITokenRepo against the given pool.
func NewAPITokenRepo(pool *pgxpool.Pool) *APITokenRepo {
	return &APITokenRepo{pool: pool}
}

const apiTokenColumns = `id, user_id, name, prefix, token_hash, scopes, last_used_at, expires_at, revoked_at, created_at`

func scanToken(row pgx.Row) (entity.APIToken, error) {
	var t entity.APIToken
	if err := row.Scan(
		&t.ID, &t.UserID, &t.Name, &t.Prefix, &t.TokenHash, &t.Scopes,
		&t.LastUsedAt, &t.ExpiresAt, &t.RevokedAt, &t.CreatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.APIToken{}, repository.ErrNotFound
		}
		return entity.APIToken{}, err
	}
	return t, nil
}

// Create inserts a new token row and returns its id.
func (r *APITokenRepo) Create(ctx context.Context, t entity.APIToken) (int, error) {
	const query = `INSERT INTO api_tokens
                       (user_id, name, prefix, token_hash, scopes, expires_at)
                   VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	var id int
	if err := r.pool.QueryRow(ctx, query,
		t.UserID, t.Name, t.Prefix, t.TokenHash, t.Scopes, t.ExpiresAt,
	).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

// GetByID returns a token by id.
func (r *APITokenRepo) GetByID(ctx context.Context, id int) (entity.APIToken, error) {
	return scanToken(r.pool.QueryRow(ctx, `SELECT `+apiTokenColumns+` FROM api_tokens WHERE id = $1`, id))
}

// GetByPrefix returns all candidate tokens with a matching prefix.
func (r *APITokenRepo) GetByPrefix(ctx context.Context, prefix string) ([]entity.APIToken, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+apiTokenColumns+` FROM api_tokens WHERE prefix = $1`, prefix)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []entity.APIToken{}
	for rows.Next() {
		t, err := scanToken(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// ListByUser returns every token belonging to a user (including revoked ones,
// for audit purposes).
func (r *APITokenRepo) ListByUser(ctx context.Context, userID int) ([]entity.APIToken, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+apiTokenColumns+` FROM api_tokens WHERE user_id = $1 ORDER BY id DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []entity.APIToken{}
	for rows.Next() {
		t, err := scanToken(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// Revoke marks a token as revoked (idempotent).
func (r *APITokenRepo) Revoke(ctx context.Context, id int) error {
	const query = `UPDATE api_tokens SET revoked_at = COALESCE(revoked_at, NOW()) WHERE id = $1`
	res, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return repository.ErrNotFound
	}
	return nil
}

// TouchLastUsed updates last_used_at to NOW(); errors are tolerable in the
// auth path (best-effort write).
func (r *APITokenRepo) TouchLastUsed(ctx context.Context, id int) error {
	_, err := r.pool.Exec(ctx, `UPDATE api_tokens SET last_used_at = NOW() WHERE id = $1`, id)
	return err
}
