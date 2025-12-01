package projection

import (
	"hermes-relay/internal/cqrs/commands"
	"testing"
)

type testEntity struct {
	ID      string
	Version int
}

func (e testEntity) GetVersion() int      { return e.Version }
func (e testEntity) WithVersion(v int) any { e.Version = v; return &e }

func TestWithVersionIncrement(t *testing.T) {
	baseReducer := func(current *testEntity, event *commands.AnyMessage) *testEntity {
		if event.Action == "Create" {
			return &testEntity{ID: event.AggregateID}
		}
		if event.Action == "Update" && current != nil {
			return &testEntity{ID: current.ID}
		}
		if event.Action == "Noop" {
			return current
		}
		return nil
	}

	reducer := WithVersionIncrement(baseReducer)

	tests := []struct {
		name            string
		current         *testEntity
		action          commands.Action
		expectedVersion int
		expectNil       bool
	}{
		{"Create sets version 1", nil, "Create", 1, false},
		{"Update increments version", &testEntity{ID: "1", Version: 1}, "Update", 2, false},
		{"Update from version 5", &testEntity{ID: "1", Version: 5}, "Update", 6, false},
		{"Noop returns same pointer no increment", &testEntity{ID: "1", Version: 1}, "Noop", 1, false},
		{"Delete returns nil", &testEntity{ID: "1", Version: 1}, "Delete", 0, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			event := &commands.AnyMessage{Action: tc.action, AggregateID: "1"}
			result := reducer(tc.current, event)

			if tc.expectNil {
				if result != nil {
					t.Errorf("expected nil, got %+v", result)
				}
				return
			}

			if result == nil {
				t.Fatal("expected non-nil result")
			}
			if result.Version != tc.expectedVersion {
				t.Errorf("expected version %d, got %d", tc.expectedVersion, result.Version)
			}
		})
	}
}
