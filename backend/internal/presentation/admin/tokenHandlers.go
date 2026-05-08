package admin

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"prodlich/internal/application"
	"prodlich/internal/domain/repository"

	"github.com/go-chi/chi/v5"
)

func listTokensHandler(uc *application.APIToken) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := strconv.Atoi(chi.URLParam(r, "uid"))
		if err != nil {
			responseError(w, http.StatusBadRequest, "invalid user id")
			return
		}
		tokens, err := uc.List(r.Context(), userID)
		if err != nil {
			responseError(w, http.StatusInternalServerError, "failed to list tokens")
			return
		}
		responseSuccess(w, http.StatusOK, tokens)
	}
}

type createTokenRequest struct {
	Name      string     `json:"name"`
	Scopes    []string   `json:"scopes"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

func createTokenHandler(uc *application.APIToken) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := strconv.Atoi(chi.URLParam(r, "uid"))
		if err != nil {
			responseError(w, http.StatusBadRequest, "invalid user id")
			return
		}
		var req createTokenRequest
		if err := jsonAPI.NewDecoder(r.Body).Decode(&req); err != nil {
			responseError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		created, err := uc.Create(r.Context(), userID, req.Name, req.Scopes, req.ExpiresAt)
		if err != nil {
			if errors.Is(err, application.ErrInvalidToken) {
				responseError(w, http.StatusBadRequest, err.Error())
				return
			}
			responseError(w, http.StatusInternalServerError, "failed to create token")
			return
		}
		responseSuccess(w, http.StatusCreated, map[string]any{
			"id":     created.Token.ID,
			"prefix": created.Token.Prefix,
			"scopes": created.Token.Scopes,
			"secret": created.Secret, // shown once
		})
	}
}

func revokeTokenHandler(uc *application.APIToken) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			responseError(w, http.StatusBadRequest, "invalid token id")
			return
		}
		if err := uc.Revoke(r.Context(), id); err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				responseError(w, http.StatusNotFound, "token not found")
				return
			}
			responseError(w, http.StatusInternalServerError, "failed to revoke token")
			return
		}
		responseSuccess(w, http.StatusOK, map[string]string{"status": "revoked"})
	}
}
