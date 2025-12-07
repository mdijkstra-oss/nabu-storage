package http

import (
	"encoding/json"
	"errors"
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/lib/utils"
	"log/slog"
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

	if result != nil && hasPartialFailures(result.Payload) {
		return partialSuccessOutput(result)
	}

	var body []byte
	if result != nil {
		body = mustMarshal(result)
	}

	return Response{
		StatusCode: status,
		Body:       body,
	}
}

func partialSuccessOutput(result *commands.AnyMessage) Response {
	pf := result.Payload.(commands.PartialFailures)
	strippedResult := *result
	strippedResult.Payload = pf.WithoutFailures()
	response := partialSuccessResponse{
		Success:  &strippedResult,
		Failures: pf.GetFailures(),
	}
	body := mustMarshal(response)
	return Response{
		StatusCode: http.StatusMultiStatus,
		Body:       body,
	}
}

func hasPartialFailures(payload any) bool {
	if pf, ok := payload.(commands.PartialFailures); ok {
		return len(pf.GetFailures()) > 0
	}
	return false
}

func successQueryOutput(result any) Response {
	return Response{
		StatusCode: http.StatusOK,
		Body:       mustMarshal(result),
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

	return Response{
		StatusCode: status,
		Body:       mustMarshal(response),
	}
}

func mustMarshal(v any) []byte {
	body, err := json.Marshal(v)
	if err != nil {
		slog.Error("failed to marshal response", "error", err)
		return []byte(`{"message":"internal server error"}`)
	}
	return body
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

func WriteResponse(w http.ResponseWriter, response Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(response.StatusCode)
	if response.StatusCode >= 300 {
		slog.Error("error response", "status", response.StatusCode, "body", string(response.Body))
	}
	utils.Should(w.Write(response.Body))
}
