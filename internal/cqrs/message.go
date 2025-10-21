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

// Todo: Important, make sure some of these cannot be modified by external party!
type Message struct {
	ID string `json:"id,omitempty"`

	Action Action      `json:"action" validate:"required"`
	Type   MessageType `json:"type"` // todo validate etc

	AggregateID   string        `json:"aggregateId,omitempty"`
	AggregateType AggregateType `json:"aggregateType,omitempty"`
	Payload       any           `json:"payload,omitempty"`

	ParentID string `json:"parent,omitempty"`
	// Todo: add again
	Timestamp time.Time //`json:"timestamp" validate:"required"`
}

func ValidateMessage(m *Message) error {
	validate := validator.New()
	return validate.Struct(m)
}

func EnsureValidPayload(m *Message, p any) error {
	validate := validator.New()

	err := UnmarshallPayload(m, p)
	if err != nil {
		return err
	}

	return validate.Struct(p)
}

func (m Message) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("id", m.ID),
		slog.String("action", string(m.Action)),
		slog.String("type", string(m.Type)),
		slog.String("aggregateId", m.AggregateID),
		slog.String("parentId", m.ParentID),
		slog.Time("timestamp", m.Timestamp),
	)
}

func NewMessage(mType MessageType, action Action, payload any, aggregateType AggregateType, aggregateID string, parent *Message) *Message {
	return &Message{
		ID:            uuid.NewString(),
		Action:        action,
		Type:          mType,
		AggregateID:   aggregateID,
		AggregateType: aggregateType,
		Payload:       payload,
		ParentID:      parent.ID,
		Timestamp:     time.Now(),
	}
}

func NewCommand(action Action, payload any, aggregateType AggregateType, aggregateID string, parent *Message) *Message {
	return NewMessage(Command, action, payload, aggregateType, aggregateID, parent)
}

func NewDomainEvent(action Action, payload any, aggregateType AggregateType, aggregateID string, parent *Message) *Message {
	return NewMessage(DomainEvent, action, payload, aggregateType, aggregateID, parent)
}

func ToDomainEvent(message *Message, event Action) *Message {
	return NewMessage(
		DomainEvent,
		event,
		message.Payload,
		message.AggregateType,
		message.AggregateID,
		message,
	)
}

func UnmarshallPayload(m *Message, payload any) error {
	data, err := json.Marshal(m.Payload)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, payload)
}
