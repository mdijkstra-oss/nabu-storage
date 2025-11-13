package projection_test

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/projection"
	"hermes-relay/internal/lib/test-helpers/domain-helpers"
	"hermes-relay/internal/lib/test-helpers/reducer-helpers"
	"testing"
)

type HealthTestEntity struct {
	ID      string
	Healthy bool
	Value   string
}

func (e *HealthTestEntity) MarkUnhealthy() { e.Healthy = false }
func (e HealthTestEntity) IsHealthy() bool { return e.Healthy }
func (e HealthTestEntity) GetID() string   { return e.ID }

var testReducer = projection.CombineReducers(
	projection.For("Create", createReducer),
	projection.IfExists(
		projection.For("Update", updateReducer),
		projection.For("Panic", panicReducer),
	),
)

func createReducer(_ *HealthTestEntity, message *commands.AnyMessage, _ any) *HealthTestEntity {
	return &HealthTestEntity{ID: message.AggregateID, Healthy: true, Value: "created"}
}

func updateReducer(current *HealthTestEntity, _ *commands.AnyMessage, _ any) *HealthTestEntity {
	current.Value = "updated"
	return current
}

func panicReducer(_ *HealthTestEntity, _ *commands.AnyMessage, _ any) *HealthTestEntity {
	panic("intentional panic for testing")
}

func TestWithHealthCheck(t *testing.T) {
	wrappedReducer := projection.WithHealthCheck[HealthTestEntity, *HealthTestEntity](testReducer)

	tests := []reducer_helpers.ReducerTestCase[*HealthTestEntity]{
		{
			Name:     "Normal event on healthy entity keeps it healthy",
			Initial:  &HealthTestEntity{ID: "entity-1", Healthy: true, Value: "initial"},
			Event:    domain_helpers.NewDomainEvent("TestEntity", "entity-1", "Update", nil),
			Expected: &HealthTestEntity{ID: "entity-1", Healthy: true, Value: "updated"},
		},
		{
			Name:     "Panic event on healthy entity marks it unhealthy",
			Initial:  &HealthTestEntity{ID: "entity-1", Healthy: true, Value: "initial"},
			Event:    domain_helpers.NewDomainEvent("TestEntity", "entity-1", "Panic", nil),
			Expected: &HealthTestEntity{ID: "entity-1", Healthy: false, Value: "initial"},
		},
		{
			Name:     "Create event on nil entity creates healthy entity",
			Initial:  nil,
			Event:    domain_helpers.NewDomainEvent("TestEntity", "entity-1", "Create", nil),
			Expected: &HealthTestEntity{ID: "entity-1", Healthy: true, Value: "created"},
		},
		{
			Name:     "Panic event on nil entity returns nil without propagating panic",
			Initial:  nil,
			Event:    domain_helpers.NewDomainEvent("TestEntity", "entity-1", "Panic", nil),
			Expected: nil,
		},
	}

	reducer_helpers.RunReducerTests(t, tests, wrappedReducer)
}
