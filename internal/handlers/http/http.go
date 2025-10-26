package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"hermes-relay/internal/cqrs"
	"hermes-relay/internal/lib/utils"
	"net/http"
	"slices"
	"strings"
	"time"
)

type ErrorResponse struct {
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

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

func CommandHandler(publish cqrs.PublishFunc) http.HandlerFunc {
	return messageHandler(publish, true, []cqrs.MessageType{cqrs.Command})
}

func EventHandler(publish cqrs.PublishFunc) http.HandlerFunc {
	return messageHandler(publish, false, []cqrs.MessageType{cqrs.DomainEvent, cqrs.SystemEvent})
}

func messageHandler(publish cqrs.PublishFunc, returnResult bool, allowedTypes []cqrs.MessageType) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var raw json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
			respondError(w, http.StatusBadRequest, err)
			return
		}

		if isBatch(raw) {
			handleBatch(w, r, raw, publish, returnResult, allowedTypes)
		} else {
			handleSingle(w, r, raw, publish, returnResult, allowedTypes)
		}
	}
}

func handleSingle(w http.ResponseWriter, r *http.Request, raw json.RawMessage, publish cqrs.PublishFunc, returnResult bool, allowedTypes []cqrs.MessageType) {
	msg, err := parseMessage(raw, allowedTypes)
	if err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}

	result, err := publish(r.Context(), msg)
	if err != nil {
		respondTypedError(w, err)
		return
	}

	if !returnResult {
		w.WriteHeader(http.StatusOK)
		return
	}

	respondSuccess(w, result)
}

func handleBatch(w http.ResponseWriter, r *http.Request, raw json.RawMessage, publish cqrs.PublishFunc, returnResult bool, allowedTypes []cqrs.MessageType) {
	var messages []cqrs.AnyMessage
	if err := json.Unmarshal(raw, &messages); err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}

	if len(messages) == 0 {
		respondError(w, http.StatusBadRequest, errors.New("empty batch"))
		return
	}

	response := processBatch(r.Context(), messages, publish, allowedTypes, returnResult)

	if !returnResult {
		w.WriteHeader(batchStatus(response))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(batchStatus(response))
	utils.WarnErr(json.NewEncoder(w).Encode(response))
}

func processBatch(ctx context.Context, messages []cqrs.AnyMessage, publish cqrs.PublishFunc, allowedTypes []cqrs.MessageType, returnResult bool) batchResponse {
	results := make([]batchResult, len(messages))
	successCount := 0
	now := time.Now()

	for i, msg := range messages {
		results[i].Index = i

		if !slices.Contains(allowedTypes, msg.Type) {
			results[i].Success = false
			results[i].Error = "message type not allowed"
			continue
		}

		msg.Timestamp = now
		result, err := publish(ctx, &msg)

		if err != nil {
			results[i].Success = false
			results[i].Error = err.Error()
		} else {
			results[i].Success = true
			successCount++
			if returnResult && result != nil {
				results[i].Result = result
			}
		}
	}

	return batchResponse{
		Results:      results,
		Total:        len(messages),
		SuccessCount: successCount,
		FailureCount: len(messages) - successCount,
	}
}

func parseMessage(raw json.RawMessage, allowedTypes []cqrs.MessageType) (*cqrs.AnyMessage, error) {
	var msg cqrs.AnyMessage
	if err := json.Unmarshal(raw, &msg); err != nil {
		return nil, err
	}

	if !slices.Contains(allowedTypes, msg.Type) {
		return nil, errors.New("message type not allowed")
	}

	msg.Timestamp = time.Now()
	return &msg, nil
}

func isBatch(raw json.RawMessage) bool {
	return bytes.HasPrefix(bytes.TrimSpace(raw), []byte("["))
}

func batchStatus(response batchResponse) int {
	if response.FailureCount == response.Total {
		return http.StatusBadRequest
	}
	if response.FailureCount > 0 {
		return http.StatusMultiStatus
	}
	return http.StatusOK
}

func respondSuccess(w http.ResponseWriter, result *cqrs.AnyMessage) {
	status := http.StatusOK
	if result != nil && isCreatedAction(result.Action) {
		status = http.StatusCreated
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if result != nil {
		utils.WarnErr(json.NewEncoder(w).Encode(result))
	}
}

func respondTypedError(w http.ResponseWriter, err error) {
	switch {
	case utils.IsValidationError(err):
		respondError(w, http.StatusBadRequest, err)
	case utils.IsNotFoundError(err):
		respondError(w, http.StatusNotFound, err)
	case utils.IsConflictError(err):
		respondError(w, http.StatusConflict, err)
	default:
		respondError(w, http.StatusInternalServerError, err)
	}
}

func respondError(w http.ResponseWriter, status int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := ErrorResponse{Message: err.Error()}

	if ve, ok := err.(*utils.ValidationError); ok {
		response.Fields = ve.Fields
	}

	utils.WarnErr(json.NewEncoder(w).Encode(response))
}

func isCreatedAction(action cqrs.Action) bool {
	s := string(action)
	return strings.HasSuffix(s, "Created") || strings.HasSuffix(s, "Added")
}
