package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"hermes-relay/internal/cqrs"
	"net/http"
	"slices"
	"time"
)

func ProcessCommand(ctx context.Context, input Input, publish cqrs.PublishFunc) Output {
	return processMessage(ctx, input, publish, false, []cqrs.MessageType{cqrs.Command})
}

// - No tests yet but unused too, could share tests with above
//func ProcessEvent(ctx context.Context, input Input, publish cqrs.PublishFunc) Output {
//	return processMessage(ctx, input, publish, true, []cqrs.MessageType{cqrs.DomainEvent, cqrs.SystemEvent})
//}

func processMessage(ctx context.Context, input Input, publish cqrs.PublishFunc, acceptedOnly bool, allowedTypes []cqrs.MessageType) Output {
	trimmed := bytes.TrimSpace(input.Body)
	if bytes.HasPrefix(trimmed, []byte("[")) {
		return processBatch(ctx, input, publish, acceptedOnly, allowedTypes)
	}
	return processSingle(ctx, input, publish, acceptedOnly, allowedTypes)
}

func processSingle(ctx context.Context, input Input, publish cqrs.PublishFunc, acceptedOnly bool, allowedTypes []cqrs.MessageType) Output {
	var msg cqrs.AnyMessage
	if err := json.Unmarshal(input.Body, &msg); err != nil {
		return errorOutput(http.StatusBadRequest, err)
	}

	if !slices.Contains(allowedTypes, msg.Type) {
		return errorOutput(http.StatusBadRequest, errors.New("message type not allowed"))
	}

	msg.Timestamp = time.Now()
	result, err := publish(ctx, &msg)
	if err != nil {
		return typedErrorOutput(err)
	}

	return successOutput(result, acceptedOnly)
}

func processBatch(ctx context.Context, input Input, publish cqrs.PublishFunc, acceptedOnly bool, allowedTypes []cqrs.MessageType) Output {
	var messages []cqrs.AnyMessage
	if err := json.Unmarshal(input.Body, &messages); err != nil {
		return errorOutput(http.StatusBadRequest, err)
	}

	if len(messages) == 0 {
		return errorOutput(http.StatusBadRequest, errors.New("empty batch"))
	}

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
			if !acceptedOnly && result != nil {
				results[i].Result = result
			}
		}
	}

	response := batchResponse{
		Results:      results,
		Total:        len(results),
		SuccessCount: successCount,
		FailureCount: len(results) - successCount,
	}

	body, _ := json.Marshal(response)
	return Output{
		StatusCode: batchStatus(len(results), successCount, acceptedOnly),
		Body:       body,
	}
}
