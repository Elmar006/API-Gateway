// Package application implements the business logic layer (use cases) of the API Gateway.
package application

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"

	"prodlich/internal/domain/entity"
	repo "prodlich/internal/domain/repository"
)

// ActiveRoutesObserver is notified whenever the active route cache is rebuilt.
// It is wired by main to update the Prometheus gauge and the trie router.
type ActiveRoutesObserver func(routes []entity.Route)

// Route is the use case for managing dynamic proxy routes and the in-memory
// active-routes cache used by the public proxy.
type Route struct {
	repo      repo.RouteRepo
	mu        sync.RWMutex
	active    []entity.Route
	autoLoad  bool
	observers []ActiveRoutesObserver
}

// NewRouterService constructs a Route use case. autoReload triggers a cache
// rebuild after every mutation.
func NewRouterService(repo repo.RouteRepo, autoReload bool) *Route {
	return &Route{repo: repo, autoLoad: autoReload}
}

// OnReload registers an observer that fires after every successful cache reload.
func (r *Route) OnReload(o ActiveRoutesObserver) {
	if o == nil {
		return
	}
	r.observers = append(r.observers, o)
}

// ReloadCache reloads the active-routes cache from the database and notifies
// any registered observers.
func (r *Route) ReloadCache(ctx context.Context) error {
	routes, err := r.repo.GetActiveRoutes(ctx)
	if err != nil {
		return fmt.Errorf("load active routes: %w", err)
	}
	r.mu.Lock()
	r.active = routes
	r.mu.Unlock()
	for _, o := range r.observers {
		o(routes)
	}
	return nil
}

// GetActiveRoutes returns a snapshot of the active routes cache.
func (r *Route) GetActiveRoutes() []entity.Route {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]entity.Route, len(r.active))
	copy(out, r.active)
	return out
}

// GetAllRoutes returns every route, regardless of active status.
func (r *Route) GetAllRoutes(ctx context.Context) ([]entity.Route, error) {
	return r.repo.GetAll(ctx)
}

// GetRouteByID fetches a single route by ID.
func (r *Route) GetRouteByID(ctx context.Context, id int) (*entity.Route, error) {
	route, err := r.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &route, nil
}

// Create validates and inserts a new route, then reloads the cache.
func (r *Route) Create(ctx context.Context, route entity.Route) (int, error) {
	if err := validateRoute(route); err != nil {
		return 0, err
	}
	route.Method = route.NormalizeMethod()
	id, err := r.repo.Create(ctx, route)
	if err != nil {
		return 0, err
	}
	if r.autoLoad {
		if err := r.ReloadCache(ctx); err != nil {
			return id, fmt.Errorf("reload cache: %w", err)
		}
	}
	return id, nil
}

// UpdateRoute validates and updates an existing route, then reloads the cache.
func (r *Route) UpdateRoute(ctx context.Context, route entity.Route) error {
	if err := validateRoute(route); err != nil {
		return err
	}
	route.Method = route.NormalizeMethod()
	if err := r.repo.Update(ctx, route); err != nil {
		return err
	}
	if r.autoLoad {
		return r.ReloadCache(ctx)
	}
	return nil
}

// DeleteRoute removes a route by ID and reloads the cache.
func (r *Route) DeleteRoute(ctx context.Context, id int) error {
	if err := r.repo.Delete(ctx, id); err != nil {
		return err
	}
	if r.autoLoad {
		return r.ReloadCache(ctx)
	}
	return nil
}

// ToggleRoute flips is_active on the given route and reloads the cache.
func (r *Route) ToggleRoute(ctx context.Context, id int) error {
	if err := r.repo.ToggleActive(ctx, id); err != nil {
		return err
	}
	if r.autoLoad {
		return r.ReloadCache(ctx)
	}
	return nil
}

// ErrInvalidRoute is returned by validateRoute when a payload is malformed.
var ErrInvalidRoute = errors.New("invalid route")

func validateRoute(r entity.Route) error {
	if strings.TrimSpace(r.PathPattern) == "" {
		return fmt.Errorf("%w: path_pattern is required", ErrInvalidRoute)
	}
	if !strings.HasPrefix(r.PathPattern, "/") {
		return fmt.Errorf("%w: path_pattern must start with '/'", ErrInvalidRoute)
	}
	// A route must either point to a single TargetURL or to a ClusterID. If
	// neither is provided, there's nowhere to forward to.
	hasTarget := strings.TrimSpace(r.TargetURL) != ""
	hasCluster := r.ClusterID != nil && *r.ClusterID > 0
	if !hasTarget && !hasCluster {
		return fmt.Errorf("%w: target_url or cluster_id is required", ErrInvalidRoute)
	}
	if hasTarget {
		u, err := url.Parse(r.TargetURL)
		if err != nil || u.Scheme == "" || u.Host == "" {
			return fmt.Errorf("%w: target_url must be an absolute URL", ErrInvalidRoute)
		}
	}
	if r.RateLimit < 0 {
		return fmt.Errorf("%w: rate_limit must be >= 0", ErrInvalidRoute)
	}
	return nil
}
