package http

import (
	"errors"
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/dispatch"
	th "hermes-relay/internal/lib/test-helpers"
	"hermes-relay/internal/lib/utils"
	"net/http"
	"testing"
)

func TestProcessCommand(t *testing.T) {
	tests := []struct {
		name         string
		body         string
		publish      dispatch.PublishFunc
		expectStatus int
		expectBody   string
	}{
		{
			name: "valid command returns 200",
			body: `{"type":"Command","action":"UpdateDocument","payload":{}}`,
			publish: func(msg *commands.AnyMessage) (*commands.AnyMessage, error) {
				th.AssertEqualSimple(t, "patient-zero", msg.Actor.UserID)
				th.AssertEqualSimple(t, commands.ActorTypeHuman, msg.Actor.ActorType)
				return &commands.AnyMessage{Action: "DocumentUpdated"}, nil
			},
			expectStatus: http.StatusOK,
			expectBody:   `{"action":"DocumentUpdated","type":"","actor":{"userId":"","actorType":""},"Timestamp":"0001-01-01T00:00:00Z"}`,
		},
		{
			name: "created action returns 201",
			body: `{"type":"Command","action":"CreateDocument","payload":{}}`,
			publish: func(msg *commands.AnyMessage) (*commands.AnyMessage, error) {
				return &commands.AnyMessage{Action: "DocumentCreated"}, nil
			},
			expectStatus: http.StatusCreated,
			expectBody:   `{"action":"DocumentCreated","type":"","actor":{"userId":"","actorType":""},"Timestamp":"0001-01-01T00:00:00Z"}`,
		},
		{
			name:         "invalid json returns 400",
			body:         `{invalid json`,
			publish:      nil,
			expectStatus: http.StatusBadRequest,
			expectBody:   `{"message":"invalid character 'i' looking for beginning of object key string"}`,
		},
		{
			name:         "wrong message type returns 400",
			body:         `{"type":"DomainEvent","action":"Something","payload":{}}`,
			publish:      nil,
			expectStatus: http.StatusBadRequest,
			expectBody:   `{"message":"message type not allowed"}`,
		},
		{
			name: "validation error returns 400 with fields",
			body: `{"type":"Command","action":"CreateDocument","payload":{}}`,
			publish: func(msg *commands.AnyMessage) (*commands.AnyMessage, error) {
				return nil, &utils.ValidationError{
					Message: "validation failed",
					Fields: map[string]string{
						"Title": "required",
					},
				}
			},
			expectStatus: http.StatusBadRequest,
			expectBody:   `{"message":"validation failed: Title required","fields":{"Title":"required"}}`,
		},
		{
			name: "not found error returns 404",
			body: `{"type":"Command","action":"UpdateDocument","payload":{}}`,
			publish: func(msg *commands.AnyMessage) (*commands.AnyMessage, error) {
				return nil, &utils.NotFoundError{Message: "document not found"}
			},
			expectStatus: http.StatusNotFound,
			expectBody:   `{"message":"document not found"}`,
		},
		{
			name: "conflict error returns 409",
			body: `{"type":"Command","action":"CreateDocument","payload":{}}`,
			publish: func(msg *commands.AnyMessage) (*commands.AnyMessage, error) {
				return nil, &utils.ConflictError{Message: "document already exists"}
			},
			expectStatus: http.StatusConflict,
			expectBody:   `{"message":"document already exists"}`,
		},
		{
			name: "generic error returns 500",
			body: `{"type":"Command","action":"UpdateDocument","payload":{}}`,
			publish: func(msg *commands.AnyMessage) (*commands.AnyMessage, error) {
				return nil, errors.New("something went wrong")
			},
			expectStatus: http.StatusInternalServerError,
			expectBody:   `{"message":"something went wrong"}`,
		},
		{
			name: "provided actor type is validated and preserved",
			body: `{"type":"Command","action":"UpdateDocument","payload":{},"actor":{"userId":"claude","actorType":"llm"}}`,
			publish: func(msg *commands.AnyMessage) (*commands.AnyMessage, error) {
				th.AssertEqualSimple(t, "patient-zero", msg.Actor.UserID)
				th.AssertEqualSimple(t, commands.ActorTypeLLM, msg.Actor.ActorType)
				return &commands.AnyMessage{Action: "DocumentUpdated"}, nil
			},
			expectStatus: http.StatusOK,
			expectBody:   `{"action":"DocumentUpdated","type":"","actor":{"userId":"","actorType":""},"Timestamp":"0001-01-01T00:00:00Z"}`,
		},
		{
			name:         "invalid actor type returns 400",
			body:         `{"type":"Command","action":"UpdateDocument","payload":{},"actor":{"actorType":"robot"}}`,
			publish:      nil,
			expectStatus: http.StatusBadRequest,
			expectBody:   `{"message":"invalid actor type: robot"}`,
		},
		{
			name: "missing actor type defaults to human",
			body: `{"type":"Command","action":"UpdateDocument","payload":{},"actor":{"userId":"specific-user"}}`,
			publish: func(msg *commands.AnyMessage) (*commands.AnyMessage, error) {
				th.AssertEqualSimple(t, "patient-zero", msg.Actor.UserID)
				th.AssertEqualSimple(t, commands.ActorTypeHuman, msg.Actor.ActorType)
				return &commands.AnyMessage{Action: "DocumentUpdated"}, nil
			},
			expectStatus: http.StatusOK,
			expectBody:   `{"action":"DocumentUpdated","type":"","actor":{"userId":"","actorType":""},"Timestamp":"0001-01-01T00:00:00Z"}`,
		},
		{
			name: "missing userId defaults to patient-zero",
			body: `{"type":"Command","action":"UpdateDocument","payload":{},"actor":{"actorType":"llm"}}`,
			publish: func(msg *commands.AnyMessage) (*commands.AnyMessage, error) {
				th.AssertEqualSimple(t, "patient-zero", msg.Actor.UserID)
				th.AssertEqualSimple(t, commands.ActorTypeLLM, msg.Actor.ActorType)
				return &commands.AnyMessage{Action: "DocumentUpdated"}, nil
			},
			expectStatus: http.StatusOK,
			expectBody:   `{"action":"DocumentUpdated","type":"","actor":{"userId":"","actorType":""},"Timestamp":"0001-01-01T00:00:00Z"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := Request{Body: []byte(tt.body)}
			response := ProcessCommand(request, tt.publish)

			th.AssertEqualSimple(t, tt.expectStatus, response.StatusCode)
			th.AssertEqualSimple(t, tt.expectBody, string(response.Body))
		})
	}
}
