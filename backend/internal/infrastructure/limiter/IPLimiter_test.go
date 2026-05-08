package limiter

import (
	"sync"
	"testing"
	"time"
)

func TestAllowDisabledByZeroLimit(t *testing.T) {
	l := NewIPLimiter(time.Minute)
	defer l.Stop()
	if !l.Allow("1.1.1.1", 0) {
		t.Fatal("expected 0 limit to bypass limiting")
	}
	if !l.Allow("1.1.1.1", -1) {
		t.Fatal("expected negative limit to bypass limiting")
	}
}

func TestAllowEnforcesBucket(t *testing.T) {
	l := NewIPLimiter(time.Minute)
	defer l.Stop()
	const ip = "10.0.0.1"
	const limit = 3 // 3 req/min
	allowed := 0
	for i := 0; i < 50; i++ {
		if l.Allow(ip, limit) {
			allowed++
		}
	}
	if allowed > limit {
		t.Fatalf("burst exceeded: allowed=%d, want<=%d", allowed, limit)
	}
}

func TestSharding(t *testing.T) {
	l := NewIPLimiter(time.Minute)
	defer l.Stop()
	ips := []string{"a", "b", "c", "d", "e"}
	for _, ip := range ips {
		l.Allow(ip, 100)
	}
	if l.Size() != len(ips) {
		t.Fatalf("size=%d, want %d", l.Size(), len(ips))
	}
}

func TestConcurrent(t *testing.T) {
	l := NewIPLimiter(time.Minute)
	defer l.Stop()
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			ip := "ip-" + string(rune('a'+i%5))
			for j := 0; j < 100; j++ {
				l.Allow(ip, 1000)
			}
		}(i)
	}
	wg.Wait()
}
