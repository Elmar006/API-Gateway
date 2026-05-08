// Package log – async, batched request-log writer.
package log

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"prodlich/internal/domain/entity"
	"prodlich/internal/domain/repository"
)

// BatchInserter persists request logs in batches. It is implemented by the
// PostgreSQL LogRepo (CopyFrom) but kept narrow for testability.
type BatchInserter interface {
	repository.LogRepo
	InsertBatch(ctx context.Context, logs []entity.RequestLog) error
}

// AsyncLogger buffers entity.RequestLog values and flushes them to the database
// in batches – either when the batch is full or when the flush ticker fires.
//
// Drops are counted via Dropped() and never block the request path. Closing
// the logger flushes any pending entries.
type AsyncLogger struct {
	repo       BatchInserter
	ch         chan entity.RequestLog
	wg         sync.WaitGroup
	dropped    atomic.Uint64
	closeOnce  sync.Once
	stop       chan struct{}
	batchSize  int
	flushEvery time.Duration
}

// NewAsyncLogger returns a started AsyncLogger. Buffer is the channel capacity,
// batchSize controls the maximum batch size, flushEvery is the periodic flush
// interval used while the buffer is partially full.
func NewAsyncLogger(repo BatchInserter, buffer, batchSize int, flushEvery time.Duration) *AsyncLogger {
	if buffer <= 0 {
		buffer = 1024
	}
	if batchSize <= 0 || batchSize > buffer {
		batchSize = buffer
	}
	if flushEvery <= 0 {
		flushEvery = 500 * time.Millisecond
	}
	al := &AsyncLogger{
		repo:       repo,
		ch:         make(chan entity.RequestLog, buffer),
		stop:       make(chan struct{}),
		batchSize:  batchSize,
		flushEvery: flushEvery,
	}
	al.wg.Add(1)
	go al.worker()
	return al
}

// Send enqueues a log entry without blocking; if the buffer is full, the
// entry is dropped and the dropped counter is incremented.
func (al *AsyncLogger) Send(entry entity.RequestLog) {
	select {
	case al.ch <- entry:
	default:
		al.dropped.Add(1)
	}
}

// Dropped returns the number of log entries dropped due to a full buffer.
func (al *AsyncLogger) Dropped() uint64 { return al.dropped.Load() }

// Close stops the worker, flushing any pending entries.
func (al *AsyncLogger) Close() {
	al.closeOnce.Do(func() {
		close(al.stop)
		al.wg.Wait()
	})
}

func (al *AsyncLogger) worker() {
	defer al.wg.Done()
	ticker := time.NewTicker(al.flushEvery)
	defer ticker.Stop()

	batch := make([]entity.RequestLog, 0, al.batchSize)
	flush := func() {
		if len(batch) == 0 {
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := al.repo.InsertBatch(ctx, batch); err != nil {
			L().WithError(err).Warn("async logger: batch insert failed")
		}
		batch = batch[:0]
	}

	for {
		select {
		case entry := <-al.ch:
			batch = append(batch, entry)
			if len(batch) >= al.batchSize {
				flush()
			}
		case <-ticker.C:
			flush()
		case <-al.stop:
			// Drain channel (best-effort) before exiting.
			for {
				select {
				case entry := <-al.ch:
					batch = append(batch, entry)
					if len(batch) >= al.batchSize {
						flush()
					}
				default:
					flush()
					return
				}
			}
		}
	}
}
