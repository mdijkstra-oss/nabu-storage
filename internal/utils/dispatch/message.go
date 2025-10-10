package dispatch

import (
	"encoding/json"
	"log/slog"
	"time"

	"github.com/go-playground/validator/v10"
)

type MessageType int

const (
	Command MessageType = iota
	Event
)

type Message struct {
	ID string `json:"id,omitempty"`

	Action string      `json:"action" validate:"required"`
	Type   MessageType `json:"-"`

	AggregateID string      `json:"aggregateId,omitempty"`
	Payload     interface{} `json:"payload,omitempty"`

	ParentID  string    `json:"parent,omitempty"`
	Timestamp time.Time `json:"timestamp" validate:"required"`
}

func ValidateMessage(m *Message) error {
	validate := validator.New()
	return validate.Struct(m)
}

func UnmarshalPayload[T any](action *Message, target *T) error {
	payloadBytes, _ := json.Marshal(action.Payload)
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

func MakeEvent(action string, aggregateId string, payload interface{}, parent *Message) *Message {
	return MakeMessage(Event, action, aggregateId, payload, parent)
}

func MakeCommand(action string, aggregateId string, payload interface{}, parent *Message) *Message {
	return MakeMessage(Command, action, aggregateId, payload, parent)
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
	case Event:
		return "event"
	default:
		return "unknown"
	}
}
