package admin

import (
	"errors"
	"net/http"
	"strconv"

	"prodlich/internal/application"
	"prodlich/internal/domain/entity"
	"prodlich/internal/domain/repository"

	"github.com/go-chi/chi/v5"
)

// middlewareDTO is the wire shape clients send: Config arrives as a JSON
// object instead of a base64-encoded byte slice.
type middlewareDTO struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Kind        string `json:"kind"`
	Config      any    `json:"config"`
	IsActive    bool   `json:"is_active"`
	Description string `json:"description"`
}

func (d middlewareDTO) toEntity() entity.Middleware {
	return entity.Middleware{
		ID:          d.ID,
		Name:        d.Name,
		Kind:        d.Kind,
		ConfigJSON:  d.Config,
		IsActive:    d.IsActive,
		Description: d.Description,
	}
}

func listMiddlewaresHandler(uc *application.Middleware) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mws, err := uc.List(r.Context())
		if err != nil {
			responseError(w, http.StatusInternalServerError, "failed to list middlewares")
			return
		}
		responseSuccess(w, http.StatusOK, mws)
	}
}

func getMiddlewareHandler(uc *application.Middleware) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			responseError(w, http.StatusBadRequest, "invalid middleware id")
			return
		}
		mw, err := uc.Get(r.Context(), id)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				responseError(w, http.StatusNotFound, "middleware not found")
				return
			}
			responseError(w, http.StatusInternalServerError, "failed to fetch middleware")
			return
		}
		responseSuccess(w, http.StatusOK, mw)
	}
}

func createMiddlewareHandler(uc *application.Middleware) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var dto middlewareDTO
		if err := jsonAPI.NewDecoder(r.Body).Decode(&dto); err != nil {
			responseError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		id, err := uc.Create(r.Context(), dto.toEntity())
		if err != nil {
			if errors.Is(err, application.ErrInvalidMiddleware) {
				responseError(w, http.StatusBadRequest, err.Error())
				return
			}
			responseError(w, http.StatusInternalServerError, "failed to create middleware")
			return
		}
		responseSuccess(w, http.StatusCreated, map[string]int{"id": id})
	}
}

func updateMiddlewareHandler(uc *application.Middleware) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			responseError(w, http.StatusBadRequest, "invalid middleware id")
			return
		}
		var dto middlewareDTO
		if err := jsonAPI.NewDecoder(r.Body).Decode(&dto); err != nil {
			responseError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		dto.ID = id
		if err := uc.Update(r.Context(), dto.toEntity()); err != nil {
			if errors.Is(err, application.ErrInvalidMiddleware) {
				responseError(w, http.StatusBadRequest, err.Error())
				return
			}
			if errors.Is(err, repository.ErrNotFound) {
				responseError(w, http.StatusNotFound, "middleware not found")
				return
			}
			responseError(w, http.StatusInternalServerError, "failed to update middleware")
			return
		}
		responseSuccess(w, http.StatusOK, map[string]string{"status": "updated"})
	}
}

func deleteMiddlewareHandler(uc *application.Middleware) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			responseError(w, http.StatusBadRequest, "invalid middleware id")
			return
		}
		if err := uc.Delete(r.Context(), id); err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				responseError(w, http.StatusNotFound, "middleware not found")
				return
			}
			responseError(w, http.StatusInternalServerError, "failed to delete middleware")
			return
		}
		responseSuccess(w, http.StatusOK, map[string]string{"status": "deleted"})
	}
}

func attachMiddlewareHandler(uc *application.Middleware) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		routeID, err := strconv.Atoi(chi.URLParam(r, "rid"))
		if err != nil {
			responseError(w, http.StatusBadRequest, "invalid route id")
			return
		}
		var body struct {
			MiddlewareID int `json:"middleware_id"`
			SortOrder    int `json:"sort_order"`
		}
		if err := jsonAPI.NewDecoder(r.Body).Decode(&body); err != nil {
			responseError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if err := uc.Attach(r.Context(), routeID, body.MiddlewareID, body.SortOrder); err != nil {
			if errors.Is(err, application.ErrInvalidMiddleware) {
				responseError(w, http.StatusBadRequest, err.Error())
				return
			}
			responseError(w, http.StatusInternalServerError, "failed to attach middleware")
			return
		}
		responseSuccess(w, http.StatusOK, map[string]string{"status": "attached"})
	}
}

func detachMiddlewareHandler(uc *application.Middleware) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		routeID, err := strconv.Atoi(chi.URLParam(r, "rid"))
		if err != nil {
			responseError(w, http.StatusBadRequest, "invalid route id")
			return
		}
		mwID, err := strconv.Atoi(chi.URLParam(r, "mid"))
		if err != nil {
			responseError(w, http.StatusBadRequest, "invalid middleware id")
			return
		}
		if err := uc.Detach(r.Context(), routeID, mwID); err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				responseError(w, http.StatusNotFound, "attachment not found")
				return
			}
			responseError(w, http.StatusInternalServerError, "failed to detach middleware")
			return
		}
		responseSuccess(w, http.StatusOK, map[string]string{"status": "detached"})
	}
}

func listRouteMiddlewaresHandler(uc *application.Middleware) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		routeID, err := strconv.Atoi(chi.URLParam(r, "rid"))
		if err != nil {
			responseError(w, http.StatusBadRequest, "invalid route id")
			return
		}
		mws, err := uc.ListByRoute(r.Context(), routeID)
		if err != nil {
			responseError(w, http.StatusInternalServerError, "failed to list route middlewares")
			return
		}
		responseSuccess(w, http.StatusOK, mws)
	}
}
