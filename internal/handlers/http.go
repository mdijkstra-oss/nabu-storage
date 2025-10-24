package handlers

import (
	"bytes"
	"encoding/json"
	"hermes-relay/internal/cqrs"
	"hermes-relay/internal/lib/utils"
	"net/http"
	"slices"
	"strings"
	"time"
)

// Todo: Future, reject duplicates
// Todo: Future refactor, also only pub func
// Todo: This makes me sad comparing to the rest

type batchResult struct {
	Index   int              `json:"index"`
	Success bool             `json:"success"`
	Result  *cqrs.AnyMessage `json:"result,omitempty"`
	Error   string           `json:"error,omitempty"`
}

type batchResponse struct {
	Results      []batchResult `json:"results"`
	Total        int           `json:"total"`
	SuccessCount int           `json:"success_count"`
	FailureCount int           `json:"failure_count"`
}

// Command endpoint - returns business result
func CommandHandler(publisher *cqrs.InMemoryPublisher) http.HandlerFunc {
	return messageHandler(publisher, true, []cqrs.MessageType{cqrs.Command})
}

// DomainEvent endpoint - just acknowledges
func EventHandler(publisher *cqrs.InMemoryPublisher) http.HandlerFunc {
	return messageHandler(publisher, false, []cqrs.MessageType{cqrs.DomainEvent, cqrs.SystemEvent})
}

func messageHandler(publisher *cqrs.InMemoryPublisher, returnResult bool, allowedMessageTypes []cqrs.MessageType) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var raw json.RawMessage
		err := json.NewDecoder(r.Body).Decode(&raw)
		if err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}

		// Check if it's an array
		trimmed := bytes.TrimSpace(raw)
		if bytes.HasPrefix(trimmed, []byte("[")) {
			handleBatch(w, r, raw, publisher, returnResult, allowedMessageTypes)
		} else {
			handleSingle(w, r, raw, publisher, returnResult, allowedMessageTypes)
		}
	}
}

func handleSingle(w http.ResponseWriter, r *http.Request, raw json.RawMessage, publisher *cqrs.InMemoryPublisher, returnResult bool, allowedMessageTypes []cqrs.MessageType) {
	var msg cqrs.AnyMessage
	if err := json.Unmarshal(raw, &msg); err != nil {
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

	if result != nil && msg.Type == cqrs.Command {
		utils.WarnErr(json.NewEncoder(w).Encode(result))
	}
}

func handleBatch(w http.ResponseWriter, r *http.Request, raw json.RawMessage, publisher *cqrs.InMemoryPublisher, returnResult bool, allowedMessageTypes []cqrs.MessageType) {
	var messages []cqrs.AnyMessage
	if err := json.Unmarshal(raw, &messages); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	if len(messages) == 0 {
		respondError(w, http.StatusBadRequest, "empty batch")
		return
	}

	results := make([]batchResult, len(messages))
	successCount := 0
	failureCount := 0
	now := time.Now()

	for i, msg := range messages {
		results[i].Index = i

		// Validate message type
		if !slices.Contains(allowedMessageTypes, msg.Type) {
			results[i].Success = false
			results[i].Error = "message type is not allowed"
			failureCount++
			continue
		}

		msg.Timestamp = now

		// Publish message
		result, err := publisher.Publish(r.Context(), &msg)
		if err != nil {
			results[i].Success = false
			results[i].Error = err.Error()
			failureCount++
		} else {
			results[i].Success = true
			successCount++
			if returnResult && result != nil {
				results[i].Result = result
			}
		}
	}

	// Determine HTTP status
	status := http.StatusOK
	if failureCount == len(messages) {
		status = http.StatusBadRequest // All failed
	} else if failureCount > 0 {
		status = http.StatusMultiStatus // Partial success (207)
	}

	if !returnResult {
		w.WriteHeader(status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := batchResponse{
		Results:      results,
		Total:        len(messages),
		SuccessCount: successCount,
		FailureCount: failureCount,
	}

	utils.WarnErr(json.NewEncoder(w).Encode(response))
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

func isCreatedType(action cqrs.Action) bool {
	actionStr := string(action)
	return strings.HasSuffix(actionStr, "Created") ||
		strings.HasSuffix(actionStr, "Added")
}
