// Package public – built-in user authentication endpoints (/auth/register, /auth/login).
package public

import (
	"errors"
	"net/http"

	"prodlich/internal/application"

	"github.com/bytedance/sonic"
)

var ajson = sonic.ConfigFastest

type registerRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// UserRegisterHandler signs up a new user and returns a JWT.
func UserRegisterHandler(uc *application.AuthUseCase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req registerRequest
		if err := ajson.NewDecoder(r.Body).Decode(&req); err != nil {
			writeAuthError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		token, err := uc.Register(r.Context(), req.Name, req.Email, req.Password)
		if err != nil {
			writeAuthError(w, mapAuthErr(err), err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, map[string]string{"token": token})
	}
}

// UserLoginHandler authenticates an existing user and returns a JWT.
func UserLoginHandler(uc *application.AuthUseCase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req loginRequest
		if err := ajson.NewDecoder(r.Body).Decode(&req); err != nil {
			writeAuthError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		token, err := uc.Login(r.Context(), req.Email, req.Password)
		if err != nil {
			writeAuthError(w, mapAuthErr(err), err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"token": token})
	}
}

func writeAuthError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func mapAuthErr(err error) int {
	switch {
	case errors.Is(err, application.ErrInvalidCredentials):
		return http.StatusUnauthorized
	case errors.Is(err, application.ErrEmailExists):
		return http.StatusConflict
	case errors.Is(err, application.ErrInvalidEmail),
		errors.Is(err, application.ErrInvalidName),
		errors.Is(err, application.ErrWeakPassword):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
