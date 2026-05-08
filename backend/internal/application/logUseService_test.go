package application

import (
	"context"
	"testing"
	"time"

	"prodlich/internal/domain/entity"
	"prodlich/internal/domain/repository"
)

type fakeLogRepo struct {
	gotLimit, gotOffset int
	gotFilters          repository.LogFilters
}

func (f *fakeLogRepo) Insert(_ context.Context, _ entity.RequestLog) error { return nil }
func (f *fakeLogRepo) GetWithFilters(_ context.Context, fl repository.LogFilters, limit, offset int) ([]entity.RequestLog, int, error) {
	f.gotLimit = limit
	f.gotOffset = offset
	f.gotFilters = fl
	return []entity.RequestLog{{ID: 1}}, 1, nil
}
func (f *fakeLogRepo) GetMetrics(_ context.Context, _, _ time.Time) (entity.Metrics, error) {
	return entity.Metrics{TotalRequests: 7}, nil
}

func TestLogUseCaseClampsPagination(t *testing.T) {
	repo := &fakeLogRepo{}
	uc := NewLogUseCase(repo)
	tests := []struct{ in, want int }{{0, 50}, {-5, 50}, {2000, 1000}, {25, 25}}
	for _, tt := range tests {
		_, _, err := uc.GetLogs(context.Background(), repository.LogFilters{}, tt.in, -1)
		if err != nil {
			t.Fatal(err)
		}
		if repo.gotLimit != tt.want {
			t.Errorf("limit in=%d -> got %d, want %d", tt.in, repo.gotLimit, tt.want)
		}
		if repo.gotOffset != 0 {
			t.Errorf("offset clamp failed: %d", repo.gotOffset)
		}
	}
}

func TestMetricUseCasePeriods(t *testing.T) {
	repo := &fakeLogRepo{}
	uc := NewMetricUseCase(repo)
	for _, p := range []string{"hour", "day", "week", ""} {
		m, err := uc.GetMetrics(context.Background(), p)
		if err != nil {
			t.Fatal(err)
		}
		if m.TotalRequests != 7 {
			t.Errorf("period %q: total=%d", p, m.TotalRequests)
		}
	}
}
