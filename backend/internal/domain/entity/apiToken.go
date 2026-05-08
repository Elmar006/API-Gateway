package entity

import (
	"strings"
	"time"
)

// Role constants used by admin_users and enforced by RBAC middleware.
const (
	RoleAdmin  = "admin"
	RoleEditor = "editor"
	RoleViewer = "viewer"
)

// Scope constants — a closed set of permissions that can be embedded into an
// API token. Wildcard "*" grants all scopes.
const (
	ScopeWildcard           = "*"
	ScopeRoutesRead         = "routes:read"
	ScopeRoutesWrite        = "routes:write"
	ScopeClustersRead       = "clusters:read"
	ScopeClustersWrite      = "clusters:write"
	ScopeMiddlewaresRead    = "middlewares:read"
	ScopeMiddlewaresWrite   = "middlewares:write"
	ScopeEnvironmentsRead   = "environments:read"
	ScopeEnvironmentsWrite  = "environments:write"
	ScopeUsersRead          = "users:read"
	ScopeUsersWrite         = "users:write"
	ScopeTokensRead         = "tokens:read"
	ScopeTokensWrite        = "tokens:write"
	ScopeMetricsRead        = "metrics:read"
	ScopeLogsRead           = "logs:read"
)

// AllScopes lists every scope value that can be requested at token creation.
// Used for validation in the application layer.
var AllScopes = []string{
	ScopeRoutesRead, ScopeRoutesWrite,
	ScopeClustersRead, ScopeClustersWrite,
	ScopeMiddlewaresRead, ScopeMiddlewaresWrite,
	ScopeEnvironmentsRead, ScopeEnvironmentsWrite,
	ScopeUsersRead, ScopeUsersWrite,
	ScopeTokensRead, ScopeTokensWrite,
	ScopeMetricsRead, ScopeLogsRead,
}

// APIToken is a programmatic access credential. The token's raw secret is
// shown to the user once at creation time and never persisted; only the
// bcrypt hash is stored.
type APIToken struct {
	ID         int        `json:"id"          db:"id"`
	UserID     int        `json:"user_id"     db:"user_id"`
	Name       string     `json:"name"        db:"name"`
	Prefix     string     `json:"prefix"      db:"prefix"`
	TokenHash  string     `json:"-"           db:"token_hash"`
	Scopes     string     `json:"scopes"      db:"scopes"` // comma-separated
	LastUsedAt *time.Time `json:"last_used_at,omitempty" db:"last_used_at"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"   db:"expires_at"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"   db:"revoked_at"`
	CreatedAt  time.Time  `json:"created_at"  db:"created_at"`
}

// HasScope reports whether the token grants the given scope (either directly
// or via the wildcard).
func (t APIToken) HasScope(scope string) bool {
	for _, s := range strings.Split(t.Scopes, ",") {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if s == ScopeWildcard || s == scope {
			return true
		}
	}
	return false
}

// IsActive returns true if the token can still be used: not revoked, and not
// expired (an empty ExpiresAt means "never expires").
func (t APIToken) IsActive(now time.Time) bool {
	if t.RevokedAt != nil {
		return false
	}
	if t.ExpiresAt != nil && !t.ExpiresAt.IsZero() && now.After(*t.ExpiresAt) {
		return false
	}
	return true
}
