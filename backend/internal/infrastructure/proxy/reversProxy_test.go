package proxy

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestReverseProxyForwarding(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Backend", "yes")
		w.WriteHeader(http.StatusTeapot)
		_, _ = io.WriteString(w, "hello "+r.URL.Path)
	}))
	defer upstream.Close()

	builder, err := NewReverseProxyBuilder(upstream.URL, DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(builder.Build())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/foo")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusTeapot {
		t.Errorf("status=%d", resp.StatusCode)
	}
	if resp.Header.Get("X-Backend") != "yes" {
		t.Errorf("missing backend header")
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "hello /foo") {
		t.Errorf("body=%q", string(body))
	}
}

func TestReverseProxyBadGateway(t *testing.T) {
	builder, err := NewReverseProxyBuilder("http://127.0.0.1:1", DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	builder.Build().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/x", nil))
	if rec.Code != http.StatusBadGateway {
		t.Errorf("status=%d", rec.Code)
	}
}

func TestNewReverseProxyBuilderInvalidURL(t *testing.T) {
	if _, err := NewReverseProxyBuilder("://bad", DefaultConfig()); err == nil {
		t.Fatal("expected error")
	}
}
