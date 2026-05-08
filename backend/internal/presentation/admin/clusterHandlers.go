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

func listClustersHandler(uc *application.Cluster) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clusters, err := uc.List(r.Context())
		if err != nil {
			responseError(w, http.StatusInternalServerError, "failed to list clusters")
			return
		}
		responseSuccess(w, http.StatusOK, clusters)
	}
}

func getClusterHandler(uc *application.Cluster) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			responseError(w, http.StatusBadRequest, "invalid cluster id")
			return
		}
		cluster, err := uc.Get(r.Context(), id)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				responseError(w, http.StatusNotFound, "cluster not found")
				return
			}
			responseError(w, http.StatusInternalServerError, "failed to fetch cluster")
			return
		}
		responseSuccess(w, http.StatusOK, cluster)
	}
}

func createClusterHandler(uc *application.Cluster) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var cluster entity.Cluster
		if err := jsonAPI.NewDecoder(r.Body).Decode(&cluster); err != nil {
			responseError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		id, err := uc.Create(r.Context(), cluster)
		if err != nil {
			if errors.Is(err, application.ErrInvalidCluster) {
				responseError(w, http.StatusBadRequest, err.Error())
				return
			}
			responseError(w, http.StatusInternalServerError, "failed to create cluster")
			return
		}
		responseSuccess(w, http.StatusCreated, map[string]int{"id": id})
	}
}

func updateClusterHandler(uc *application.Cluster) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			responseError(w, http.StatusBadRequest, "invalid cluster id")
			return
		}
		var cluster entity.Cluster
		if err := jsonAPI.NewDecoder(r.Body).Decode(&cluster); err != nil {
			responseError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		cluster.ID = id
		if err := uc.Update(r.Context(), cluster); err != nil {
			if errors.Is(err, application.ErrInvalidCluster) {
				responseError(w, http.StatusBadRequest, err.Error())
				return
			}
			if errors.Is(err, repository.ErrNotFound) {
				responseError(w, http.StatusNotFound, "cluster not found")
				return
			}
			responseError(w, http.StatusInternalServerError, "failed to update cluster")
			return
		}
		responseSuccess(w, http.StatusOK, map[string]string{"status": "updated"})
	}
}

func deleteClusterHandler(uc *application.Cluster) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			responseError(w, http.StatusBadRequest, "invalid cluster id")
			return
		}
		if err := uc.Delete(r.Context(), id); err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				responseError(w, http.StatusNotFound, "cluster not found")
				return
			}
			responseError(w, http.StatusInternalServerError, "failed to delete cluster")
			return
		}
		responseSuccess(w, http.StatusOK, map[string]string{"status": "deleted"})
	}
}

func addClusterTargetHandler(uc *application.Cluster) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clusterID, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			responseError(w, http.StatusBadRequest, "invalid cluster id")
			return
		}
		var target entity.ClusterTarget
		if err := jsonAPI.NewDecoder(r.Body).Decode(&target); err != nil {
			responseError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		target.ClusterID = clusterID
		if target.Weight == 0 {
			target.Weight = 1
		}
		target.IsHealthy = true
		id, err := uc.AddTarget(r.Context(), target)
		if err != nil {
			if errors.Is(err, application.ErrInvalidCluster) {
				responseError(w, http.StatusBadRequest, err.Error())
				return
			}
			responseError(w, http.StatusInternalServerError, "failed to add target")
			return
		}
		responseSuccess(w, http.StatusCreated, map[string]int{"id": id})
	}
}

func updateClusterTargetHandler(uc *application.Cluster) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		targetID, err := strconv.Atoi(chi.URLParam(r, "tid"))
		if err != nil {
			responseError(w, http.StatusBadRequest, "invalid target id")
			return
		}
		var target entity.ClusterTarget
		if err := jsonAPI.NewDecoder(r.Body).Decode(&target); err != nil {
			responseError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		target.ID = targetID
		if err := uc.UpdateTarget(r.Context(), target); err != nil {
			if errors.Is(err, application.ErrInvalidCluster) {
				responseError(w, http.StatusBadRequest, err.Error())
				return
			}
			if errors.Is(err, repository.ErrNotFound) {
				responseError(w, http.StatusNotFound, "target not found")
				return
			}
			responseError(w, http.StatusInternalServerError, "failed to update target")
			return
		}
		responseSuccess(w, http.StatusOK, map[string]string{"status": "updated"})
	}
}

func deleteClusterTargetHandler(uc *application.Cluster) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		targetID, err := strconv.Atoi(chi.URLParam(r, "tid"))
		if err != nil {
			responseError(w, http.StatusBadRequest, "invalid target id")
			return
		}
		if err := uc.RemoveTarget(r.Context(), targetID); err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				responseError(w, http.StatusNotFound, "target not found")
				return
			}
			responseError(w, http.StatusInternalServerError, "failed to remove target")
			return
		}
		responseSuccess(w, http.StatusOK, map[string]string{"status": "deleted"})
	}
}
