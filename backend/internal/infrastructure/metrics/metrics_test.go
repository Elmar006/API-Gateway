package metrics

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestRegistryNew(t *testing.T) {
	r := New()
	if r.HTTPRequestsTotal == nil || r.HTTPRequestDuration == nil ||
		r.RateLimitHits == nil || r.ActiveRoutes == nil || r.DBConnections == nil {
		t.Fatal("nil collector")
	}
	r.RateLimitHits.Inc()
	r.ActiveRoutes.Set(3)
	if got := testutil.ToFloat64(r.RateLimitHits); got != 1 {
		t.Errorf("rate hits=%v", got)
	}
	if got := testutil.ToFloat64(r.ActiveRoutes); got != 3 {
		t.Errorf("active routes=%v", got)
	}
}

func TestHandlerExposesMetrics(t *testing.T) {
	r := New()
	r.RateLimitHits.Inc()
	srv := httptest.NewServer(r.Handler())
	defer srv.Close()
	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	body := string(buf)
	if !strings.Contains(body, "rate_limit_hits_total 1") {
		t.Errorf("metrics body missing rate hits: %s", body)
	}
}

func TestHTTPMiddlewareRecordsCounters(t *testing.T) {
	r := New()
	mw := r.HTTPMiddleware("public")
	rec := httptest.NewRecorder()
	mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		_, _ = w.Write([]byte("x"))
	})).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/x", nil))

	if got := testutil.CollectAndCount(r.HTTPRequestsTotal); got == 0 {
		t.Fatalf("no counter samples")
	}
}

func TestHTTPMiddlewareImplicit200(t *testing.T) {
	r := New()
	mw := r.HTTPMiddleware("admin")
	rec := httptest.NewRecorder()
	mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/y", nil))

	if rec.Code != http.StatusOK {
		t.Errorf("status=%d", rec.Code)
	}
}

func TestPoolWatcherNilPoolNoop(t *testing.T) {
	r := New()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	r.StartPoolWatcher(ctx, nil, time.Millisecond)
	// just ensure no panic; nothing else to assert.
}
