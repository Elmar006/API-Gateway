package public

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"prodlich/internal/application"
	"prodlich/internal/domain/entity"
	"prodlich/internal/domain/repository"
	"prodlich/internal/infrastructure/limiter"
)

type stubRoutes struct {
	routes []entity.Route
	subs   []application.ActiveRoutesObserver
}

func (s *stubRoutes) GetActiveRoutes() []entity.Route { return s.routes }
func (s *stubRoutes) OnReload(o application.ActiveRoutesObserver) {
	s.subs = append(s.subs, o)
	o(s.routes)
}

type stubChecker struct{}

func (stubChecker) HealthCheck(_ context.Context) error { return nil }

type stubAuth struct{ ok bool }

func (s stubAuth) ValidateToken(_ context.Context, _ string) (*entity.AdminUser, error) {
	if !s.ok {
		return nil, errInvalid
	}
	return &entity.AdminUser{ID: 1}, nil
}

var errInvalid = stubErr("invalid")

type stubErr string

func (s stubErr) Error() string { return string(s) }

type fakeUserRepoP struct{}

func (fakeUserRepoP) Create(context.Context, entity.User) (int64, error) { return 0, nil }
func (fakeUserRepoP) GetByEmail(context.Context, string) (*entity.User, error) {
	return nil, repository.ErrNotFound
}
func (fakeUserRepoP) GetByID(context.Context, int64) (*entity.User, error) {
	return nil, repository.ErrNotFound
}

type stubAlog struct{ entries []entity.RequestLog }

func (s *stubAlog) Send(e entity.RequestLog) { s.entries = append(s.entries, e) }

type stubHits struct{ count int }

func (s *stubHits) Inc() { s.count++ }

func TestPublicHandlerProxiesAndLogs(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Param-User", r.Header.Get("X-Path-Param-User-Id"))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("hi"))
	}))
	defer upstream.Close()

	routes := &stubRoutes{routes: []entity.Route{
		{ID: 1, Method: "GET", PathPattern: "/api/users/{user-id}", TargetURL: upstream.URL, IsActive: true},
	}}
	lim := limiter.NewIPLimiter(time.Second)
	defer lim.Stop()
	alog := &stubAlog{}
	deps := ProxyDeps{
		Routes:       routes,
		IPLimiter:    lim,
		AdminAuth:    stubAuth{ok: true},
		AsyncLog:     alog,
		DBHealth:     stubChecker{},
		MaxBodyBytes: 0,
	}
	h := Handler(deps)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/users/42", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d", rec.Code)
	}
	if got := rec.Header().Get("X-Param-User"); got != "42" {
		t.Errorf("missing path param header: %q", got)
	}
	time.Sleep(5 * time.Millisecond)
	if len(alog.entries) != 1 {
		t.Fatalf("alog entries=%d", len(alog.entries))
	}
	if alog.entries[0].RouteID == nil || *alog.entries[0].RouteID != 1 {
		t.Errorf("route id missing in log entry: %+v", alog.entries[0])
	}
}

func TestPublicHandlerNotFound(t *testing.T) {
	routes := &stubRoutes{}
	lim := limiter.NewIPLimiter(time.Second)
	defer lim.Stop()
	h := Handler(ProxyDeps{Routes: routes, IPLimiter: lim, AsyncLog: &stubAlog{}})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/missing", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status=%d", rec.Code)
	}
}

func TestPublicHandlerRateLimit(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()
	routes := &stubRoutes{routes: []entity.Route{
		{ID: 1, Method: "GET", PathPattern: "/x", TargetURL: upstream.URL, IsActive: true, RateLimit: 1},
	}}
	lim := limiter.NewIPLimiter(time.Second)
	defer lim.Stop()
	hits := &stubHits{}
	h := Handler(ProxyDeps{Routes: routes, IPLimiter: lim, AsyncLog: &stubAlog{}, RateLimitHits: hits})

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/x", nil)
		req.RemoteAddr = "1.2.3.4:9000"
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
	}
	if hits.count == 0 {
		t.Fatalf("expected limiter rejections, got %d", hits.count)
	}
}

func TestPublicHandlerRequiresAuthOnProtectedRoute(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()
	routes := &stubRoutes{routes: []entity.Route{
		{ID: 1, Method: "GET", PathPattern: "/secure", TargetURL: upstream.URL, IsActive: true, RequireAuth: true},
	}}
	lim := limiter.NewIPLimiter(time.Second)
	defer lim.Stop()
	h := Handler(ProxyDeps{Routes: routes, IPLimiter: lim, AdminAuth: stubAuth{ok: false}, AsyncLog: &stubAlog{}})

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/secure", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d", rec.Code)
	}
}

func TestPublicHandlerHealthAndAuth(t *testing.T) {
	routes := &stubRoutes{}
	lim := limiter.NewIPLimiter(time.Second)
	defer lim.Stop()
	uc := application.NewAuthUseCase(fakeUserRepoP{}, "this-is-a-secret-test-key-1234")
	h := Handler(ProxyDeps{Routes: routes, IPLimiter: lim, AsyncLog: &stubAlog{}, UserAuth: uc, DBHealth: stubChecker{}})

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("health status=%d", rec.Code)
	}
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ready", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("ready status=%d", rec.Code)
	}
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/live", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("live status=%d", rec.Code)
	}
}
