// Package limiter implements per-client rate limiting using sharded token
// buckets. Sharding lowers contention on the per-IP map under heavy load.
package limiter

import (
	"hash/fnv"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

const defaultShards = 16

// limiterEntry stores a token bucket and its last-touched timestamp.
type limiterEntry struct {
	limiter   *rate.Limiter
	rateLimit int
	lastUsed  int64 // unix nano
}

type shard struct {
	mu sync.Mutex
	m  map[string]*limiterEntry
}

// IPLimiter is a sharded per-IP token-bucket rate limiter that purges idle
// entries via a background goroutine.
type IPLimiter struct {
	shards []*shard
	ttl    time.Duration
	stop   chan struct{}
	once   sync.Once
}

// NewIPLimiter constructs an IPLimiter with the given idle TTL.
func NewIPLimiter(ttl time.Duration) *IPLimiter {
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	shards := make([]*shard, defaultShards)
	for i := range shards {
		shards[i] = &shard{m: make(map[string]*limiterEntry)}
	}
	l := &IPLimiter{
		shards: shards,
		ttl:    ttl,
		stop:   make(chan struct{}),
	}
	go l.cleanupLoop()
	return l
}

// Stop terminates the cleanup goroutine.
func (l *IPLimiter) Stop() {
	l.once.Do(func() { close(l.stop) })
}

// Allow returns true if the request from ip is admitted under the configured
// per-minute rate. ratePerMinute <= 0 disables limiting entirely.
func (l *IPLimiter) Allow(ip string, ratePerMinute int) bool {
	if ratePerMinute <= 0 {
		return true
	}
	sh := l.shardFor(ip)
	now := time.Now().UnixNano()

	sh.mu.Lock()
	entry, ok := sh.m[ip]
	if !ok || entry.rateLimit != ratePerMinute {
		entry = &limiterEntry{
			limiter:   rate.NewLimiter(rate.Limit(float64(ratePerMinute)/60.0), ratePerMinute),
			rateLimit: ratePerMinute,
		}
		sh.m[ip] = entry
	}
	entry.lastUsed = now
	allowed := entry.limiter.Allow()
	sh.mu.Unlock()
	return allowed
}

// Size returns the total number of tracked IPs across all shards.
func (l *IPLimiter) Size() int {
	total := 0
	for _, sh := range l.shards {
		sh.mu.Lock()
		total += len(sh.m)
		sh.mu.Unlock()
	}
	return total
}

func (l *IPLimiter) shardFor(ip string) *shard {
	h := fnv.New32a()
	_, _ = h.Write([]byte(ip))
	return l.shards[h.Sum32()%uint32(len(l.shards))]
}

func (l *IPLimiter) cleanupLoop() {
	ticker := time.NewTicker(l.ttl)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			cutoff := time.Now().Add(-l.ttl).UnixNano()
			for _, sh := range l.shards {
				sh.mu.Lock()
				for ip, e := range sh.m {
					if e.lastUsed < cutoff {
						delete(sh.m, ip)
					}
				}
				sh.mu.Unlock()
			}
		case <-l.stop:
			return
		}
	}
}
