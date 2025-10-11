package handlers

import (
	"encoding/json"
	"hermes-relay/internal/utils"
	"hermes-relay/internal/utils/dispatch"
	"net/http"
	"slices"
	"strings"
	"time"
)

// Command endpoint - returns business result
func CommandHandler(publisher *dispatch.InMemoryPublisher) http.HandlerFunc {
	return messageHandler(publisher, true, []dispatch.MessageType{dispatch.Command})
}

// DomainEvent endpoint - just acknowledges
func EventHandler(publisher *dispatch.InMemoryPublisher) http.HandlerFunc {
	return messageHandler(publisher, false, []dispatch.MessageType{dispatch.DomainEvent, dispatch.SystemEvent})
}

func messageHandler(publisher *dispatch.InMemoryPublisher, returnResult bool, allowedMessageTypes []dispatch.MessageType) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var msg dispatch.Message
		err := json.NewDecoder(r.Body).Decode(&msg)
		if err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}

		if !slices.Contains(allowedMessageTypes, msg.Type) {
			respondError(w, http.StatusBadRequest, "message type is not allowed")
			return
		}

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

		if result != nil && msg.Type == dispatch.Command {
			utils.WarnErr(json.NewEncoder(w).Encode(result))
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
	utils.WarnErr(json.NewEncoder(w).Encode(ErrorResponse{
		Error: errors,
		Type:  getErrorTypeFromStatus(status),
	}))
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
