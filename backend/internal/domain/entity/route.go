// Package entity defines domain models used across the gateway.
package entity

import (
	"strings"
	"time"
)

// HTTPMethodAll is a sentinel that matches any HTTP method.
const HTTPMethodAll = "ALL"

// Route represents a dynamic proxy rule loaded from the database.
//
// A route forwards requests to either a single TargetURL or, if ClusterID is
// set, to one of the targets in that cluster (chosen by the cluster's
// strategy). EnvironmentID is purely organisational and does not affect
// routing decisions.
type Route struct {
	ID            int       `json:"id"             db:"id"`
	Method        string    `json:"method"         db:"method"`
	PathPattern   string    `json:"path_pattern"   db:"path_pattern"`
	TargetURL     string    `json:"target_url"     db:"target_url"`
	IsActive      bool      `json:"is_active"      db:"is_active"`
	Priority      int       `json:"priority"       db:"priority"`
	RateLimit     int       `json:"rate_limit"     db:"rate_limit"`
	RequireAuth   bool      `json:"require_auth"   db:"require_auth"`
	EnvironmentID *int      `json:"environment_id,omitempty" db:"environment_id"`
	ClusterID     *int      `json:"cluster_id,omitempty"     db:"cluster_id"`
	CreatedAt     time.Time `json:"created_at"     db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"     db:"updated_at"`
}

// NormalizeMethod returns Method in upper case, defaulting to ALL when empty.
func (r Route) NormalizeMethod() string {
	m := strings.TrimSpace(strings.ToUpper(r.Method))
	if m == "" {
		return HTTPMethodAll
	}
	return m
}
