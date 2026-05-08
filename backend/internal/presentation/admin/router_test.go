package admin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"prodlich/internal/application"
	"prodlich/internal/domain/entity"
	"prodlich/internal/domain/repository"
	authpkg "prodlich/internal/infrastructure/auth"

	"golang.org/x/crypto/bcrypt"
)

type stubAdmin struct {
	byName map[string]entity.AdminUser
	byID   map[int]entity.AdminUser
}

func (s *stubAdmin) GetByUsername(_ context.Context, u string) (entity.AdminUser, error) {
	if v, ok := s.byName[u]; ok {
		return v, nil
	}
	return entity.AdminUser{}, repository.ErrNotFound
}
func (s *stubAdmin) GetByID(_ context.Context, id int) (entity.AdminUser, error) {
	if v, ok := s.byID[id]; ok {
		return v, nil
	}
	return entity.AdminUser{}, repository.ErrNotFound
}
func (s *stubAdmin) CreateAdmin(_ context.Context, _ entity.AdminUser) (int, error) { return 0, nil }
func (s *stubAdmin) UpdateLastLogin(_ context.Context, _ int) error                  { return nil }
func (s *stubAdmin) GetAll(_ context.Context) ([]entity.AdminUser, error)            { return nil, nil }
func (s *stubAdmin) UpdateAdmin(_ context.Context, _ entity.AdminUser) error         { return nil }
func (s *stubAdmin) DeleteAdmin(_ context.Context, _ int) error                      { return nil }

type stubRoute struct {
	store map[int]entity.Route
}

func newStubRoute() *stubRoute { return &stubRoute{store: map[int]entity.Route{}} }

func (s *stubRoute) Create(_ context.Context, r entity.Route) (int, error) {
	r.ID = len(s.store) + 1
	s.store[r.ID] = r
	return r.ID, nil
}
func (s *stubRoute) GetByID(_ context.Context, id int) (entity.Route, error) {
	if r, ok := s.store[id]; ok {
		return r, nil
	}
	return entity.Route{}, repository.ErrNotFound
}
func (s *stubRoute) GetAll(_ context.Context) ([]entity.Route, error) {
	out := make([]entity.Route, 0, len(s.store))
	for _, r := range s.store {
		out = append(out, r)
	}
	return out, nil
}
func (s *stubRoute) Update(_ context.Context, r entity.Route) error {
	if _, ok := s.store[r.ID]; !ok {
		return repository.ErrNotFound
	}
	s.store[r.ID] = r
	return nil
}
func (s *stubRoute) Delete(_ context.Context, id int) error {
	if _, ok := s.store[id]; !ok {
		return repository.ErrNotFound
	}
	delete(s.store, id)
	return nil
}
func (s *stubRoute) ToggleActive(_ context.Context, id int) error {
	r, ok := s.store[id]
	if !ok {
		return repository.ErrNotFound
	}
	r.IsActive = !r.IsActive
	s.store[id] = r
	return nil
}
func (s *stubRoute) GetActiveRoutes(_ context.Context) ([]entity.Route, error) {
	out := []entity.Route{}
	for _, r := range s.store {
		if r.IsActive {
			out = append(out, r)
		}
	}
	return out, nil
}

func setupRouter(t *testing.T) (http.Handler, string) {
	t.Helper()
	hash, _ := bcrypt.GenerateFromPassword([]byte("admin-pass"), bcrypt.MinCost)
	adminRepo := &stubAdmin{
		byName: map[string]entity.AdminUser{"admin": {ID: 1, Username: "admin", PasswordHash: string(hash), Role: "admin"}},
		byID:   map[int]entity.AdminUser{1: {ID: 1, Username: "admin", PasswordHash: string(hash), Role: "admin"}},
	}
	tokSvc := authpkg.NewJWTService("router-test-secret-1234567890")
	authUC := application.NewAdminAuthUseCase(adminRepo, tokSvc)
	routeUC := application.NewRouterService(newStubRoute(), false)
	logUC := application.NewLogUseCase(&fakeLogRepoLite{})
	metricUC := application.NewMetricUseCase(&fakeLogRepoLite{})

	router := NewRouter(RouterConfig{
		RouteUC:  routeUC,
		AuthUC:   authUC,
		LogUC:    logUC,
		MetricUC: metricUC,
	})

	tok, err := authUC.Login(context.Background(), "admin", "admin-pass")
	if err != nil {
		t.Fatalf("seed login: %v", err)
	}
	return router, tok
}

type fakeLogRepoLite struct{}

func (fakeLogRepoLite) Insert(context.Context, entity.RequestLog) error { return nil }
func (fakeLogRepoLite) GetWithFilters(context.Context, repository.LogFilters, int, int) ([]entity.RequestLog, int, error) {
	return nil, 0, nil
}
func (fakeLogRepoLite) GetMetrics(context.Context, time.Time, time.Time) (entity.Metrics, error) {
	return entity.Metrics{}, nil
}

func TestAdminRouterHealthOpen(t *testing.T) {
	router, _ := setupRouter(t)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAdminRouterRequiresJWT(t *testing.T) {
	router, _ := setupRouter(t)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/routes", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d", rec.Code)
	}
}

func TestAdminRouterAcceptsValidJWT(t *testing.T) {
	router, tok := setupRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/routes", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAdminLoginEndpoint(t *testing.T) {
	router, _ := setupRouter(t)
	body := strings.NewReader(`{"username":"admin","password":"admin-pass"}`)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/login", body))
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"token"`) {
		t.Errorf("missing token: %s", rec.Body.String())
	}
}

func TestAdminLoginRejectsInvalid(t *testing.T) {
	router, _ := setupRouter(t)
	body := strings.NewReader(`{"username":"admin","password":"wrong"}`)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/login", body))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d", rec.Code)
	}
}

func TestAdminCreateAndDeleteRoute(t *testing.T) {
	router, tok := setupRouter(t)

	create := strings.NewReader(`{"method":"GET","path_pattern":"/x","target_url":"http://up:8080","is_active":true}`)
	req := httptest.NewRequest(http.MethodPost, "/routes", create)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", rec.Code, rec.Body.String())
	}

	// invalid payload (missing path)
	bad := strings.NewReader(`{"target_url":"http://up:8080"}`)
	req = httptest.NewRequest(http.MethodPost, "/routes", bad)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("invalid create status=%d", rec.Code)
	}

	// toggle id=1
	req = httptest.NewRequest(http.MethodPatch, "/routes/1/toggle", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("toggle status=%d", rec.Code)
	}

	// update id=1
	upd := strings.NewReader(`{"method":"GET","path_pattern":"/y","target_url":"http://up:8080","is_active":true}`)
	req = httptest.NewRequest(http.MethodPut, "/routes/1", upd)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("update status=%d body=%s", rec.Code, rec.Body.String())
	}

	// delete id=1
	req = httptest.NewRequest(http.MethodDelete, "/routes/1", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("delete status=%d", rec.Code)
	}

	// delete again => 404
	req = httptest.NewRequest(http.MethodDelete, "/routes/1", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestAdminLogsAndMetrics(t *testing.T) {
	router, tok := setupRouter(t)
	for _, p := range []string{"/logs", "/metrics-summary"} {
		req := httptest.NewRequest(http.MethodGet, p, nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s status=%d body=%s", p, rec.Code, rec.Body.String())
		}
	}
}

func TestAdminInvalidIDRoute(t *testing.T) {
	router, tok := setupRouter(t)
	for _, p := range []string{"/routes/abc", "/routes/abc/toggle"} {
		method := http.MethodDelete
		if strings.HasSuffix(p, "/toggle") {
			method = http.MethodPatch
		}
		req := httptest.NewRequest(method, p, nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("%s status=%d", p, rec.Code)
		}
	}
}
