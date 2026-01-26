package http

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"hermes-relay/internal/domain"
	"hermes-relay/internal/lib/utils"

	"github.com/go-chi/chi/v5"
)

type StatusResponse struct {
	Status string `json:"status"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func CommandHandler(baseDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID := chi.URLParam(r, "projectId")
		if !utils.ValidID(projectID) {
			writeError(w, http.StatusBadRequest, "invalid projectId")
			return
		}

		cmd, err := decodeCommand(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
			return
		}

		if err := domain.Execute(cmd, projectID, baseDir); err != nil {
			writeTypedError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, StatusResponse{Status: "ok"})
	}
}

func decodeCommand(r *http.Request) (*domain.Command, error) {
	var cmd domain.Command
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		return nil, err
	}
	return &cmd, nil
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	utils.ShouldWork(json.NewEncoder(w).Encode(data))
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, ErrorResponse{Error: message})
}

func writeTypedError(w http.ResponseWriter, err error) {
	switch e := err.(type) {
	case *utils.ValidationError:
		writeJSON(w, http.StatusBadRequest, e)
	case *utils.NotFoundError:
		writeJSON(w, http.StatusNotFound, e)
	case *utils.ConflictError:
		writeJSON(w, http.StatusConflict, e)
	default:
		slog.Error("internal error", "error", err)
		writeError(w, http.StatusInternalServerError, "internal error")
	}
}
