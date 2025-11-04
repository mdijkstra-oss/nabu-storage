package typedquery

import (
	"hermes-relay/internal/cqrs"
	httphandlers "hermes-relay/internal/handlers/http"
	th "hermes-relay/internal/lib/test-helpers"
	"hermes-relay/internal/projection"
	"testing"
)

type TestEntity struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Value int    `json:"value"`
}

func (t TestEntity) GetID() string {
	return t.ID
}

type MappedEntity struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

type MinValueQuery struct {
	MinValue int `path:"minValue" validate:"min=0"`
}

func getByMinValue(items []TestEntity, q MinValueQuery) []TestEntity {
	var result []TestEntity
	for _, e := range items {
		if e.Value >= q.MinValue {
			result = append(result, e)
		}
	}
	return result
}

func TestQuery(t *testing.T) {
	singleEntityStore := projection.NewStoreWithDefaults(
		func(_ *TestEntity, msg *cqrs.AnyMessage) *TestEntity { return nil },
		[]TestEntity{{ID: "test-1", Name: "Entity One", Value: 100}},
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
			processor:    Query(projection.BindQueryOne(singleEntityStore, projection.ByID[TestEntity])),
			pathParams:   map[string]string{"id": "test-1"},
			expectStatus: 200,
			expectBody:   `{"id":"test-1","name":"Entity One","value":100}`,
		},
		{
			name:         "GetByID with missing ID returns 404",
			processor:    Query(projection.BindQueryOne(emptyStore, projection.ByID[TestEntity])),
			pathParams:   map[string]string{"id": "nonexistent"},
			expectStatus: 404,
			expectBody:   `{"message":"No results found"}`,
		},
		{
			name:         "GetByID with empty ID returns 400 validation error",
			processor:    Query(projection.BindQueryOne(singleEntityStore, projection.ByID[TestEntity])),
			pathParams:   map[string]string{"id": ""},
			expectStatus: 400,
			expectBody:   `{"message":"validation failed: ID is required","fields":{"ID":"required"}}`,
		},
		{
			name:         "GetByID without path param returns 400 validation error",
			processor:    Query(projection.BindQueryOne(emptyStore, projection.ByID[TestEntity])),
			pathParams:   map[string]string{},
			expectStatus: 400,
			expectBody:   `{"message":"validation failed: ID is required","fields":{"ID":"required"}}`,
		},
		{
			name:         "GetAll with single entity returns 200",
			processor:    Query(projection.BindQuery(singleEntityStore, projection.ByAll[TestEntity])),
			pathParams:   map[string]string{},
			expectStatus: 200,
			expectBody:   `[{"id":"test-1","name":"Entity One","value":100}]`,
		},
		{
			name:         "GetAll with empty store returns 200 with empty array",
			processor:    Query(projection.BindQuery(emptyStore, projection.ByAll[TestEntity])),
			pathParams:   map[string]string{},
			expectStatus: 200,
			expectBody:   "[]",
		},
		{
			name:         "Path param binding with valid int succeeds",
			processor:    Query(projection.BindQuery(singleEntityStore, getByMinValue)),
			pathParams:   map[string]string{"minValue": "50"},
			expectStatus: 200,
			expectBody:   `[{"id":"test-1","name":"Entity One","value":100}]`,
		},
		{
			name:         "Path param binding with invalid int returns 400",
			processor:    Query(projection.BindQuery(emptyStore, getByMinValue)),
			pathParams:   map[string]string{"minValue": "not-a-number"},
			expectStatus: 400,
			expectBody:   `{"message":"invalid int value for minValue: strconv.ParseInt: parsing \"not-a-number\": invalid syntax"}`,
		},
		{
			name:         "Validation error on min constraint returns 400",
			processor:    Query(projection.BindQuery(emptyStore, getByMinValue)),
			pathParams:   map[string]string{"minValue": "-10"},
			expectStatus: 400,
			expectBody:   `{"message":"validation failed: MinValue must be at least 0 characters","fields":{"MinValue":"min"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httphandlers.Request{Path: tt.pathParams}
			response := tt.processor(request)

			th.AssertEqualSimple(t, tt.expectStatus, response.StatusCode)
			th.AssertEqualSimple(t, tt.expectBody, string(response.Body))
		})
	}
}

func TestQueryWithMap(t *testing.T) {
	filledStore := projection.NewStoreWithDefaults(
		func(_ *TestEntity, msg *cqrs.AnyMessage) *TestEntity { return nil },
		[]TestEntity{{ID: "test-1", Name: "Entity One", Value: 100}},
	)
	emptyStore := projection.NewStore(func(_ *TestEntity, msg *cqrs.AnyMessage) *TestEntity { return nil })

	mapFn := func(e TestEntity) MappedEntity {
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
			processor:    Query(projection.BindQueryOne(filledStore, projection.ThenMap(projection.ByID[TestEntity], mapFn))),
			pathParams:   map[string]string{"id": "test-1"},
			expectStatus: 200,
			expectBody:   `{"name":"Entity One","value":100}`,
		},
		{
			name:         "QueryWithMap with missing entity returns 404",
			processor:    Query(projection.BindQueryOne(emptyStore, projection.ThenMap(projection.ByID[TestEntity], mapFn))),
			pathParams:   map[string]string{"id": "nonexistent"},
			expectStatus: 404,
			expectBody:   `{"message":"No results found"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httphandlers.Request{Path: tt.pathParams}
			response := tt.processor(request)

			th.AssertEqualSimple(t, tt.expectStatus, response.StatusCode)
			th.AssertEqualSimple(t, tt.expectBody, string(response.Body))
		})
	}
}
