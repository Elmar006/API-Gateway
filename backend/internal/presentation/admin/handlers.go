// Package admin – HTTP handlers for admin operations.
package admin

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"prodlich/internal/application"
	"prodlich/internal/domain/entity"
	"prodlich/internal/domain/repository"

	"github.com/bytedance/sonic"
	"github.com/go-chi/chi/v5"
)

var jsonAPI = sonic.ConfigFastest

func responseError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = jsonAPI.NewEncoder(w).Encode(map[string]string{"error": message})
}

func responseSuccess(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = jsonAPI.NewEncoder(w).Encode(data)
}

func loginHandler(authUC *application.AdminAuthUseCase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := jsonAPI.NewDecoder(r.Body).Decode(&req); err != nil {
			responseError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		token, err := authUC.Login(r.Context(), req.Username, req.Password)
		if err != nil {
			responseError(w, http.StatusUnauthorized, err.Error())
			return
		}
		responseSuccess(w, http.StatusOK, map[string]string{"token": token})
	}
}

func getRoutesHandler(routeUC *application.Route) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		routes, err := routeUC.GetAllRoutes(r.Context())
		if err != nil {
			responseError(w, http.StatusInternalServerError, "failed to fetch routes")
			return
		}
		responseSuccess(w, http.StatusOK, routes)
	}
}

func createRouteHandler(routeUC *application.Route) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var route entity.Route
		if err := jsonAPI.NewDecoder(r.Body).Decode(&route); err != nil {
			responseError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		id, err := routeUC.Create(r.Context(), route)
		if err != nil {
			if errors.Is(err, application.ErrInvalidRoute) {
				responseError(w, http.StatusBadRequest, err.Error())
				return
			}
			responseError(w, http.StatusInternalServerError, "failed to create route")
			return
		}
		responseSuccess(w, http.StatusCreated, map[string]int{"id": id})
	}
}

func updateRouteHandler(routeUC *application.Route) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			responseError(w, http.StatusBadRequest, "invalid route id")
			return
		}
		var route entity.Route
		if err := jsonAPI.NewDecoder(r.Body).Decode(&route); err != nil {
			responseError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		route.ID = id
		if err := routeUC.UpdateRoute(r.Context(), route); err != nil {
			if errors.Is(err, application.ErrInvalidRoute) {
				responseError(w, http.StatusBadRequest, err.Error())
				return
			}
			if errors.Is(err, repository.ErrNotFound) {
				responseError(w, http.StatusNotFound, "route not found")
				return
			}
			responseError(w, http.StatusInternalServerError, "failed to update route")
			return
		}
		responseSuccess(w, http.StatusOK, map[string]string{"status": "updated"})
	}
}

func deleteRouteHandler(routeUC *application.Route) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			responseError(w, http.StatusBadRequest, "invalid route id")
			return
		}
		if err := routeUC.DeleteRoute(r.Context(), id); err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				responseError(w, http.StatusNotFound, "route not found")
				return
			}
			responseError(w, http.StatusInternalServerError, "failed to delete route")
			return
		}
		responseSuccess(w, http.StatusOK, map[string]string{"status": "deleted"})
	}
}

func toggleRouteHandler(routeUC *application.Route) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			responseError(w, http.StatusBadRequest, "invalid route id")
			return
		}
		if err := routeUC.ToggleRoute(r.Context(), id); err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				responseError(w, http.StatusNotFound, "route not found")
				return
			}
			responseError(w, http.StatusInternalServerError, "failed to toggle route")
			return
		}
		responseSuccess(w, http.StatusOK, map[string]string{"status": "toggled"})
	}
}

func getLogsHandler(logUC *application.LogUseCase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		filters := repository.LogFilters{
			Path:       query.Get("path"),
			StatusCode: 0,
		}
		if sc := query.Get("status"); sc != "" {
			if code, err := strconv.Atoi(sc); err == nil {
				filters.StatusCode = code
			}
		}
		if from := query.Get("from"); from != "" {
			if t, err := time.Parse(time.RFC3339, from); err == nil {
				filters.From = &t
			}
		}
		if to := query.Get("to"); to != "" {
			if t, err := time.Parse(time.RFC3339, to); err == nil {
				filters.To = &t
			}
		}
		limit, _ := strconv.Atoi(query.Get("limit"))
		if limit <= 0 {
			limit = 50
		}
		if limit > 1000 {
			limit = 1000
		}
		offset, _ := strconv.Atoi(query.Get("offset"))
		if offset < 0 {
			offset = 0
		}
		logs, total, err := logUC.GetLogs(r.Context(), filters, limit, offset)
		if err != nil {
			responseError(w, http.StatusInternalServerError, "failed to fetch logs")
			return
		}
		responseSuccess(w, http.StatusOK, map[string]any{
			"logs":   logs,
			"total":  total,
			"limit":  limit,
			"offset": offset,
		})
	}
}

func getMetricsHandler(metricUC *application.MetricUseCase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		period := r.URL.Query().Get("period")
		if period == "" {
			period = "day"
		}
		metrics, err := metricUC.GetMetrics(r.Context(), period)
		if err != nil {
			responseError(w, http.StatusInternalServerError, "failed to fetch metrics")
			return
		}
		responseSuccess(w, http.StatusOK, metrics)
	}
}

// healthHandler returns 200 unconditionally; mirrors the public /health.
func healthHandler(w http.ResponseWriter, _ *http.Request) {
	responseSuccess(w, http.StatusOK, map[string]string{"status": "ok"})
}
