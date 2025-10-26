package http

import (
	"context"
	"encoding/json"
	"errors"
	"hermes-relay/internal/cqrs"
	"hermes-relay/internal/lib/utils"
	"net/http"
	"testing"
)

func TestProcessCommand(t *testing.T) {
	tests := []struct {
		name         string
		body         string
		publish      cqrs.PublishFunc
		expectStatus int
		checkBody    func(t *testing.T, body []byte)
	}{
		{
			name: "valid command returns 200",
			body: `{"type":"Command","action":"UpdateDocument","payload":{}}`,
			publish: func(ctx context.Context, msg *cqrs.AnyMessage) (*cqrs.AnyMessage, error) {
				return &cqrs.AnyMessage{Action: "DocumentUpdated"}, nil
			},
			expectStatus: http.StatusOK,
			checkBody:    expectAction("DocumentUpdated"),
		},
		{
			name: "created action returns 201",
			body: `{"type":"Command","action":"CreateDocument","payload":{}}`,
			publish: func(ctx context.Context, msg *cqrs.AnyMessage) (*cqrs.AnyMessage, error) {
				return &cqrs.AnyMessage{Action: "DocumentCreated"}, nil
			},
			expectStatus: http.StatusCreated,
			checkBody:    expectAction("DocumentCreated"),
		},
		{
			name:         "invalid json returns 400",
			body:         `{invalid json`,
			publish:      nil,
			expectStatus: http.StatusBadRequest,
			checkBody:    expectErrorWithMessage(),
		},
		{
			name:         "wrong message type returns 400",
			body:         `{"type":"DomainEvent","action":"Something","payload":{}}`,
			publish:      nil,
			expectStatus: http.StatusBadRequest,
			checkBody:    expectErrorMessage("message type not allowed"),
		},
		{
			name: "validation error returns 400 with fields",
			body: `{"type":"Command","action":"CreateDocument","payload":{}}`,
			publish: func(ctx context.Context, msg *cqrs.AnyMessage) (*cqrs.AnyMessage, error) {
				return nil, &utils.ValidationError{
					Message: "validation failed",
					Fields: map[string]string{
						"Title": "required",
					},
				}
			},
			expectStatus: http.StatusBadRequest,
			checkBody: expectValidationError("validation failed", map[string]string{
				"Title": "required",
			}),
		},
		{
			name: "not found error returns 404",
			body: `{"type":"Command","action":"UpdateDocument","payload":{}}`,
			publish: func(ctx context.Context, msg *cqrs.AnyMessage) (*cqrs.AnyMessage, error) {
				return nil, &utils.NotFoundError{Message: "document not found"}
			},
			expectStatus: http.StatusNotFound,
			checkBody:    expectErrorMessage("document not found"),
		},
		{
			name: "conflict error returns 409",
			body: `{"type":"Command","action":"CreateDocument","payload":{}}`,
			publish: func(ctx context.Context, msg *cqrs.AnyMessage) (*cqrs.AnyMessage, error) {
				return nil, &utils.ConflictError{Message: "document already exists"}
			},
			expectStatus: http.StatusConflict,
			checkBody:    expectErrorMessage("document already exists"),
		},
		{
			name: "generic error returns 500",
			body: `{"type":"Command","action":"UpdateDocument","payload":{}}`,
			publish: func(ctx context.Context, msg *cqrs.AnyMessage) (*cqrs.AnyMessage, error) {
				return nil, errors.New("something went wrong")
			},
			expectStatus: http.StatusInternalServerError,
			checkBody:    expectErrorMessage("something went wrong"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := Input{Body: []byte(tt.body)}
			output := ProcessCommand(context.Background(), input, tt.publish)

			assertEqual(t, tt.expectStatus, output.StatusCode)

			if tt.checkBody != nil {
				tt.checkBody(t, output.Body)
			}
		})
	}
}

// Expectation helpers
func expectAction(action string) func(*testing.T, []byte) {
	return func(t *testing.T, body []byte) {
		var result cqrs.AnyMessage
		mustUnmarshal(t, body, &result)
		assertEqual(t, action, string(result.Action))
	}
}

func expectErrorMessage(msg string) func(*testing.T, []byte) {
	return func(t *testing.T, body []byte) {
		var errResp ErrorResponse
		mustUnmarshal(t, body, &errResp)
		assertEqual(t, msg, errResp.Message)
	}
}

func expectErrorWithMessage() func(*testing.T, []byte) {
	return func(t *testing.T, body []byte) {
		var errResp ErrorResponse
		mustUnmarshal(t, body, &errResp)
		assertNotEmpty(t, errResp.Message)
	}
}

func expectValidationError(msg string, fields map[string]string) func(*testing.T, []byte) {
	return func(t *testing.T, body []byte) {
		var errResp ErrorResponse
		mustUnmarshal(t, body, &errResp)
		assertEqual(t, msg, errResp.Message)
		for field, expected := range fields {
			assertEqual(t, expected, errResp.Fields[field])
		}
	}
}

// Test helpers
func mustUnmarshal(t *testing.T, data []byte, v any) {
	t.Helper()
	if err := json.Unmarshal(data, v); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
}

func assertEqual(t *testing.T, expected, actual any) {
	t.Helper()
	if expected != actual {
		t.Errorf("expected %v, got %v", expected, actual)
	}
}

func assertNotEmpty(t *testing.T, s string) {
	t.Helper()
	if s == "" {
		t.Error("expected non-empty string")
	}
}
