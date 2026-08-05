package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"nabu-storage/internal/domain"
	"nabu-storage/internal/lib/utils"

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
		projectID, ok := utils.CanonicalID(chi.URLParam(r, "projectId"))
		if !ok {
			writeError(w, http.StatusBadRequest, "invalid projectId")
			return
		}

		cmd, err := decodeCommand(r)
		if err != nil {
			writeDecodeError(w, err)
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

func writeDecodeError(w http.ResponseWriter, err error) {
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		writeError(w, http.StatusRequestEntityTooLarge, "request body too large")
		return
	}
	writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
}

// Every failure leaves by writeError, so a client reads one key to learn what went
// wrong rather than one key per kind of failure.
func writeTypedError(w http.ResponseWriter, err error) {
	switch e := err.(type) {
	case *utils.ValidationError:
		writeError(w, http.StatusBadRequest, e.Error())
	case *utils.NotFoundError:
		writeError(w, http.StatusNotFound, e.Error())
	default:
		slog.Error("internal error", "error", err)
		writeError(w, http.StatusInternalServerError, "internal error")
	}
}
