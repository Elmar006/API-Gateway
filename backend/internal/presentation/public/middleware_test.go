package public

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"prodlich/internal/domain/entity"
)

func TestSecurityHeadersMiddleware(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	SecurityHeadersMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rec, req)
	for _, h := range []string{"X-Content-Type-Options", "X-Frame-Options", "Referrer-Policy"} {
		if rec.Header().Get(h) == "" {
			t.Errorf("missing %s", h)
		}
	}
}

func TestMaxBodyBytesMiddleware(t *testing.T) {
	body := bytes.Repeat([]byte("a"), 4096)
	req := httptest.NewRequest(http.MethodPost, "/", io.NopCloser(bytes.NewReader(body)))
	rec := httptest.NewRecorder()
	MaxBodyBytesMiddleware(1024)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := io.ReadAll(r.Body)
		if err == nil {
			t.Error("expected MaxBytesReader to return an error for oversized payload")
		}
		w.WriteHeader(http.StatusRequestEntityTooLarge)
	})).ServeHTTP(rec, req)
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("status=%d", rec.Code)
	}
}

func TestRecoveryMiddleware(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	RecoveryMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		panic("kaboom")
	})).ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status=%d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Internal Server Error") {
		t.Errorf("body=%q", rec.Body.String())
	}
}

func TestRequestIDMiddleware(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	var got string
	RequestIDMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got, _ = r.Context().Value(ctxRequestID).(string)
	})).ServeHTTP(rec, req)
	if got == "" {
		t.Error("missing request id in context")
	}
	if rec.Header().Get("X-Request-Id") == "" {
		t.Error("missing request id in response")
	}
}

type fakeAsyncLogger struct {
	entries []entity.RequestLog
}

func (f *fakeAsyncLogger) Send(e entity.RequestLog) { f.entries = append(f.entries, e) }

func TestLoggingMiddlewareCapturesEntry(t *testing.T) {
	alog := &fakeAsyncLogger{}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/foo?x=1", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.1")
	req.RemoteAddr = "127.0.0.1:54321"

	mid := func(next http.Handler) http.Handler {
		return RequestIDMiddleware(LoggingMiddleware(alog)(next))
	}
	mid(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})).ServeHTTP(rec, req)

	if rec.Code != http.StatusTeapot {
		t.Fatalf("status=%d", rec.Code)
	}
	// give the goroutine a moment in case it ever becomes async
	time.Sleep(5 * time.Millisecond)
	if len(alog.entries) != 1 {
		t.Fatalf("entries=%d", len(alog.entries))
	}
	e := alog.entries[0]
	if e.Method != http.MethodGet || e.Path != "/foo" || e.StatusCode != http.StatusTeapot {
		t.Errorf("unexpected entry: %+v", e)
	}
	if e.ClientIP != "203.0.113.1" {
		t.Errorf("client ip=%q", e.ClientIP)
	}
	if e.Query == nil || *e.Query != "x=1" {
		t.Errorf("query=%v", e.Query)
	}
}
