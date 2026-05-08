// Package db – migrations runner that uses golang-migrate with embed.FS.
package db

import (
	"errors"
	"fmt"
	"net/url"

	"prodlich/internal/infrastructure/log"
	"prodlich/internal/model"
	"prodlich/migrations"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // postgres driver
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// MigrateUp applies all pending migrations from the embedded migrations FS.
func MigrateUp(cfg *model.DB) error {
	src, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("init iofs source: %w", err)
	}
	url := postgresURL(cfg)
	m, err := migrate.NewWithSourceInstance("iofs", src, url)
	if err != nil {
		return fmt.Errorf("init migrate: %w", err)
	}
	defer func() {
		_, _ = m.Close()
	}()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate up: %w", err)
	}
	v, dirty, err := m.Version()
	switch {
	case errors.Is(err, migrate.ErrNilVersion):
		log.L().Info("migrations: no migrations applied")
	case err != nil:
		return fmt.Errorf("read migrate version: %w", err)
	default:
		log.L().WithField("version", v).WithField("dirty", dirty).Info("migrations applied")
	}
	return nil
}

// MigrateDown rolls back N migrations (or all if n == -1).
func MigrateDown(cfg *model.DB, n int) error {
	src, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("init iofs source: %w", err)
	}
	m, err := migrate.NewWithSourceInstance("iofs", src, postgresURL(cfg))
	if err != nil {
		return fmt.Errorf("init migrate: %w", err)
	}
	defer func() {
		_, _ = m.Close()
	}()
	var execErr error
	if n < 0 {
		execErr = m.Down()
	} else {
		execErr = m.Steps(-n)
	}
	if execErr != nil && !errors.Is(execErr, migrate.ErrNoChange) {
		return fmt.Errorf("migrate down: %w", execErr)
	}
	return nil
}

// postgresURL builds a postgres:// connection URL for golang-migrate.
//
// Uses net/url so that user/password/host/db are properly percent-encoded.
// A naive fmt.Sprintf produces malformed URLs when any field contains a
// reserved character (e.g. a password with `@`, `:`, `/` or `?` would either
// be silently misinterpreted by the URL parser or break the connection
// outright). golang-migrate itself parses this string with net/url, so the
// only correct way to construct it is via the same package.
func postgresURL(cfg *model.DB) string {
	host := cfg.Host
	if cfg.Port != "" {
		host = cfg.Host + ":" + cfg.Port
	}
	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.User, cfg.Password),
		Host:   host,
		Path:   "/" + cfg.DBName,
	}
	q := u.Query()
	if cfg.SSLMode != "" {
		q.Set("sslmode", cfg.SSLMode)
	}
	u.RawQuery = q.Encode()
	return u.String()
}
