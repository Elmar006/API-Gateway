package entity

import "time"

// Middleware kinds supported by the runtime resolver.
const (
	MiddlewareKindCORS              = "cors"
	MiddlewareKindHeaderRewrite     = "header_rewrite"
	MiddlewareKindRateLimitOverride = "rate_limit_override"
	MiddlewareKindRequestID         = "request_id"
	MiddlewareKindStripPrefix       = "strip_prefix"
)

// Middleware is a reusable per-route handler defined in the database.
// Config is a free-form JSON document; its shape depends on Kind and is
// validated at attachment time by the application layer.
type Middleware struct {
	ID          int       `json:"id"          db:"id"`
	Name        string    `json:"name"        db:"name"`
	Kind        string    `json:"kind"        db:"kind"`
	Config      []byte    `json:"-"           db:"config"`
	ConfigJSON  any       `json:"config"`
	IsActive    bool      `json:"is_active"   db:"is_active"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at"  db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"  db:"updated_at"`
}

// RouteMiddleware is the join row binding a middleware to a route in a
// specific position.
type RouteMiddleware struct {
	RouteID      int       `json:"route_id"      db:"route_id"`
	MiddlewareID int       `json:"middleware_id" db:"middleware_id"`
	SortOrder    int       `json:"sort_order"    db:"sort_order"`
	CreatedAt    time.Time `json:"created_at"    db:"created_at"`
}
