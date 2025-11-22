package projection

import (
	"hermes-relay/internal/cqrs/commands"
	"os"
	"testing"
)

type TestEntity struct {
	ID      string
	Data    map[string]string
	Items   []string
	Nested  *NestedData
	Healthy bool
}

func (e TestEntity) GetID() string { return e.ID }

type NestedData struct {
	Values []int
	Lookup map[string]int
}

func TestHasSharedState(t *testing.T) {
	tests := []struct {
		name     string
		before   any
		after    any
		expected bool
	}{
		{
			name:     "Different maps",
			before:   &TestEntity{Data: map[string]string{"a": "1"}},
			after:    &TestEntity{Data: map[string]string{"a": "1"}},
			expected: false,
		},
		{
			name: "Same map pointer",
			before: &TestEntity{
				Data: map[string]string{"a": "1"},
			},
			after:    nil,
			expected: false,
		},
		{
			name:     "Different slices",
			before:   &TestEntity{Items: []string{"a", "b"}},
			after:    &TestEntity{Items: []string{"a", "b"}},
			expected: false,
		},
		{
			name:     "Nil values",
			before:   &TestEntity{},
			after:    &TestEntity{},
			expected: false,
		},
		{
			name: "Nested different maps",
			before: &TestEntity{
				Nested: &NestedData{
					Lookup: map[string]int{"x": 1},
				},
			},
			after: &TestEntity{
				Nested: &NestedData{
					Lookup: map[string]int{"x": 1},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasSharedState(tt.before, tt.after)
			if result != tt.expected {
				t.Errorf("hasSharedState() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestHasSharedState_SharedMemory(t *testing.T) {
	sharedMap := map[string]string{"a": "1"}
	before := &TestEntity{Data: sharedMap}
	after := &TestEntity{Data: sharedMap}

	if !hasSharedState(before, after) {
		t.Error("Expected to detect shared map memory")
	}
}

func TestHasSharedState_SharedSlice(t *testing.T) {
	sharedSlice := []string{"a", "b"}
	before := &TestEntity{Items: sharedSlice}
	after := &TestEntity{Items: sharedSlice}

	if !hasSharedState(before, after) {
		t.Error("Expected to detect shared slice memory")
	}
}

func TestHasSharedState_NestedSharedMap(t *testing.T) {
	sharedMap := map[string]int{"x": 1}
	before := &TestEntity{
		Nested: &NestedData{Lookup: sharedMap},
	}
	after := &TestEntity{
		Nested: &NestedData{Lookup: sharedMap},
	}

	if !hasSharedState(before, after) {
		t.Error("Expected to detect shared nested map memory")
	}
}

func TestWithImmutabilityCheck_DevMode(t *testing.T) {
	os.Setenv("HERMES_DEV", "true")
	defer os.Unsetenv("HERMES_DEV")

	sharedMap := map[string]string{"a": "1"}

	impureReducer := func(current *TestEntity, event *commands.AnyMessage) *TestEntity {
		return &TestEntity{
			ID:   "test",
			Data: sharedMap,
		}
	}

	wrappedReducer := WithImmutabilityCheck(impureReducer)

	current := &TestEntity{ID: "old", Data: sharedMap}
	event := &commands.AnyMessage{Action: "test"}

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for shared memory, but didn't panic")
		}
	}()

	wrappedReducer(current, event)
}

func TestWithImmutabilityCheck_ProductionMode(t *testing.T) {
	os.Unsetenv("HERMES_DEV")

	sharedMap := map[string]string{"a": "1"}

	impureReducer := func(current *TestEntity, event *commands.AnyMessage) *TestEntity {
		return &TestEntity{
			ID:   "test",
			Data: sharedMap,
		}
	}

	wrappedReducer := WithImmutabilityCheck(impureReducer)

	current := &TestEntity{ID: "old", Data: sharedMap}
	event := &commands.AnyMessage{Action: "test"}

	result := wrappedReducer(current, event)

	if result == nil {
		t.Error("Expected result even with shared memory in production mode")
	}
}

func TestWithImmutabilityCheck_PureReducer(t *testing.T) {
	os.Setenv("HERMES_DEV", "true")
	defer os.Unsetenv("HERMES_DEV")

	pureReducer := func(current *TestEntity, event *commands.AnyMessage) *TestEntity {
		newData := make(map[string]string)
		for k, v := range current.Data {
			newData[k] = v
		}

		return &TestEntity{
			ID:   "new",
			Data: newData,
		}
	}

	wrappedReducer := WithImmutabilityCheck(pureReducer)

	current := &TestEntity{
		ID:   "old",
		Data: map[string]string{"a": "1"},
	}
	event := &commands.AnyMessage{Action: "test"}

	result := wrappedReducer(current, event)

	if result == nil {
		t.Error("Expected result for pure reducer")
		return
	}
	if result.ID != "new" {
		t.Errorf("Expected ID 'new', got %s", result.ID)
	}
}
