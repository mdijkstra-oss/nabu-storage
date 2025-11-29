package projection

import (
	th "hermes-relay/internal/lib/test-helpers"
	"testing"
)

type testItem struct {
	ID        string
	ActorType string
}

func (t testItem) GetID() string {
	return t.ID
}

func makeItems(ids ...string) []testItem {
	items := make([]testItem, len(ids))
	for i, id := range ids {
		items[i] = testItem{ID: id, ActorType: "human"}
	}
	return items
}

func makeItemsWithActor(pairs ...string) []testItem {
	items := make([]testItem, len(pairs)/2)
	for i := 0; i < len(pairs); i += 2 {
		items[i/2] = testItem{ID: pairs[i], ActorType: pairs[i+1]}
	}
	return items
}

func getActorType(t testItem) string {
	return t.ActorType
}

func TestCursorFilter(t *testing.T) {
	tests := []struct {
		Name     string
		Input    cursorInput
		Expected CursorResult[testItem]
	}{
		{
			Name: "returns all items when no filters",
			Input: cursorInput{
				items: makeItems("a", "b", "c"),
				query: CursorQuery{Limit: 20},
			},
			Expected: CursorResult[testItem]{
				Items:      makeItems("a", "b", "c"),
				NextCursor: "c",
				HasMore:    false,
			},
		},
		{
			Name: "returns items after since_id",
			Input: cursorInput{
				items: makeItems("a", "b", "c", "d"),
				query: CursorQuery{SinceID: "b", Limit: 20},
			},
			Expected: CursorResult[testItem]{
				Items:      makeItems("c", "d"),
				NextCursor: "d",
				HasMore:    false,
			},
		},
		{
			Name: "respects limit",
			Input: cursorInput{
				items: makeItems("a", "b", "c", "d", "e"),
				query: CursorQuery{Limit: 2},
			},
			Expected: CursorResult[testItem]{
				Items:      makeItems("a", "b"),
				NextCursor: "b",
				HasMore:    true,
			},
		},
		{
			Name: "filters by actor_type",
			Input: cursorInput{
				items: makeItemsWithActor("a", "human", "b", "llm", "c", "human", "d", "llm"),
				query: CursorQuery{ActorType: "human", Limit: 20},
			},
			Expected: CursorResult[testItem]{
				Items:      makeItemsWithActor("a", "human", "c", "human"),
				NextCursor: "c",
				HasMore:    false,
			},
		},
		{
			Name: "combines since_id and actor_type filters",
			Input: cursorInput{
				items: makeItemsWithActor("a", "human", "b", "llm", "c", "human", "d", "human"),
				query: CursorQuery{SinceID: "a", ActorType: "human", Limit: 20},
			},
			Expected: CursorResult[testItem]{
				Items:      makeItemsWithActor("c", "human", "d", "human"),
				NextCursor: "d",
				HasMore:    false,
			},
		},
		{
			Name: "empty result returns last item id as cursor",
			Input: cursorInput{
				items: makeItems("a", "b", "c"),
				query: CursorQuery{SinceID: "c", Limit: 20},
			},
			Expected: CursorResult[testItem]{
				Items:      []testItem{},
				NextCursor: "c",
				HasMore:    false,
			},
		},
		{
			Name: "invalid since_id returns all items",
			Input: cursorInput{
				items: makeItems("a", "b", "c"),
				query: CursorQuery{SinceID: "invalid", Limit: 20},
			},
			Expected: CursorResult[testItem]{
				Items:      makeItems("a", "b", "c"),
				NextCursor: "c",
				HasMore:    false,
			},
		},
		{
			Name: "empty items returns since_id as cursor",
			Input: cursorInput{
				items: []testItem{},
				query: CursorQuery{SinceID: "last-known", Limit: 20},
			},
			Expected: CursorResult[testItem]{
				Items:      []testItem{},
				NextCursor: "last-known",
				HasMore:    false,
			},
		},
	}

	th.RunFunctionTests(t, tests, func(input cursorInput) CursorResult[testItem] {
		return CursorFilter(input.items, input.query, getActorType)
	})
}

type cursorInput struct {
	items []testItem
	query CursorQuery
}
