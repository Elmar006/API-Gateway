package repository

import (
	"context"

	"prodlich/internal/domain/entity"
)

// MiddlewareRepo is the persistence contract for middlewares and their
// route attachments.
type MiddlewareRepo interface {
	Create(ctx context.Context, mw entity.Middleware) (int, error)
	GetByID(ctx context.Context, id int) (entity.Middleware, error)
	GetAll(ctx context.Context) ([]entity.Middleware, error)
	Update(ctx context.Context, mw entity.Middleware) error
	Delete(ctx context.Context, id int) error

	// Attach links middleware to a route at a given sort order; if the link
	// already exists, sort_order is updated.
	Attach(ctx context.Context, routeID, middlewareID, sortOrder int) error
	Detach(ctx context.Context, routeID, middlewareID int) error
	ListByRoute(ctx context.Context, routeID int) ([]entity.Middleware, error)
}
