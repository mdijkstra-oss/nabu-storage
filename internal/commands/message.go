package commands

import (
	"encoding/json"
	"log/slog"
	"time"

	"github.com/go-playground/validator/v10"
)

type MessageType string

const (
	Command     MessageType = "command"
	DomainEvent MessageType = "domainEvent"
	SystemEvent MessageType = "systemEvent"
)

type Message struct {
	ID string `json:"id,omitempty"`

	Action string      `json:"action" validate:"required"`
	Type   MessageType `json:"type"`

	AggregateID string      `json:"aggregateId,omitempty"`
	Payload     interface{} `json:"payload,omitempty"`

	ParentID  string    `json:"parent,omitempty"`
	Timestamp time.Time `json:"timestamp" validate:"required"`
}

func ValidateMessage(m *Message) error {
	validate := validator.New()
	return validate.Struct(m)
}

func UnmarshalPayload[T any](message *Message, target *T) error {
	payloadBytes, _ := json.Marshal(message.Payload)
	return json.Unmarshal(payloadBytes, target)
}

func MakeMessage(msgType MessageType, action string, aggregateId string, payload interface{}, parent *Message) *Message {
	msg := &Message{
		Action:      action,
		AggregateID: aggregateId,
		Payload:     payload,
		Type:        msgType,
		Timestamp:   time.Now(),
	}

	if parent != nil {
		msg.ParentID = parent.ID
	}

	return msg
}

func MakeDomainEvent(action string, aggregateId string, payload interface{}, parent *Message) *Message {
	return MakeMessage(DomainEvent, action, aggregateId, payload, parent)
}

func MakeCommand(action string, aggregateId string, payload interface{}, parent *Message) *Message {
	return MakeMessage(Command, action, aggregateId, payload, parent)
}

func CommandToDomainEvent(command *Message) *Message {
	// Todo: translate to thing in past
	return MakeDomainEvent("Updated", command.AggregateID, command.Payload, command)
}

func MakeSystemEvent(action string, payload interface{}, parent *Message) *Message {
	return MakeMessage(SystemEvent, action, "", payload, parent)
}

func (m Message) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("id", m.ID),
		slog.String("action", m.Action),
		slog.String("type", m.Type.String()),
		slog.String("aggregateId", m.AggregateID),
		slog.String("parentId", m.ParentID),
		slog.Time("timestamp", m.Timestamp),
	)
}

func (mt MessageType) String() string {
	switch mt {
	case Command:
		return "command"
	case DomainEvent:
		return "event"
	default:
		return "unknown"
	}
}
