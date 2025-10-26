package http

import (
	"encoding/json"
	"hermes-relay/internal/cqrs"
	"hermes-relay/internal/lib/utils"
	"net/http"
	"strings"
)

func successOutput(result *cqrs.AnyMessage, acceptedOnly bool) Output {
	if acceptedOnly {
		return Output{StatusCode: http.StatusAccepted}
	}

	status := http.StatusOK
	if result != nil && isCreatedAction(result.Action) {
		status = http.StatusCreated
	}

	var body []byte
	if result != nil {
		body, _ = json.Marshal(result)
	}

	return Output{
		StatusCode: status,
		Body:       body,
	}
}

func typedErrorOutput(err error) Output {
	switch {
	case utils.IsValidationError(err):
		return errorOutput(http.StatusBadRequest, err)
	case utils.IsNotFoundError(err):
		return errorOutput(http.StatusNotFound, err)
	case utils.IsConflictError(err):
		return errorOutput(http.StatusConflict, err)
	default:
		return errorOutput(http.StatusInternalServerError, err)
	}
}

func errorOutput(status int, err error) Output {
	response := ErrorResponse{Message: err.Error()}

	if ve, ok := err.(*utils.ValidationError); ok {
		response.Fields = ve.Fields
	}

	body, _ := json.Marshal(response)
	return Output{
		StatusCode: status,
		Body:       body,
	}
}

func batchStatus(total, successCount int, acceptedOnly bool) int {
	failureCount := total - successCount
	if failureCount == total {
		return http.StatusBadRequest
	}
	if failureCount > 0 {
		return http.StatusMultiStatus
	}
	if acceptedOnly {
		return http.StatusAccepted
	}
	return http.StatusOK
}

func isCreatedAction(action cqrs.Action) bool {
	s := string(action)
	return strings.HasSuffix(s, "Created") || strings.HasSuffix(s, "Added")
}
