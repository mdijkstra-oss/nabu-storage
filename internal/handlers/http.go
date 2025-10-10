package handlers

import (
	"encoding/json"
	"hermes-relay/internal/utils"
	"hermes-relay/internal/utils/dispatch"
	"net/http"
	"strings"
	"time"
)

// Command endpoint - returns business result
func CommandHandler(publisher *dispatch.InMemoryPublisher) http.HandlerFunc {
	return messageHandler(publisher, true, dispatch.Command)
}

// Event endpoint - just acknowledges
func EventHandler(publisher *dispatch.InMemoryPublisher) http.HandlerFunc {
	return messageHandler(publisher, false, dispatch.Event)
}

func messageHandler(publisher *dispatch.InMemoryPublisher, returnResult bool, messageType dispatch.MessageType) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var msg dispatch.Message
		err := json.NewDecoder(r.Body).Decode(&msg)
		if err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}

		msg.Type = messageType
		msg.Timestamp = time.Now()

		result, err := publisher.Publish(r.Context(), &msg)
		if err != nil {
			handleError(w, err)
			return
		}

		if !returnResult {
			w.WriteHeader(http.StatusOK)
			return
		}

		status := http.StatusOK
		if result != nil && isCreatedType(result.Action) {
			status = http.StatusCreated
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)

		if result != nil {
			json.NewEncoder(w).Encode(result)
		}
	}
}

func handleError(w http.ResponseWriter, err error) {
	switch {
	case utils.IsValidationError(err):
		respondError(w, http.StatusBadRequest, err.Error())
	case utils.IsNotFoundError(err):
		respondError(w, http.StatusNotFound, err.Error())
	case utils.IsConflictError(err):
		respondError(w, http.StatusConflict, err.Error())
	default:
		respondError(w, http.StatusInternalServerError, err.Error())
	}
}

func respondError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	errors := strings.Split(message, "\n")
	json.NewEncoder(w).Encode(ErrorResponse{
		Error: errors,
		Type:  getErrorTypeFromStatus(status),
	})
}

func getErrorTypeFromStatus(status int) string {
	switch status {
	case http.StatusBadRequest:
		return "validation_error"
	case http.StatusNotFound:
		return "not_found"
	case http.StatusConflict:
		return "conflict"
	default:
		return "internal_error"
	}
}

func isCreatedType(msgType string) bool {
	return strings.HasSuffix(msgType, "Created") ||
		strings.HasSuffix(msgType, "Added")
}
