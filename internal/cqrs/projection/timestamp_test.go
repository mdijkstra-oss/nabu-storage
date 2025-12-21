package projection

import (
	"hermes-relay/internal/cqrs/commands"
	"testing"
	"time"
)

type timestampEntity struct {
	ID        string
	UpdatedAt time.Time
}

func (e timestampEntity) WithUpdatedAt(t time.Time) any { e.UpdatedAt = t; return &e }

func TestWithTimestamp(t *testing.T) {
	baseReducer := func(current *timestampEntity, event *commands.AnyMessage) *timestampEntity {
		if event.Action == "Create" {
			return &timestampEntity{ID: event.AggregateID}
		}
		if event.Action == "Update" && current != nil {
			return &timestampEntity{ID: current.ID}
		}
		if event.Action == "Noop" {
			return current
		}
		return nil
	}

	reducer := WithTimestamp(baseReducer)
	eventTime := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name            string
		current         *timestampEntity
		action          commands.Action
		expectedTime    time.Time
		expectNil       bool
		expectUnchanged bool
	}{
		{"Create sets timestamp", nil, "Create", eventTime, false, false},
		{"Update sets timestamp", &timestampEntity{ID: "1"}, "Update", eventTime, false, false},
		{"Noop returns same pointer no timestamp change", &timestampEntity{ID: "1"}, "Noop", time.Time{}, false, true},
		{"Delete returns nil", &timestampEntity{ID: "1"}, "Delete", time.Time{}, true, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			event := &commands.AnyMessage{Action: tc.action, AggregateID: "1", Timestamp: eventTime}
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

			if tc.expectUnchanged {
				if result != tc.current {
					t.Error("expected same pointer for noop")
				}
				return
			}

			if !result.UpdatedAt.Equal(tc.expectedTime) {
				t.Errorf("expected timestamp %v, got %v", tc.expectedTime, result.UpdatedAt)
			}
		})
	}
}
