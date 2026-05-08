// Package application implements the business logic layer (use cases) of the API Gateway.
package application

import (
	"context"

	"prodlich/internal/domain/entity"
	"prodlich/internal/domain/repository"
)

// LogUseCase handles retrieval and filtering of request logs for the admin dashboard.
type LogUseCase struct {
	logRepo repository.LogRepo
}

func NewLogUseCase(logRepo repository.LogRepo) *LogUseCase {
	return &LogUseCase{logRepo: logRepo}
}

func (uc *LogUseCase) GetLogs(ctx context.Context, filters repository.LogFilters, limit, offset int) ([]entity.RequestLog, int, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 1000 {
		limit = 1000
	}
	if offset < 0 {
		offset = 0
	}
	return uc.logRepo.GetWithFilters(ctx, filters, limit, offset)
}
