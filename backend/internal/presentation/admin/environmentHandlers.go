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

func listEnvironmentsHandler(uc *application.Environment) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		envs, err := uc.List(r.Context())
		if err != nil {
			responseError(w, http.StatusInternalServerError, "failed to list environments")
			return
		}
		responseSuccess(w, http.StatusOK, envs)
	}
}

func getEnvironmentHandler(uc *application.Environment) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			responseError(w, http.StatusBadRequest, "invalid environment id")
			return
		}
		env, err := uc.Get(r.Context(), id)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				responseError(w, http.StatusNotFound, "environment not found")
				return
			}
			responseError(w, http.StatusInternalServerError, "failed to fetch environment")
			return
		}
		responseSuccess(w, http.StatusOK, env)
	}
}

func createEnvironmentHandler(uc *application.Environment) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var env entity.Environment
		if err := jsonAPI.NewDecoder(r.Body).Decode(&env); err != nil {
			responseError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		id, err := uc.Create(r.Context(), env)
		if err != nil {
			if errors.Is(err, application.ErrInvalidEnvironment) {
				responseError(w, http.StatusBadRequest, err.Error())
				return
			}
			responseError(w, http.StatusInternalServerError, "failed to create environment")
			return
		}
		responseSuccess(w, http.StatusCreated, map[string]int{"id": id})
	}
}

func updateEnvironmentHandler(uc *application.Environment) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			responseError(w, http.StatusBadRequest, "invalid environment id")
			return
		}
		var env entity.Environment
		if err := jsonAPI.NewDecoder(r.Body).Decode(&env); err != nil {
			responseError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		env.ID = id
		if err := uc.Update(r.Context(), env); err != nil {
			if errors.Is(err, application.ErrInvalidEnvironment) {
				responseError(w, http.StatusBadRequest, err.Error())
				return
			}
			if errors.Is(err, repository.ErrNotFound) {
				responseError(w, http.StatusNotFound, "environment not found")
				return
			}
			responseError(w, http.StatusInternalServerError, "failed to update environment")
			return
		}
		responseSuccess(w, http.StatusOK, map[string]string{"status": "updated"})
	}
}

func deleteEnvironmentHandler(uc *application.Environment) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			responseError(w, http.StatusBadRequest, "invalid environment id")
			return
		}
		if err := uc.Delete(r.Context(), id); err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				responseError(w, http.StatusNotFound, "environment not found")
				return
			}
			responseError(w, http.StatusInternalServerError, "failed to delete environment")
			return
		}
		responseSuccess(w, http.StatusOK, map[string]string{"status": "deleted"})
	}
}
