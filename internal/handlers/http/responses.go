package http

import (
	"encoding/json"
	"errors"
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/lib/utils"
	"net/http"
	"strings"
)

func successOutput(result *commands.AnyMessage, acceptedOnly bool) Response {
	if acceptedOnly {
		return Response{StatusCode: http.StatusAccepted}
	}

	status := http.StatusOK
	if result != nil && isCreatedAction(result.Action) {
		status = http.StatusCreated
	}

	var body []byte
	if result != nil {
		body, _ = json.Marshal(result)
	}

	return Response{
		StatusCode: status,
		Body:       body,
	}
}

func typedErrorOutput(err error) Response {
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

func errorOutput(status int, err error) Response {
	response := ErrorResponse{Message: err.Error()}

	var ve *utils.ValidationError
	if errors.As(err, &ve) {
		response.Fields = ve.Fields
	}

	body, _ := json.Marshal(response)
	return Response{
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

func isCreatedAction(action commands.Action) bool {
	s := string(action)
	return strings.HasSuffix(s, "Created") || strings.HasSuffix(s, "Added")
}
