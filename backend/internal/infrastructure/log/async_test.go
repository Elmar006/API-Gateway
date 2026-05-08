package log

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"prodlich/internal/domain/entity"
	"prodlich/internal/domain/repository"
)

type stubBatchRepo struct {
	mu       sync.Mutex
	received []entity.RequestLog
	calls    atomic.Int64
}

func (s *stubBatchRepo) Insert(_ context.Context, e entity.RequestLog) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.received = append(s.received, e)
	return nil
}

func (s *stubBatchRepo) GetWithFilters(context.Context, repository.LogFilters, int, int) ([]entity.RequestLog, int, error) {
	return nil, 0, nil
}

func (s *stubBatchRepo) GetMetrics(context.Context, time.Time, time.Time) (entity.Metrics, error) {
	return entity.Metrics{}, nil
}

func (s *stubBatchRepo) InsertBatch(_ context.Context, batch []entity.RequestLog) error {
	s.calls.Add(1)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.received = append(s.received, batch...)
	return nil
}

func TestAsyncLoggerFlushOnFullBatch(t *testing.T) {
	repo := &stubBatchRepo{}
	al := NewAsyncLogger(repo, 16, 4, time.Second)
	defer al.Close()
	for i := 0; i < 8; i++ {
		al.Send(entity.RequestLog{Method: "GET", Path: "/x"})
	}
	// flush by closing
	al.Close()
	repo.mu.Lock()
	got := len(repo.received)
	repo.mu.Unlock()
	if got != 8 {
		t.Fatalf("received=%d", got)
	}
}

func TestAsyncLoggerFlushOnTicker(t *testing.T) {
	repo := &stubBatchRepo{}
	al := NewAsyncLogger(repo, 16, 100, 10*time.Millisecond)
	defer al.Close()
	al.Send(entity.RequestLog{Method: "GET"})
	time.Sleep(50 * time.Millisecond)
	repo.mu.Lock()
	got := len(repo.received)
	repo.mu.Unlock()
	if got != 1 {
		t.Fatalf("expected 1 flushed entry, got %d", got)
	}
}

func TestAsyncLoggerDrops(t *testing.T) {
	// 0-buffer is replaced with 1024 by NewAsyncLogger; use a small buffer and
	// stall the worker by passing a slow repo.
	slow := &slowRepo{delay: 50 * time.Millisecond}
	al := NewAsyncLogger(slow, 2, 2, time.Second)
	defer al.Close()
	for i := 0; i < 50; i++ {
		al.Send(entity.RequestLog{})
	}
	if al.Dropped() == 0 {
		t.Fatal("expected drops")
	}
}

type slowRepo struct {
	stubBatchRepo
	delay time.Duration
}

func (s *slowRepo) InsertBatch(_ context.Context, batch []entity.RequestLog) error {
	time.Sleep(s.delay)
	return s.stubBatchRepo.InsertBatch(context.Background(), batch)
}
