// Package application implements the business logic layer (use cases) of the API Gateway.
package application

import (
	"context"
	"prodlich/internal/domain/entity"
	"time"
)

// MetricRepo defines the interface for retrieving aggregated metrics data.
type MetricRepo interface {
	GetMetrics(ctx context.Context, from, to time.Time) (entity.Metrics, error)
}

type MetricUseCase struct {
	metricRepo MetricRepo
}

func NewMetricUseCase(metricRepo MetricRepo) *MetricUseCase {
	return &MetricUseCase{metricRepo: metricRepo}
}

func (uc *MetricUseCase) GetMetrics(ctx context.Context, period string) (entity.Metrics, error) {
	now := time.Now()
	var from time.Time

	switch period {
	case "hour":
		from = now.Add(-1 * time.Hour)
	case "day":
		from = now.Add(-24 * time.Hour)
	case "week":
		from = now.Add(-7 * 24 * time.Hour)
	default:
		from = now.Add(-24 * time.Hour)
	}

	return uc.metricRepo.GetMetrics(ctx, from, now)
}
