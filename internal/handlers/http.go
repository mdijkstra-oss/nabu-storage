package handlers

import (
	"encoding/json"
	commands2 "hermes-relay/internal/commands"
	"hermes-relay/internal/utils"
	"net/http"
	"slices"
	"strings"
	"time"
)

// Todo: Future, reject duplicates

// Command endpoint - returns business result
func CommandHandler(publisher *commands2.InMemoryPublisher) http.HandlerFunc {
	return messageHandler(publisher, true, []commands2.MessageType{commands2.Command})
}

// DomainEvent endpoint - just acknowledges
func EventHandler(publisher *commands2.InMemoryPublisher) http.HandlerFunc {
	return messageHandler(publisher, false, []commands2.MessageType{commands2.DomainEvent, commands2.SystemEvent})
}

func messageHandler(publisher *commands2.InMemoryPublisher, returnResult bool, allowedMessageTypes []commands2.MessageType) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var msg commands2.Message
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

		if result != nil && msg.Type == commands2.Command {
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
