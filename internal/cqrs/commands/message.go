package commands

import (
	"encoding/json"
	"hermes-relay/internal/lib/utils"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

type Action string
type MessageType string

type AggregateType string

// Errors can be part of entity story! Eg oh no, at this point in time I could not parse file
// So set uploaded file to state 'broken' or something so that can be reflected in client too!
const (
	Command     MessageType = "Command"
	DomainEvent MessageType = "DomainEvent"
	SystemEvent MessageType = "SystemEvent"
)

// Potential future events
//Integration events: when other bounded contexts need notification. OrderPlaced → email service, shipping service. Don't replay - they're outputs.
//System events: scheduled jobs, webhooks, monitoring. PaymentWebhook → generates command → domain event. Cron cleanup → generates command → domain event.
//Integration/system can trigger commands, which generate domain events. But integration/system events themselves aren't replayed.

// Todo: validation here?

type Message[T any] struct {
	ID string `json:"id,omitempty"`

	Action Action      `json:"action" validate:"required"`
	Type   MessageType `json:"type"` // todo validate etc

	AggregateID   string        `json:"aggregateId,omitempty"`
	AggregateType AggregateType `json:"aggregateType,omitempty"`
	Payload       T             `json:"payload,omitempty"`

	CausationID string    `json:"causationId,omitempty"`
	Timestamp   time.Time //`json:"timestamp" validate:"required"`
}

type AnyMessage = Message[any]

// Todo more message things for tracking etc.
// Versioning
// Parent
/*
type Event struct {
    ID            string
    AggregateID   string
    Version       int
    CausationID   string  // immediate cause (command/event ID)
    CorrelationID string  // workflow/request ID
}
Example:

User request: correlation ID = "req-123"
CreateAccount command: causation = "req-123"
AccountCreated event: causation = "cmd-456", correlation = "req-123"
WelcomeEmailSent event: causation = "account-created-789", correlation = "req-123"

Causation = parent. Correlation = trace entire flow.
*/

func ValidateMessage[T any](m *Message[T]) *utils.ValidationError {
	return utils.ToValidationError(utils.Validate.Struct(m))
}

func EnsureValidPayload[T any](m *Message[T], p any) *utils.ValidationError {
	if err := UnmarshallPayload(m, p); err != nil {
		return utils.ToValidationError(err)
	}

	return utils.ToValidationError(utils.Validate.Struct(p))
}

func (m Message[T]) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("id", m.ID),
		slog.String("action", string(m.Action)),
		slog.String("type", string(m.Type)),
		slog.String("aggregateId", m.AggregateID),
		slog.String("causationId", m.CausationID),
		slog.Time("timestamp", m.Timestamp),
	)
}

func NewMessage[T, C any](mType MessageType, action Action, payload T, aggregateType AggregateType, aggregateID string, cause *Message[C]) *Message[T] {
	var causationID = ""
	if cause != nil {
		causationID = cause.ID
	}

	return &Message[T]{
		ID:            uuid.NewString(),
		Action:        action,
		Type:          mType,
		AggregateID:   aggregateID,
		AggregateType: aggregateType,
		Payload:       payload,
		CausationID:   causationID,
		Timestamp:     time.Now(),
	}
}

func NewCommand[T, P any](action Action, payload T, aggregateType AggregateType, aggregateID string, cause *Message[P]) *Message[T] {
	return NewMessage(Command, action, payload, aggregateType, aggregateID, cause)
}

func NewDomainEvent[T, P any](action Action, payload T, aggregateType AggregateType, aggregateID string, cause *Message[P]) *Message[T] {
	return NewMessage(DomainEvent, action, payload, aggregateType, aggregateID, cause)
}

func ToDomainEvent[T any](message *Message[T], event Action) *Message[T] {
	return NewMessage(
		DomainEvent,
		event,
		message.Payload,
		message.AggregateType,
		message.AggregateID,
		message,
	)
}

func UnmarshallPayload[T any](m *Message[T], payload any) error {
	data, err := json.Marshal(m.Payload)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, payload)
}

func ToAny[P any](m *Message[P]) *AnyMessage {
	return &AnyMessage{
		ID:            m.ID,
		Action:        m.Action,
		Type:          m.Type,
		AggregateID:   m.AggregateID,
		AggregateType: m.AggregateType,
		CausationID:   m.CausationID,
		Timestamp:     m.Timestamp,
		Payload:       m.Payload,
	}
}

func ExtractProjectID(message *AnyMessage) string {
	if message.AggregateType == "Project" {
		return message.AggregateID
	}

	if payload, ok := message.Payload.(map[string]any); ok {
		if projectID, ok := payload["project_id"].(string); ok {
			return projectID
		}
	}

	return ""
}
