// Package public contains handlers and middleware for the public proxy.
package public

import (
	"context"
	"net/http"
	"strings"
	"sync/atomic"

	"prodlich/internal/application"
	"prodlich/internal/domain/entity"
	"prodlich/internal/infrastructure/limiter"
	"prodlich/internal/infrastructure/proxy"
	"prodlich/internal/infrastructure/router"

	"github.com/go-chi/chi/v5"
)

// RouteSnapshotter exposes the active routes cache.
type RouteSnapshotter interface {
	GetActiveRoutes() []entity.Route
	OnReload(application.ActiveRoutesObserver)
}

// AuthValidator – интерфейс для проверки JWT токенов администраторов
type AuthValidator interface {
	ValidateToken(ctx context.Context, tokenString string) (*entity.AdminUser, error)
}

// HealthChecker is implemented by anything that can verify backing services.
type HealthChecker interface {
	HealthCheck(ctx context.Context) error
}

// ProxyDeps bundles the dependencies needed to build the public handler.
type ProxyDeps struct {
	Routes          RouteSnapshotter
	IPLimiter       *limiter.IPLimiter
	AdminAuth       AuthValidator
	UserAuth        *application.AuthUseCase
	AsyncLog        asyncLogger
	DBHealth        HealthChecker
	MaxBodyBytes    int64
	RateLimitHits   rateHitsCounter
}

// Handler builds the public proxy handler with all middleware applied.
//
// The handler routes special endpoints (/health, /live, /ready, /auth/*) and
// falls through to the dynamic proxy for everything else.
func Handler(deps ProxyDeps) http.Handler {
	r := chi.NewRouter()

	r.Use(RequestIDMiddleware)
	r.Use(TracingMiddleware)
	r.Use(SecurityHeadersMiddleware)
	r.Use(LoggingMiddleware(deps.AsyncLog))
	r.Use(RecoveryMiddleware)
	if deps.MaxBodyBytes > 0 {
		r.Use(MaxBodyBytesMiddleware(deps.MaxBodyBytes))
	}

	// Health endpoints – always available, never proxied.
	r.Get("/health", LiveHandler)
	r.Get("/live", LiveHandler)
	r.Get("/ready", ReadyHandler(deps.DBHealth))

	// User auth – mounted only when configured.
	if deps.UserAuth != nil {
		r.Route("/auth", func(sub chi.Router) {
			sub.Post("/register", UserRegisterHandler(deps.UserAuth))
			sub.Post("/login", UserLoginHandler(deps.UserAuth))
		})
	}

	matcher := newDynamicMatcher(deps.Routes)
	dyn := dynamicProxy(matcher, deps.IPLimiter, deps.AdminAuth, deps.RateLimitHits)
	r.HandleFunc("/*", dyn)
	return r
}

// dynamicMatcher wraps a trie router that gets rebuilt whenever the active
// routes cache is reloaded.
type dynamicMatcher struct {
	tr atomic.Pointer[router.Router]
}

func newDynamicMatcher(snap RouteSnapshotter) *dynamicMatcher {
	dm := &dynamicMatcher{}
	dm.rebuild(snap.GetActiveRoutes())
	snap.OnReload(dm.rebuild)
	return dm
}

func (dm *dynamicMatcher) rebuild(routes []entity.Route) {
	tr := router.New()
	for i := range routes {
		rt := routes[i]
		tr.Add(rt.NormalizeMethod(), rt.PathPattern, rt.Priority, &rt)
	}
	dm.tr.Store(tr)
}

func (dm *dynamicMatcher) lookup(method, path string) (*entity.Route, map[string]string, bool) {
	tr := dm.tr.Load()
	if tr == nil {
		return nil, nil, false
	}
	m, ok := tr.Lookup(method, path)
	if !ok {
		return nil, nil, false
	}
	rt, _ := m.Value.(*entity.Route)
	if rt == nil {
		return nil, nil, false
	}
	return rt, m.Params, true
}

func dynamicProxy(matcher *dynamicMatcher, lim *limiter.IPLimiter, auth AuthValidator, hits rateHitsCounter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		route, params, ok := matcher.lookup(r.Method, r.URL.Path)
		if !ok {
			http.NotFound(w, r)
			return
		}

		if route.RequireAuth {
			if !validateBearer(r, auth) {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
		}

		if route.RateLimit > 0 {
			ip := clientIP(r)
			if !lim.Allow(ip, route.RateLimit) {
				if hits != nil {
					hits.Inc()
				}
				http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
				return
			}
		}

		// Forward path params to the upstream as headers.
		for k, v := range params {
			r.Header.Set("X-Path-Param-"+http.CanonicalHeaderKey(k), v)
		}

		builder, err := proxy.NewReverseProxyBuilder(route.TargetURL, proxy.DefaultConfig())
		if err != nil {
			http.Error(w, "bad gateway", http.StatusBadGateway)
			return
		}

		// Propagate the matched route id for the logger middleware upstream.
		setRouteID(r.Context(), route.ID)
		builder.Build().ServeHTTP(w, r)
	}
}

func validateBearer(r *http.Request, auth AuthValidator) bool {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return false
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return false
	}
	if _, err := auth.ValidateToken(r.Context(), parts[1]); err != nil {
		return false
	}
	return true
}
