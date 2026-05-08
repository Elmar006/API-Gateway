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

func listUsersHandler(uc *application.User) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		users, err := uc.List(r.Context())
		if err != nil {
			responseError(w, http.StatusInternalServerError, "failed to list users")
			return
		}
		responseSuccess(w, http.StatusOK, users)
	}
}

func getUserHandler(uc *application.User) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			responseError(w, http.StatusBadRequest, "invalid user id")
			return
		}
		user, err := uc.Get(r.Context(), id)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				responseError(w, http.StatusNotFound, "user not found")
				return
			}
			responseError(w, http.StatusInternalServerError, "failed to fetch user")
			return
		}
		responseSuccess(w, http.StatusOK, user)
	}
}

type userDTO struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	Password string `json:"password,omitempty"`
}

func createUserHandler(uc *application.User) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var dto userDTO
		if err := jsonAPI.NewDecoder(r.Body).Decode(&dto); err != nil {
			responseError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		id, err := uc.Create(r.Context(), entity.AdminUser{Username: dto.Username, Role: dto.Role}, dto.Password)
		if err != nil {
			if errors.Is(err, application.ErrInvalidUser) {
				responseError(w, http.StatusBadRequest, err.Error())
				return
			}
			responseError(w, http.StatusInternalServerError, "failed to create user")
			return
		}
		responseSuccess(w, http.StatusCreated, map[string]int{"id": id})
	}
}

func updateUserHandler(uc *application.User) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			responseError(w, http.StatusBadRequest, "invalid user id")
			return
		}
		var dto userDTO
		if err := jsonAPI.NewDecoder(r.Body).Decode(&dto); err != nil {
			responseError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if err := uc.Update(r.Context(), entity.AdminUser{ID: id, Username: dto.Username, Role: dto.Role}, dto.Password); err != nil {
			if errors.Is(err, application.ErrInvalidUser) {
				responseError(w, http.StatusBadRequest, err.Error())
				return
			}
			if errors.Is(err, repository.ErrNotFound) {
				responseError(w, http.StatusNotFound, "user not found")
				return
			}
			responseError(w, http.StatusInternalServerError, "failed to update user")
			return
		}
		responseSuccess(w, http.StatusOK, map[string]string{"status": "updated"})
	}
}

func deleteUserHandler(uc *application.User) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			responseError(w, http.StatusBadRequest, "invalid user id")
			return
		}
		if err := uc.Delete(r.Context(), id); err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				responseError(w, http.StatusNotFound, "user not found")
				return
			}
			responseError(w, http.StatusInternalServerError, "failed to delete user")
			return
		}
		responseSuccess(w, http.StatusOK, map[string]string{"status": "deleted"})
	}
}
