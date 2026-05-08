package public

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakeChecker struct{ err error }

func (f fakeChecker) HealthCheck(_ context.Context) error { return f.err }

func TestLiveHandler(t *testing.T) {
	rec := httptest.NewRecorder()
	LiveHandler(rec, httptest.NewRequest(http.MethodGet, "/health", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d", rec.Code)
	}
}

func TestReadyHandlerHealthy(t *testing.T) {
	rec := httptest.NewRecorder()
	ReadyHandler(fakeChecker{})(rec, httptest.NewRequest(http.MethodGet, "/ready", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d", rec.Code)
	}
}

func TestReadyHandlerUnavailable(t *testing.T) {
	rec := httptest.NewRecorder()
	ReadyHandler(fakeChecker{err: errors.New("db down")})(rec, httptest.NewRequest(http.MethodGet, "/ready", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status=%d", rec.Code)
	}
}

func TestReadyHandlerNilChecker(t *testing.T) {
	rec := httptest.NewRecorder()
	ReadyHandler(nil)(rec, httptest.NewRequest(http.MethodGet, "/ready", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d", rec.Code)
	}
}
