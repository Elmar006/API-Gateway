// Package admin provides routing for the Admin API endpoints.
package admin

import (
	"net/http"

	"prodlich/internal/application"
	"prodlich/internal/infrastructure/metrics"

	"github.com/go-chi/chi/v5"
)

// RouterConfig describes the dependencies required by NewRouter.
type RouterConfig struct {
	RouteUC       *application.Route
	AuthUC        *application.AdminAuthUseCase
	LogUC         *application.LogUseCase
	MetricUC      *application.MetricUseCase
	EnvironmentUC *application.Environment
	ClusterUC     *application.Cluster
	MiddlewareUC  *application.Middleware
	UserUC        *application.User
	TokenUC       *application.APIToken
	Metrics       *metrics.Registry
	ReadyFn       http.HandlerFunc
}

// NewRouter assembles the admin chi.Mux with middleware in this order:
// CORS → Metrics → Logging → Recovery → JWT (with a small allow-list).
func NewRouter(cfg RouterConfig) *chi.Mux {
	r := chi.NewRouter()
	r.Use(CORSMiddleware)
	if cfg.Metrics != nil {
		r.Use(cfg.Metrics.HTTPMiddleware("admin"))
	}
	r.Use(LoggingMiddleware)
	r.Use(RecoveryMiddleware)
	r.Use(AuthMiddleware(AuthConfig{
		Auth:      cfg.AuthUC,
		APIToken:  cfg.TokenUC,
		Admins:    cfg.UserUC,
		OpenPaths: []string{"/login", "/health", "/live", "/ready"},
	}))

	// Public health endpoints – open by virtue of the JWT allow-list above.
	r.Get("/health", healthHandler)
	r.Get("/live", healthHandler)
	if cfg.ReadyFn != nil {
		r.Get("/ready", cfg.ReadyFn)
	}

	r.Post("/login", loginHandler(cfg.AuthUC))

	// Routes (existing).
	r.Get("/routes", getRoutesHandler(cfg.RouteUC))
	r.Post("/routes", createRouteHandler(cfg.RouteUC))
	r.Put("/routes/{id}", updateRouteHandler(cfg.RouteUC))
	r.Delete("/routes/{id}", deleteRouteHandler(cfg.RouteUC))
	r.Patch("/routes/{id}/toggle", toggleRouteHandler(cfg.RouteUC))

	r.Get("/logs", getLogsHandler(cfg.LogUC))
	r.Get("/metrics-summary", getMetricsHandler(cfg.MetricUC))

	// Environments.
	if cfg.EnvironmentUC != nil {
		r.Get("/environments", listEnvironmentsHandler(cfg.EnvironmentUC))
		r.Post("/environments", createEnvironmentHandler(cfg.EnvironmentUC))
		r.Get("/environments/{id}", getEnvironmentHandler(cfg.EnvironmentUC))
		r.Put("/environments/{id}", updateEnvironmentHandler(cfg.EnvironmentUC))
		r.Delete("/environments/{id}", deleteEnvironmentHandler(cfg.EnvironmentUC))
	}

	// Clusters + targets.
	if cfg.ClusterUC != nil {
		r.Get("/clusters", listClustersHandler(cfg.ClusterUC))
		r.Post("/clusters", createClusterHandler(cfg.ClusterUC))
		r.Get("/clusters/{id}", getClusterHandler(cfg.ClusterUC))
		r.Put("/clusters/{id}", updateClusterHandler(cfg.ClusterUC))
		r.Delete("/clusters/{id}", deleteClusterHandler(cfg.ClusterUC))
		r.Post("/clusters/{id}/targets", addClusterTargetHandler(cfg.ClusterUC))
		r.Put("/clusters/{id}/targets/{tid}", updateClusterTargetHandler(cfg.ClusterUC))
		r.Delete("/clusters/{id}/targets/{tid}", deleteClusterTargetHandler(cfg.ClusterUC))
	}

	// Middlewares + per-route attachments.
	if cfg.MiddlewareUC != nil {
		r.Get("/middlewares", listMiddlewaresHandler(cfg.MiddlewareUC))
		r.Post("/middlewares", createMiddlewareHandler(cfg.MiddlewareUC))
		r.Get("/middlewares/{id}", getMiddlewareHandler(cfg.MiddlewareUC))
		r.Put("/middlewares/{id}", updateMiddlewareHandler(cfg.MiddlewareUC))
		r.Delete("/middlewares/{id}", deleteMiddlewareHandler(cfg.MiddlewareUC))
		r.Get("/routes/{rid}/middlewares", listRouteMiddlewaresHandler(cfg.MiddlewareUC))
		r.Post("/routes/{rid}/middlewares", attachMiddlewareHandler(cfg.MiddlewareUC))
		r.Delete("/routes/{rid}/middlewares/{mid}", detachMiddlewareHandler(cfg.MiddlewareUC))
	}

	// Users (admin RBAC).
	if cfg.UserUC != nil {
		r.Get("/users", listUsersHandler(cfg.UserUC))
		r.Post("/users", createUserHandler(cfg.UserUC))
		r.Get("/users/{id}", getUserHandler(cfg.UserUC))
		r.Put("/users/{id}", updateUserHandler(cfg.UserUC))
		r.Delete("/users/{id}", deleteUserHandler(cfg.UserUC))
	}

	// API tokens (per-user).
	if cfg.TokenUC != nil {
		r.Get("/users/{uid}/tokens", listTokensHandler(cfg.TokenUC))
		r.Post("/users/{uid}/tokens", createTokenHandler(cfg.TokenUC))
		r.Delete("/tokens/{id}", revokeTokenHandler(cfg.TokenUC))
	}

	if cfg.Metrics != nil {
		r.Method(http.MethodGet, "/metrics", cfg.Metrics.Handler())
	}
	return r
}
