package application

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"prodlich/internal/domain/entity"
	repo "prodlich/internal/domain/repository"

	"github.com/bytedance/sonic"
)

// ErrInvalidMiddleware is returned for middleware validation failures.
var ErrInvalidMiddleware = errors.New("invalid middleware")

// validMiddlewareKinds enumerates the kinds the runtime resolver knows how to
// instantiate.
var validMiddlewareKinds = map[string]struct{}{
	entity.MiddlewareKindCORS:              {},
	entity.MiddlewareKindHeaderRewrite:     {},
	entity.MiddlewareKindRateLimitOverride: {},
	entity.MiddlewareKindRequestID:         {},
	entity.MiddlewareKindStripPrefix:       {},
}

// Middleware is the use case for managing reusable middlewares and their
// route attachments.
type Middleware struct {
	repo repo.MiddlewareRepo
}

// NewMiddlewareUseCase wires a Middleware use case.
func NewMiddlewareUseCase(r repo.MiddlewareRepo) *Middleware {
	return &Middleware{repo: r}
}

// List returns every middleware.
func (m *Middleware) List(ctx context.Context) ([]entity.Middleware, error) {
	return m.repo.GetAll(ctx)
}

// Get returns a single middleware.
func (m *Middleware) Get(ctx context.Context, id int) (entity.Middleware, error) {
	return m.repo.GetByID(ctx, id)
}

// Create validates the JSON config and inserts the row.
func (m *Middleware) Create(ctx context.Context, mw entity.Middleware) (int, error) {
	if err := validateMiddleware(&mw); err != nil {
		return 0, err
	}
	return m.repo.Create(ctx, mw)
}

// Update validates the JSON config and persists changes.
func (m *Middleware) Update(ctx context.Context, mw entity.Middleware) error {
	if err := validateMiddleware(&mw); err != nil {
		return err
	}
	return m.repo.Update(ctx, mw)
}

// Delete removes a middleware.
func (m *Middleware) Delete(ctx context.Context, id int) error {
	return m.repo.Delete(ctx, id)
}

// Attach associates a middleware with a route at a given position.
func (m *Middleware) Attach(ctx context.Context, routeID, middlewareID, sortOrder int) error {
	if routeID <= 0 || middlewareID <= 0 {
		return fmt.Errorf("%w: route_id and middleware_id are required", ErrInvalidMiddleware)
	}
	if sortOrder < 0 {
		return fmt.Errorf("%w: sort_order must be >= 0", ErrInvalidMiddleware)
	}
	return m.repo.Attach(ctx, routeID, middlewareID, sortOrder)
}

// Detach removes the (route, middleware) link.
func (m *Middleware) Detach(ctx context.Context, routeID, middlewareID int) error {
	return m.repo.Detach(ctx, routeID, middlewareID)
}

// ListByRoute returns the middlewares attached to a route, in declared order.
func (m *Middleware) ListByRoute(ctx context.Context, routeID int) ([]entity.Middleware, error) {
	return m.repo.ListByRoute(ctx, routeID)
}

func validateMiddleware(mw *entity.Middleware) error {
	if strings.TrimSpace(mw.Name) == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidMiddleware)
	}
	if _, ok := validMiddlewareKinds[mw.Kind]; !ok {
		return fmt.Errorf("%w: unknown kind %q", ErrInvalidMiddleware, mw.Kind)
	}
	if len(mw.Config) == 0 && mw.ConfigJSON != nil {
		raw, err := sonic.Marshal(mw.ConfigJSON)
		if err != nil {
			return fmt.Errorf("%w: config not serialisable: %v", ErrInvalidMiddleware, err)
		}
		mw.Config = raw
	}
	if len(mw.Config) > 0 {
		var dummy any
		if err := sonic.Unmarshal(mw.Config, &dummy); err != nil {
			return fmt.Errorf("%w: config must be valid JSON", ErrInvalidMiddleware)
		}
	}
	return nil
}
