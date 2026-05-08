// Package model contains domain models used across the application,
// including configuration structures for the App and Database.
package model

import "time"

// Config holds application-level configuration loaded from environment variables.
type Config struct {
	// HTTP servers.
	Port      string
	AdminPort string

	// Logging.
	LogLvl    string
	LogFormat string // "text" or "json"

	// Auth.
	JWTSecret     string
	AdminUsername string
	AdminPassword string

	// Limits.
	MaxBodyBytes int64
	RateLimitTTL time.Duration

	// TLS.
	TLSEnabled bool
	CertFile   string
	KeyFile    string

	// Async logger tuning.
	LogBufferSize  int
	LogBatchSize   int
	LogFlushPeriod time.Duration

	// Tracing. Enabled is read from OTEL_TRACES_ENABLED (default false). When
	// enabled, the standard OTEL_EXPORTER_OTLP_* env vars configure the
	// destination.
	TracingEnabled     bool
	TracingServiceName string
}

// DB describes the PostgreSQL connection settings.
type DB struct {
	Host         string
	Port         string
	User         string
	Password     string
	DBName       string
	SSLMode      string
	MaxOpenConns int
	MaxIdleConns int
	MaxLifetime  int // minutes
}

// DSN returns a libpq-style DSN string.
func (d DB) DSN() string {
	return "host=" + d.Host +
		" port=" + d.Port +
		" user=" + d.User +
		" password=" + d.Password +
		" dbname=" + d.DBName +
		" sslmode=" + d.SSLMode
}
