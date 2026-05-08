// Package metrics holds Prometheus collectors and helpers used by both the
// public proxy and the admin API.
package metrics

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Registry holds all collectors. Use it instead of prometheus.DefaultRegisterer
// so tests can spin up isolated copies.
type Registry struct {
	Reg *prometheus.Registry

	HTTPRequestsTotal   *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec
	RateLimitHits       prometheus.Counter
	ActiveRoutes        prometheus.Gauge
	DBConnections       *prometheus.GaugeVec
}

// New returns a fully wired Registry with default collectors and gateway-specific metrics.
func New() *Registry {
	reg := prometheus.NewRegistry()

	r := &Registry{
		Reg: reg,
		HTTPRequestsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of handled HTTP requests, partitioned by component, method and status.",
		}, []string{"component", "method", "status"}),
		HTTPRequestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Latency of handled HTTP requests in seconds.",
			Buckets: prometheus.DefBuckets,
		}, []string{"component", "method"}),
		RateLimitHits: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "rate_limit_hits_total",
			Help: "Total number of requests rejected by the public rate limiter.",
		}),
		ActiveRoutes: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "active_routes_count",
			Help: "Current number of active routes loaded into the in-memory cache.",
		}),
		DBConnections: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "db_connections",
			Help: "Current pgxpool connection counters.",
		}, []string{"state"}),
	}

	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		r.HTTPRequestsTotal,
		r.HTTPRequestDuration,
		r.RateLimitHits,
		r.ActiveRoutes,
		r.DBConnections,
	)
	return r
}

// Handler returns the http.Handler that exposes the registry to scrapers.
func (r *Registry) Handler() http.Handler {
	return promhttp.HandlerFor(r.Reg, promhttp.HandlerOpts{Registry: r.Reg})
}

// HTTPMiddleware wraps the next handler, recording counters and a histogram.
// component is a free-form label (e.g. "public", "admin").
func (r *Registry) HTTPMiddleware(component string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			start := time.Now()
			ww := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(ww, req)
			r.HTTPRequestsTotal.WithLabelValues(component, req.Method, strconv.Itoa(ww.status)).Inc()
			r.HTTPRequestDuration.WithLabelValues(component, req.Method).Observe(time.Since(start).Seconds())
		})
	}
}

// StartPoolWatcher periodically copies pgxpool stats into the gauge vector.
func (r *Registry) StartPoolWatcher(ctx context.Context, pool *pgxpool.Pool, every time.Duration) {
	if pool == nil {
		return
	}
	if every <= 0 {
		every = 5 * time.Second
	}
	go func() {
		t := time.NewTicker(every)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				stat := pool.Stat()
				r.DBConnections.WithLabelValues("acquired").Set(float64(stat.AcquiredConns()))
				r.DBConnections.WithLabelValues("idle").Set(float64(stat.IdleConns()))
				r.DBConnections.WithLabelValues("total").Set(float64(stat.TotalConns()))
				r.DBConnections.WithLabelValues("max").Set(float64(stat.MaxConns()))
			}
		}
	}()
}

type statusRecorder struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (s *statusRecorder) WriteHeader(code int) {
	if s.wroteHeader {
		return
	}
	s.status = code
	s.wroteHeader = true
	s.ResponseWriter.WriteHeader(code)
}

func (s *statusRecorder) Write(b []byte) (int, error) {
	if !s.wroteHeader {
		s.wroteHeader = true
	}
	return s.ResponseWriter.Write(b)
}
