package commands

import (
	"encoding/json"
	"hermes-relay/internal/lib/utils"
	"log/slog"
	"time"
)

type Action string
type MessageType string

type AggregateType string

type ActorType string

const (
	ActorTypeHuman  ActorType = "human"
	ActorTypeLLM    ActorType = "llm"
	ActorTypeSystem ActorType = "system"
)

type Actor struct {
	UserID    string    `json:"user_id"`
	ActorType ActorType `json:"actor_type"`
}

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
	Type   MessageType `json:"type" validate:"oneof=Command DomainEvent SystemEvent"`

	AggregateID   string        `json:"aggregate_id,omitempty"`
	AggregateType AggregateType `json:"aggregate_type,omitempty"`
	Payload       T             `json:"payload,omitempty"`

	Actor       Actor     `json:"actor"`
	CausationID string    `json:"causation_id,omitempty"`
	Timestamp   time.Time `json:"timestamp"`

	// Event schema versioning - currently always 1
	// When event schema changes, increment version and add upcasting logic in LoadAllEvents()
	// Example: V1 CreatedProject{Name} → V2 CreatedProject{Title, Description}
	Version int `json:"version,omitempty"`

	// Optimistic concurrency control - required for update/delete commands
	// Client must provide current entity version to prevent concurrent modification conflicts
	// Server rejects if ExpectedVersion != entity.Version (HTTP 409 Conflict)
	ExpectedVersion *int `json:"expected_version,omitempty"`
}

type AnyMessage = Message[any]

// Todo more message things for tracking etc.
// Versioning
// Parent
/*
type Event struct {
    ChunkID            string
    AggregateID   string
    Version       int
    CausationID   string  // immediate cause (command/event ChunkID)
    CorrelationID string  // workflow/request ChunkID
}
Example:

User request: correlation ChunkID = "req-123"
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

	var payload any = m.Payload
	if sourceMap, ok := payload.(map[string]any); ok {
		if err := utils.ApplyDefaultsFromMap(p, sourceMap); err != nil {
			return &utils.ValidationError{Message: err.Error()}
		}
	}

	return utils.ToValidationError(utils.Validate.Struct(p))
}

func (m Message[T]) GetID() string {
	return m.ID
}

func (m Message[T]) GetActorType() string {
	return string(m.Actor.ActorType)
}

func (m Message[T]) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("id", m.ID),
		slog.String("action", string(m.Action)),
		slog.String("type", string(m.Type)),
		slog.String("aggregateId", m.AggregateID),
		slog.String("actorUserId", m.Actor.UserID),
		slog.String("actorType", string(m.Actor.ActorType)),
		slog.String("causationId", m.CausationID),
		slog.Time("timestamp", m.Timestamp),
		slog.Int("version", m.Version),
	)
}

func NewMessage[T, C any](mType MessageType, action Action, payload T, aggregateType AggregateType, aggregateID string, actor Actor, cause *Message[C]) *Message[T] {
	var causationID = ""
	if cause != nil {
		causationID = cause.ID
	}

	return &Message[T]{
		ID:            utils.NewID(),
		Action:        action,
		Type:          mType,
		AggregateID:   aggregateID,
		AggregateType: aggregateType,
		Payload:       payload,
		Actor:         actor,
		CausationID:   causationID,
		Timestamp:     time.Now(),
		Version:       1,
	}
}

func NewCommand[T, P any](action Action, payload T, aggregateType AggregateType, aggregateID string, actor Actor, cause *Message[P]) *Message[T] {
	return NewMessage(Command, action, payload, aggregateType, aggregateID, actor, cause)
}

func NewDomainEvent[T, P any](action Action, payload T, aggregateType AggregateType, aggregateID string, actor Actor, cause *Message[P]) *Message[T] {
	return NewMessage(DomainEvent, action, payload, aggregateType, aggregateID, actor, cause)
}

func NewSystemEvent[T, P any](action Action, payload T, aggregateType AggregateType, aggregateID string, actor Actor, cause *Message[P]) *Message[T] {
	return NewMessage(SystemEvent, action, payload, aggregateType, aggregateID, actor, cause)
}

func ToDomainEvent[T any](message *Message[T], event Action, payload ...T) *Message[T] {
	var p T
	if len(payload) > 0 {
		p = payload[0] // Use provided payload
	} else {
		p = message.Payload // Use message's payload
	}

	return NewMessage(
		DomainEvent,
		event,
		p,
		message.AggregateType,
		message.AggregateID,
		message.Actor,
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
		ID:              m.ID,
		Action:          m.Action,
		Type:            m.Type,
		AggregateID:     m.AggregateID,
		AggregateType:   m.AggregateType,
		Actor:           m.Actor,
		CausationID:     m.CausationID,
		Timestamp:       m.Timestamp,
		Payload:         m.Payload,
		Version:         m.Version,
		ExpectedVersion: m.ExpectedVersion,
	}
}

// Todo: Probably fix this, but is bound to this func so its fine
// This only works for messages with a ProjectID or Project type and AggregateID!
func ExtractProjectID(message *AnyMessage) string {
	if message.AggregateType == "Project" {
		return message.AggregateID
	}

	// Try direct map access first (production case after JSON unmarshalling)
	if payload, ok := message.Payload.(map[string]any); ok {
		if projectID, ok := payload["project_id"].(string); ok {
			return projectID
		}
	}

	// Fall back to marshalling/unmarshalling for struct payloads (test case)
	// This handles cases where payload is a struct with a ProjectID field
	data, err := json.Marshal(message.Payload)
	if err == nil {
		var mapPayload map[string]any
		if err := json.Unmarshal(data, &mapPayload); err == nil {
			if projectID, ok := mapPayload["project_id"].(string); ok {
				return projectID
			}
		}
	}

	return ""
}

func IsCreatedEvent(action Action) bool {
	return len(action) >= 7 && action[:7] == "Created"
}

func IsDeletedEvent(action Action) bool {
	return len(action) >= 7 && action[:7] == "Deleted"
}

type EmptyPayload = struct{}

type PartialFailures interface {
	GetFailures() map[int]string
	WithoutFailures() any
}
