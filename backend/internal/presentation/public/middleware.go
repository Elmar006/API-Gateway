// Package public middleware utilities for request processing in the public proxy.
package public

import (
	"bytes"
	"context"
	"net"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"prodlich/internal/domain/entity"
	"prodlich/internal/infrastructure/log"
	"prodlich/internal/infrastructure/tracing"

	"github.com/gofrs/uuid/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

type ctxKey string

const (
	ctxRequestID ctxKey = "request_id"
	ctxRouteID   ctxKey = "route_id"
)

// asyncLogger is the narrow interface required by LoggingMiddleware.
type asyncLogger interface {
	Send(entry entity.RequestLog)
}

// rateHitsCounter is incremented every time the limiter rejects a request.
type rateHitsCounter interface {
	Inc()
}

// pool of responseWriterWrappers to avoid per-request allocation.
var rwPool = sync.Pool{
	New: func() any { return &responseWriterWrapper{} },
}

type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode  int
	wroteHeader bool
}

func acquireRW(w http.ResponseWriter) *responseWriterWrapper {
	rw := rwPool.Get().(*responseWriterWrapper)
	rw.ResponseWriter = w
	rw.statusCode = http.StatusOK
	rw.wroteHeader = false
	return rw
}

func releaseRW(rw *responseWriterWrapper) {
	rw.ResponseWriter = nil
	rwPool.Put(rw)
}

func (rw *responseWriterWrapper) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}
	rw.statusCode = code
	rw.wroteHeader = true
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriterWrapper) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.wroteHeader = true
	}
	return rw.ResponseWriter.Write(b)
}

// Hijack delegates to the underlying ResponseWriter when it implements
// http.Hijacker (required for WebSocket upgrades).
func (rw *responseWriterWrapper) Unwrap() http.ResponseWriter { return rw.ResponseWriter }

// RequestIDMiddleware injects (or propagates) X-Request-Id.
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := strings.TrimSpace(r.Header.Get("X-Request-Id"))
		if _, err := uuid.FromString(reqID); err != nil || reqID == "" {
			reqID = uuid.Must(uuid.NewV4()).String()
		}
		ctx := context.WithValue(r.Context(), ctxRequestID, reqID)
		w.Header().Set("X-Request-Id", reqID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// TracingMiddleware extracts/creates an OpenTelemetry span for each request
// and stamps the resulting trace_id into the response headers (X-Trace-Id) so
// downstream tools can correlate logs with traces. When tracing is disabled
// (no-op TracerProvider) the span is essentially free.
func TracingMiddleware(next http.Handler) http.Handler {
	tracer := otel.Tracer("prodlich/public")
	propagator := otel.GetTextMapPropagator()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := propagator.Extract(r.Context(), propagation.HeaderCarrier(r.Header))
		ctx, span := tracer.Start(ctx, r.Method+" "+r.URL.Path)
		defer span.End()
		if id := tracing.SpanContextID(ctx); id != "" {
			w.Header().Set("X-Trace-Id", id)
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// SecurityHeadersMiddleware adds OWASP-recommended response headers.
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		next.ServeHTTP(w, r)
	})
}

// MaxBodyBytesMiddleware limits the request body size on the public proxy.
func MaxBodyBytesMiddleware(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if maxBytes > 0 && r.Body != nil {
				r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RecoveryMiddleware recovers from panics and logs the stack trace.
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				stack := bytes.TrimSpace(debug.Stack())
				log.L().WithFields(map[string]any{
					"path":       r.URL.Path,
					"method":     r.Method,
					"request_id": r.Header.Get("X-Request-Id"),
					"panic":      rec,
				}).Errorf("recovered from panic\n%s", stack)
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("Internal Server Error"))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// LoggingMiddleware records each request to the async logger and structured logs.
func LoggingMiddleware(alog asyncLogger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := acquireRW(w)
			defer releaseRW(ww)

			holder := new(int)
			ctx := context.WithValue(r.Context(), ctxRouteID, holder)
			next.ServeHTTP(ww, r.WithContext(ctx))

			duration := time.Since(start)
			reqID, _ := r.Context().Value(ctxRequestID).(string)
			parsedID, err := uuid.FromString(reqID)
			if err != nil {
				parsedID = uuid.Must(uuid.NewV4())
			}

			var query *string
			if q := r.URL.RawQuery; q != "" {
				query = &q
			}
			var routeID *int
			if *holder > 0 {
				rid := *holder
				routeID = &rid
			}

			ip := clientIP(r)

			var tracePtr *string
			if id := tracing.SpanContextID(r.Context()); id != "" {
				tracePtr = &id
			}

			entry := entity.RequestLog{
				RequestID:      parsedID,
				Method:         r.Method,
				Path:           r.URL.Path,
				Query:          query,
				RouteID:        routeID,
				ClientIP:       ip,
				StatusCode:     ww.statusCode,
				ResponseTimeMs: int(duration.Milliseconds()),
				TraceID:        tracePtr,
				CreatedAt:      time.Now().UTC(),
			}
			if alog != nil {
				alog.Send(entry)
			}

			log.L().WithFields(map[string]any{
				"component":   "public",
				"method":      r.Method,
				"path":        r.URL.Path,
				"status":      ww.statusCode,
				"duration_ms": duration.Milliseconds(),
				"client_ip":   ip,
				"request_id":  reqID,
			}).Info("request handled")
		})
	}
}

// setRouteID writes the matched route id into the holder placed on the
// context by LoggingMiddleware. It is a no-op when the holder is absent
// (e.g. when the proxy is invoked without the middleware stack).
func setRouteID(ctx context.Context, id int) {
	if h, ok := ctx.Value(ctxRouteID).(*int); ok && h != nil {
		*h = id
	}
}

// clientIP returns the originating IP, honouring X-Forwarded-For and X-Real-IP
// (only the first hop is trusted; deployments behind multiple proxies should
// configure the upstream to overwrite these headers).
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if comma := strings.IndexByte(xff, ','); comma > 0 {
			xff = xff[:comma]
		}
		ip := strings.TrimSpace(xff)
		if ip != "" {
			return ip
		}
	}
	if xri := strings.TrimSpace(r.Header.Get("X-Real-IP")); xri != "" {
		return xri
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
