package cqrs

import (
	"encoding/json"
	"log/slog"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type Action string
type MessageType string

type AggregateType string

const (
	Command     MessageType = "Command"
	DomainEvent MessageType = "DomainEvent"
	SystemEvent MessageType = "systemEvent"
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

func ValidateMessage[T any](m *Message[T]) error {
	validate := validator.New()
	return validate.Struct(m)
}

func EnsureValidPayload[T any](m *Message[T], p any) error {
	validate := validator.New()

	err := UnmarshallPayload(m, p)
	if err != nil {
		return err
	}

	return validate.Struct(p)
}

// Wastefull generic fornow.
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

func NewMessage[T any, C any](mType MessageType, action Action, payload T, aggregateType AggregateType, aggregateID string, cause *Message[C]) *Message[T] {
	return &Message[T]{
		ID:            uuid.NewString(),
		Action:        action,
		Type:          mType,
		AggregateID:   aggregateID,
		AggregateType: aggregateType,
		Payload:       payload,
		CausationID:   cause.ID,
		Timestamp:     time.Now(),
	}
}

func NewCommand[T any, P any](action Action, payload T, aggregateType AggregateType, aggregateID string, parent *Message[P]) *Message[T] {
	return NewMessage(Command, action, payload, aggregateType, aggregateID, parent)
}

func NewDomainEvent[T any, P any](action Action, payload T, aggregateType AggregateType, aggregateID string, parent *Message[P]) *Message[T] {
	return NewMessage(DomainEvent, action, payload, aggregateType, aggregateID, parent)
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
		ID:          m.ID,
		Action:      m.Action,
		Type:        m.Type,
		AggregateID: m.AggregateID,
		CausationID: m.CausationID,
		Timestamp:   m.Timestamp,
		Payload:     m.Payload,
	}
}
