// Package config loads, validates, and exposes runtime configuration sourced
// from environment variables (with optional .env hydration via Load).
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"prodlich/internal/model"
)

// AppConfig is the fully validated application + database configuration.
type AppConfig struct {
	App model.Config
	DB  model.DB
}

// Load reads environment variables, applies defaults, validates required fields
// and returns the resulting AppConfig. It also lightly hydrates from a `.env`
// file in the current working directory if present (without overriding env vars).
func Load() (AppConfig, error) {
	hydrateFromDotEnv(".env")

	cfg := AppConfig{
		App: model.Config{
			Port:           getEnv("PORT", "8080"),
			AdminPort:      getEnv("ADMIN_PORT", "9090"),
			LogLvl:         getEnv("LOG_LEVEL", "info"),
			LogFormat:      strings.ToLower(getEnv("LOG_FORMAT", "text")),
			JWTSecret:      getEnv("JWT_SECRET", ""),
			AdminUsername:  getEnv("ADMIN_USERNAME", ""),
			AdminPassword:  getEnv("ADMIN_PASSWORD", ""),
			MaxBodyBytes:   getEnvInt64("MAX_BODY_BYTES", 10*1024*1024),
			RateLimitTTL:   getEnvDuration("RATE_LIMIT_TTL", 5*time.Minute),
			TLSEnabled:     getEnvBool("TLS_ENABLED", false),
			CertFile:       getEnv("CERT_FILE", ""),
			KeyFile:        getEnv("KEY_FILE", ""),
			LogBufferSize:      getEnvInt("LOG_BUFFER_SIZE", 4096),
			LogBatchSize:       getEnvInt("LOG_BATCH_SIZE", 200),
			LogFlushPeriod:     getEnvDuration("LOG_FLUSH_PERIOD", 500*time.Millisecond),
			TracingEnabled:     getEnvBool("OTEL_TRACES_ENABLED", false),
			TracingServiceName: getEnv("OTEL_SERVICE_NAME", "routeflow-gateway"),
		},
		DB: model.DB{
			Host:         getEnv("DB_HOST", "localhost"),
			Port:         getEnv("DB_PORT", "5432"),
			User:         getEnv("DB_USER", "postgres"),
			Password:     getEnv("DB_PASSWORD", ""),
			DBName:       getEnv("DB_NAME", "postgres"),
			SSLMode:      getEnv("DB_SSLMODE", "disable"),
			MaxOpenConns: getEnvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns: getEnvInt("DB_MAX_IDLE_CONNS", 10),
			MaxLifetime:  getEnvInt("DB_MAX_LIFETIME_MIN", 30),
		},
	}

	if err := cfg.Validate(); err != nil {
		return AppConfig{}, err
	}
	return cfg, nil
}

// Validate verifies that all mandatory fields are populated and consistent.
func (c AppConfig) Validate() error {
	var errs []string

	required := map[string]string{
		"JWT_SECRET":     c.App.JWTSecret,
		"DB_PASSWORD":    c.DB.Password,
		"ADMIN_USERNAME": c.App.AdminUsername,
		"ADMIN_PASSWORD": c.App.AdminPassword,
	}
	for k, v := range required {
		if strings.TrimSpace(v) == "" {
			errs = append(errs, fmt.Sprintf("%s is required", k))
		}
	}
	if len(c.App.JWTSecret) > 0 && len(c.App.JWTSecret) < 16 {
		errs = append(errs, "JWT_SECRET must be at least 16 characters long")
	}
	if c.App.TLSEnabled {
		if c.App.CertFile == "" || c.App.KeyFile == "" {
			errs = append(errs, "TLS_ENABLED requires CERT_FILE and KEY_FILE")
		}
	}
	if c.App.MaxBodyBytes <= 0 {
		errs = append(errs, "MAX_BODY_BYTES must be > 0")
	}
	if c.App.LogBufferSize <= 0 {
		errs = append(errs, "LOG_BUFFER_SIZE must be > 0")
	}
	if c.App.LogBatchSize <= 0 || c.App.LogBatchSize > c.App.LogBufferSize {
		errs = append(errs, "LOG_BATCH_SIZE must be in (0, LOG_BUFFER_SIZE]")
	}
	if c.App.LogFlushPeriod <= 0 {
		errs = append(errs, "LOG_FLUSH_PERIOD must be > 0")
	}

	if len(errs) > 0 {
		return errors.New("invalid configuration: " + strings.Join(errs, "; "))
	}
	return nil
}

// hydrateFromDotEnv reads simple KEY=VALUE pairs from a file (if it exists)
// and sets them in the process env unless already set. We intentionally do
// not pull in a third-party dependency for this trivial loader.
func hydrateFromDotEnv(path string) {
	data, err := os.ReadFile(path) // #nosec G304 - path is a fixed config file
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		eq := strings.IndexByte(line, '=')
		if eq <= 0 {
			continue
		}
		key := strings.TrimSpace(line[:eq])
		val := strings.TrimSpace(line[eq+1:])
		val = strings.Trim(val, `"'`)
		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, val)
		}
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return defaultValue
}

func getEnvInt64(key string, defaultValue int64) int64 {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			return n
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return defaultValue
}
