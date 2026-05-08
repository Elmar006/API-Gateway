package config

import (
	"strings"
	"testing"
	"time"

	"prodlich/internal/model"
)

func validBase() AppConfig {
	return AppConfig{
		App: model.Config{
			Port:           "8080",
			AdminPort:      "9090",
			LogLvl:         "info",
			LogFormat:      "text",
			JWTSecret:      "0123456789abcdef",
			AdminUsername:  "admin",
			AdminPassword:  "admin-password",
			MaxBodyBytes:   10 * 1024 * 1024,
			RateLimitTTL:   5 * time.Minute,
			LogBufferSize:  1024,
			LogBatchSize:   100,
			LogFlushPeriod: 200 * time.Millisecond,
		},
		DB: model.DB{
			Host:     "localhost",
			Port:     "5432",
			User:     "postgres",
			Password: "secret",
			DBName:   "gateway",
			SSLMode:  "disable",
		},
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(c *AppConfig)
		wantSub string // expected substring of the error
	}{
		{name: "ok", mutate: func(*AppConfig) {}},
		{name: "missing JWT secret", mutate: func(c *AppConfig) { c.App.JWTSecret = "" }, wantSub: "JWT_SECRET is required"},
		{name: "short JWT secret", mutate: func(c *AppConfig) { c.App.JWTSecret = "short" }, wantSub: "at least 16"},
		{name: "missing DB password", mutate: func(c *AppConfig) { c.DB.Password = "" }, wantSub: "DB_PASSWORD"},
		{name: "missing admin user", mutate: func(c *AppConfig) { c.App.AdminUsername = "" }, wantSub: "ADMIN_USERNAME"},
		{name: "missing admin pwd", mutate: func(c *AppConfig) { c.App.AdminPassword = "" }, wantSub: "ADMIN_PASSWORD"},
		{name: "tls without cert", mutate: func(c *AppConfig) { c.App.TLSEnabled = true }, wantSub: "CERT_FILE"},
		{name: "negative max body", mutate: func(c *AppConfig) { c.App.MaxBodyBytes = 0 }, wantSub: "MAX_BODY_BYTES"},
		{name: "buffer zero", mutate: func(c *AppConfig) { c.App.LogBufferSize = 0 }, wantSub: "LOG_BUFFER_SIZE"},
		{name: "batch larger than buffer", mutate: func(c *AppConfig) { c.App.LogBatchSize = 9999 }, wantSub: "LOG_BATCH_SIZE"},
		{name: "flush negative", mutate: func(c *AppConfig) { c.App.LogFlushPeriod = 0 }, wantSub: "LOG_FLUSH_PERIOD"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := validBase()
			tt.mutate(&c)
			err := c.Validate()
			if tt.wantSub == "" {
				if err != nil {
					t.Fatalf("expected nil error, got %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.wantSub)
			}
			if !strings.Contains(err.Error(), tt.wantSub) {
				t.Fatalf("error %q does not contain %q", err.Error(), tt.wantSub)
			}
		})
	}
}

func TestLoadFromEnv(t *testing.T) {
	t.Setenv("PORT", "8081")
	t.Setenv("ADMIN_PORT", "9091")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("LOG_FORMAT", "JSON")
	t.Setenv("JWT_SECRET", "0123456789abcdef")
	t.Setenv("ADMIN_USERNAME", "root")
	t.Setenv("ADMIN_PASSWORD", "rootpassword")
	t.Setenv("DB_PASSWORD", "secret")
	t.Setenv("DB_NAME", "gateway")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.App.Port != "8081" || cfg.App.AdminPort != "9091" || cfg.App.LogLvl != "debug" || cfg.App.LogFormat != "json" {
		t.Errorf("env not applied: %+v", cfg.App)
	}
	if cfg.DB.DBName != "gateway" {
		t.Errorf("db name=%q", cfg.DB.DBName)
	}
}

func TestDSN(t *testing.T) {
	dsn := (model.DB{Host: "h", Port: "5432", User: "u", Password: "p", DBName: "d", SSLMode: "disable"}).DSN()
	want := "host=h port=5432 user=u password=p dbname=d sslmode=disable"
	if dsn != want {
		t.Fatalf("dsn=%q", dsn)
	}
}

func TestEnvHelpers(t *testing.T) {
	t.Setenv("FOO_INT", "42")
	t.Setenv("FOO_INT64", "9876543210")
	t.Setenv("FOO_BOOL", "true")
	t.Setenv("FOO_DUR", "750ms")
	t.Setenv("FOO_STR", "value")

	if got := getEnv("FOO_STR", "default"); got != "value" {
		t.Errorf("getEnv: %s", got)
	}
	if got := getEnv("MISSING", "fallback"); got != "fallback" {
		t.Errorf("getEnv default: %s", got)
	}
	if got := getEnvInt("FOO_INT", 0); got != 42 {
		t.Errorf("getEnvInt: %d", got)
	}
	if got := getEnvInt("MISSING_INT", 7); got != 7 {
		t.Errorf("getEnvInt default: %d", got)
	}
	if got := getEnvInt64("FOO_INT64", 0); got != 9876543210 {
		t.Errorf("getEnvInt64: %d", got)
	}
	if got := getEnvBool("FOO_BOOL", false); !got {
		t.Errorf("getEnvBool: %v", got)
	}
	if got := getEnvDuration("FOO_DUR", time.Second); got != 750*time.Millisecond {
		t.Errorf("getEnvDuration: %v", got)
	}
}
