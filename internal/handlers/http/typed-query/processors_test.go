package typedquery

import (
	"context"
	"hermes-relay/internal/cqrs"
	httphandlers "hermes-relay/internal/handlers/http"
	th "hermes-relay/internal/lib/test-helpers"
	"hermes-relay/internal/projection"
	"testing"
)

// Test entity
type TestEntity struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Value int    `json:"value"`
}

// Test mapped response
type MappedEntity struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

// Custom query for testing validation
type MinValueQuery struct {
	MinValue int `path:"minValue" validate:"min=0"`
}

func getByMinValue(ctx context.Context, store *projection.Store[TestEntity], q MinValueQuery) ([]TestEntity, error) {
	all := store.GetAll()
	var result []TestEntity
	for _, e := range all {
		if e.Value >= q.MinValue {
			result = append(result, e)
		}
	}
	return result, nil
}

func TestQuery(t *testing.T) {
	singleEntityStore := projection.NewStoreWithDefaults(
		func(_ *TestEntity, msg *cqrs.AnyMessage) *TestEntity { return nil },
		map[string]TestEntity{"test-1": {ID: "test-1", Name: "Entity One", Value: 100}},
	)
	emptyStore := projection.NewStore(func(_ *TestEntity, msg *cqrs.AnyMessage) *TestEntity { return nil })

	tests := []struct {
		name         string
		processor    ProcessorFunc
		pathParams   map[string]string
		expectStatus int
		expectBody   string
	}{
		{
			name:         "GetByID with valid ID returns 200",
			processor:    Query(singleEntityStore, GetById[TestEntity]),
			pathParams:   map[string]string{"id": "test-1"},
			expectStatus: 200,
			expectBody:   `{"id":"test-1","name":"Entity One","value":100}`,
		},
		{
			name:         "GetByID with missing ID returns 404",
			processor:    Query(emptyStore, GetById[TestEntity]),
			pathParams:   map[string]string{"id": "nonexistent"},
			expectStatus: 404,
			expectBody:   `{"message":"not found: nonexistent"}`,
		},
		{
			name:         "GetByID with empty ID returns 400 validation error",
			processor:    Query(singleEntityStore, GetById[TestEntity]),
			pathParams:   map[string]string{"id": ""},
			expectStatus: 400,
			expectBody:   `{"message":"Key: 'GetByIDQuery.ID' Error:Field validation for 'ID' failed on the 'required' tag"}`,
		},
		{
			name:         "GetByID without path param returns 400 validation error",
			processor:    Query(emptyStore, GetById[TestEntity]),
			pathParams:   map[string]string{},
			expectStatus: 400,
			expectBody:   `{"message":"Key: 'GetByIDQuery.ID' Error:Field validation for 'ID' failed on the 'required' tag"}`,
		},
		{
			name:         "GetAll with single entity returns 200",
			processor:    Query(singleEntityStore, GetAll[TestEntity]),
			pathParams:   map[string]string{},
			expectStatus: 200,
			expectBody:   `[{"id":"test-1","name":"Entity One","value":100}]`,
		},
		{
			name:         "GetAll with empty store returns 200 with empty array",
			processor:    Query(emptyStore, GetAll[TestEntity]),
			pathParams:   map[string]string{},
			expectStatus: 200,
			expectBody:   "[]",
		},
		{
			name:         "Path param binding with valid int succeeds",
			processor:    Query(singleEntityStore, getByMinValue),
			pathParams:   map[string]string{"minValue": "50"},
			expectStatus: 200,
			expectBody:   `[{"id":"test-1","name":"Entity One","value":100}]`,
		},
		{
			name:         "Path param binding with invalid int returns 400",
			processor:    Query(emptyStore, getByMinValue),
			pathParams:   map[string]string{"minValue": "not-a-number"},
			expectStatus: 400,
			expectBody:   `{"message":"invalid int value for minValue: strconv.ParseInt: parsing \"not-a-number\": invalid syntax"}`,
		},
		{
			name:         "Validation error on min constraint returns 400",
			processor:    Query(emptyStore, getByMinValue),
			pathParams:   map[string]string{"minValue": "-10"},
			expectStatus: 400,
			expectBody:   `{"message":"Key: 'MinValueQuery.MinValue' Error:Field validation for 'MinValue' failed on the 'min' tag"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httphandlers.Request{Path: tt.pathParams}
			response := tt.processor(context.Background(), request)

			th.AssertEqualSimple(t, tt.expectStatus, response.StatusCode)
			th.AssertEqualSimple(t, tt.expectBody, string(response.Body))
		})
	}
}

func TestQueryWithMap(t *testing.T) {
	filledStore := projection.NewStoreWithDefaults(
		func(_ *TestEntity, msg *cqrs.AnyMessage) *TestEntity { return nil },
		map[string]TestEntity{"test-1": {ID: "test-1", Name: "Entity One", Value: 100}},
	)
	emptyStore := projection.NewStore(func(_ *TestEntity, msg *cqrs.AnyMessage) *TestEntity { return nil })

	mapFn := func(e *TestEntity) MappedEntity {
		return MappedEntity{Name: e.Name, Value: e.Value}
	}

	tests := []struct {
		name         string
		processor    ProcessorFunc
		pathParams   map[string]string
		expectStatus int
		expectBody   string
	}{
		{
			name:         "QueryWithMap successfully maps entity",
			processor:    QueryWithMap(filledStore, GetById[TestEntity], mapFn),
			pathParams:   map[string]string{"id": "test-1"},
			expectStatus: 200,
			expectBody:   `{"name":"Entity One","value":100}`,
		},
		{
			name:         "QueryWithMap with missing entity returns 404",
			processor:    QueryWithMap(emptyStore, GetById[TestEntity], mapFn),
			pathParams:   map[string]string{"id": "nonexistent"},
			expectStatus: 404,
			expectBody:   `{"message":"not found: nonexistent"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httphandlers.Request{Path: tt.pathParams}
			response := tt.processor(context.Background(), request)

			th.AssertEqualSimple(t, tt.expectStatus, response.StatusCode)
			th.AssertEqualSimple(t, tt.expectBody, string(response.Body))
		})
	}
}

