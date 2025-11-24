package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/dispatch"
	"hermes-relay/internal/lib/utils"
	"net/http"
	"slices"
	"time"
)

func ProcessCommand(request Request, publish dispatch.PublishFunc) Response {
	return processMessage(request, publish, false, []commands.MessageType{commands.Command})
}

// - No tests yet but unused too, could share tests with above
//func ProcessEvent(request Request, publish cqrs.PublishFunc) Response {
//	return processMessage(request, publish, true, []cqrs.MessageType{cqrs.DomainEvent, cqrs.SystemEvent})
//}

func processMessage(request Request, publish dispatch.PublishFunc, acceptedOnly bool, allowedTypes []commands.MessageType) Response {
	trimmed := bytes.TrimSpace(request.Body)
	if bytes.HasPrefix(trimmed, []byte("[")) {
		return processBatch(request, publish, acceptedOnly, allowedTypes)
	}
	return processSingle(request, publish, acceptedOnly, allowedTypes)
}

func processSingle(request Request, publish dispatch.PublishFunc, acceptedOnly bool, allowedTypes []commands.MessageType) Response {
	var msg commands.AnyMessage
	if err := json.Unmarshal(request.Body, &msg); err != nil {
		return errorOutput(http.StatusBadRequest, err)
	}

	if !slices.Contains(allowedTypes, msg.Type) {
		return errorOutput(http.StatusBadRequest, errors.New("message type not allowed"))
	}

	if err := ensureActor(&msg); err != nil {
		return errorOutput(http.StatusBadRequest, err)
	}

	msg.ID = utils.NewID()
	msg.Timestamp = time.Now()
	result, err := publish(&msg)
	if err != nil {
		return typedErrorOutput(err)
	}

	return successOutput(result, acceptedOnly)
}

func processBatch(request Request, publish dispatch.PublishFunc, acceptedOnly bool, allowedTypes []commands.MessageType) Response {
	var messages []commands.AnyMessage
	if err := json.Unmarshal(request.Body, &messages); err != nil {
		return errorOutput(http.StatusBadRequest, err)
	}

	if len(messages) == 0 {
		return errorOutput(http.StatusBadRequest, errors.New("empty batch"))
	}

	now := time.Now()

	results := utils.MapWithIndex(messages, func(i int, msg commands.AnyMessage) batchResult {
		if !slices.Contains(allowedTypes, msg.Type) {
			return batchResult{
				Index:   i,
				Success: false,
				Error:   "message type not allowed",
			}
		}

		if err := ensureActor(&msg); err != nil {
			return batchResult{
				Index:   i,
				Success: false,
				Error:   err.Error(),
			}
		}

		msg.ID = utils.NewID()
		msg.Timestamp = now
		result, err := publish(&msg)

		if err != nil {
			return batchResult{
				Index:   i,
				Success: false,
				Error:   err.Error(),
			}
		}

		var resultData *commands.AnyMessage
		if !acceptedOnly && result != nil {
			resultData = result
		}

		return batchResult{
			Index:   i,
			Success: true,
			Result:  resultData,
		}
	})

	successCount := utils.Reduce(results, 0, func(count int, r batchResult) int {
		if r.Success {
			return count + 1
		}
		return count
	})

	response := batchResponse{
		Results:      results,
		Total:        len(results),
		SuccessCount: successCount,
		FailureCount: len(results) - successCount,
	}

	body, _ := json.Marshal(response)
	return Response{
		StatusCode: batchStatus(len(results), successCount, acceptedOnly),
		Body:       body,
	}
}

func ensureActor(msg *commands.AnyMessage) error {
	msg.Actor.UserID = "patient-zero"

	if msg.Actor.ActorType == "" {
		msg.Actor.ActorType = commands.ActorTypeHuman
	}

	validTypes := []commands.ActorType{
		commands.ActorTypeHuman,
		commands.ActorTypeLLM,
		commands.ActorTypeSystem,
	}

	if !slices.Contains(validTypes, msg.Actor.ActorType) {
		return fmt.Errorf("invalid actor type: %s", msg.Actor.ActorType)
	}

	return nil
}

// Todo: Reject client-provided Actor fields when auth is implemented
// Once proper authentication exists:
// - Extract UserID from auth context (JWT, session, etc)
// - Reject any client-provided Actor.UserID that doesn't match authenticated user
// - Only allow ActorType override for specific privileged users (e.g., system accounts)
// - Validate that client cannot impersonate other users or claim system/LLM actor types
// This ensures audit trail integrity and prevents actor spoofing
