// Package db provides the database connection pool and lifecycle management.
package db

import (
	"context"
	"fmt"
	"time"

	"prodlich/internal/infrastructure/log"
	"prodlich/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Pool is the global pgxpool.Pool instance shared across the application.
var Pool *pgxpool.Pool

// Connect establishes a new database connection pool using the provided config
// and stores it in the package-level Pool variable.
func Connect(ctx context.Context, cfg *model.DB) error {
	pool, err := New(ctx, cfg)
	if err != nil {
		return err
	}
	Pool = pool
	return nil
}

// New creates a new pgxpool.Pool independent of the package-level Pool. This is
// the preferred constructor for tests and embedded use cases.
func New(ctx context.Context, cfg *model.DB) (*pgxpool.Pool, error) {
	pgConfig, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("parse pgx config: %w", err)
	}
	pgConfig.MaxConns = int32(cfg.MaxOpenConns)
	pgConfig.MinConns = int32(cfg.MaxIdleConns)
	pgConfig.MaxConnLifetime = time.Duration(cfg.MaxLifetime) * time.Minute

	connectCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(connectCtx, pgConfig)
	if err != nil {
		return nil, fmt.Errorf("create pgx pool: %w", err)
	}
	if err := pool.Ping(connectCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	log.L().Info("database connected")
	return pool, nil
}

// Close shuts down the package-level connection pool, if any.
func Close() {
	if Pool != nil {
		log.L().Info("closing database pool")
		Pool.Close()
		Pool = nil
	}
}

// HealthCheck performs a lightweight ping to verify database connectivity.
func HealthCheck(ctx context.Context) error {
	if Pool == nil {
		return fmt.Errorf("db pool is not initialized")
	}
	return Pool.Ping(ctx)
}
