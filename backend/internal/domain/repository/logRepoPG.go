// Package repository defines interfaces for data persistence and access patterns.
package repository

import (
	"context"
	"prodlich/internal/domain/entity"
	"time"
)

// LogFilters holds query parameters for filtering request logs.
type LogFilters struct {
	Path       string
	StatusCode int
	From       *time.Time
	To         *time.Time
}

type LogRepo interface {
	Insert(ctx context.Context, log entity.RequestLog) error
	GetWithFilters(ctx context.Context, filters LogFilters, limit, offset int) ([]entity.RequestLog, int, error)
	GetMetrics(ctx context.Context, from, to time.Time) (entity.Metrics, error) // новый метод
}
