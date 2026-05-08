// Package entity defines domain models used across the gateway.
package entity

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

// RequestLog is a record of a single proxied request, persisted asynchronously.
type RequestLog struct {
	ID             int64     `json:"id" db:"id"`
	RequestID      uuid.UUID `json:"request_id" db:"request_id"`
	Method         string    `json:"method" db:"method"`
	Path           string    `json:"path" db:"path"`
	Query          *string   `json:"query,omitempty" db:"query"`
	RouteID        *int      `json:"route_id,omitempty" db:"route_id"`
	UserID         *int      `json:"user_id,omitempty" db:"user_id"`
	ClientIP       string    `json:"client_ip" db:"client_ip"`
	StatusCode     int       `json:"status_code" db:"status_code"`
	ResponseTimeMs int       `json:"response_time_ms" db:"response_time_ms"`
	TargetURL      *string   `json:"target_url,omitempty" db:"target_url"`
	TraceID        *string   `json:"trace_id,omitempty"  db:"trace_id"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}
