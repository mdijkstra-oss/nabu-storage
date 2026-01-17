package document

import (
	"maps"
	"slices"
	"testing"

	th "hermes-relay/internal/lib/test-helpers"
)

func tagsFromIDs(ids []string) map[string]Tag {
	if ids == nil {
		return nil
	}
	result := make(map[string]Tag, len(ids))
	for _, id := range ids {
		result[id] = Tag{ID: id}
	}
	return result
}

func TestNormalizeTag(t *testing.T) {
	th.RunMapTests(t, map[string]string{
		"hello":           "hello",
		"  hello  ":       "hello",
		"Hello World":     "hello world",
		"  HELLO   WORLD": "hello world",
		"🧳 Travel":       "🧳 travel",
		"🔥 HOT TOPIC":    "🔥 hot topic",
		"   ":             "",
	}, NormalizeTag)
}

func TestAddTags(t *testing.T) {
	tests := []struct {
		name    string
		current []string
		add     []string
		want    []string
	}{
		{"empty to empty", nil, []string{"a"}, []string{"a"}},
		{"add single", []string{"a"}, []string{"b"}, []string{"a", "b"}},
		{"add multiple", []string{"a"}, []string{"b", "c"}, []string{"a", "b", "c"}},
		{"dedup existing", []string{"a", "b"}, []string{"b"}, []string{"a", "b"}},
		{"dedup in add", []string{"a"}, []string{"b", "b"}, []string{"a", "b"}},
		{"normalize on add", []string{"a"}, []string{"  B  "}, []string{"a", "b"}},
		{"sorted output", []string{"z"}, []string{"a", "m"}, []string{"a", "m", "z"}},
		{"skip empty", []string{"a"}, []string{"", "  "}, []string{"a"}},
		{"emoji tag", []string{}, []string{"🧳 travel"}, []string{"🧳 travel"}},
		{"normalize current", []string{"hello"}, []string{"hello"}, []string{"hello"}},
		{"skip empty in current", []string{"a"}, []string{"b"}, []string{"a", "b"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AddTags(tagsFromIDs(tt.current), tt.add)
			th.AssertEqual(t, slices.Sorted(maps.Keys(got)), tt.want, "tags")
		})
	}
}

func TestRemoveTags(t *testing.T) {
	tests := []struct {
		name    string
		current []string
		remove  []string
		want    []string
	}{
		{"remove single", []string{"a", "b", "c"}, []string{"b"}, []string{"a", "c"}},
		{"remove multiple", []string{"a", "b", "c"}, []string{"a", "c"}, []string{"b"}},
		{"remove nonexistent", []string{"a", "b"}, []string{"z"}, []string{"a", "b"}},
		{"remove all", []string{"a", "b"}, []string{"a", "b"}, nil},
		{"normalize on remove", []string{"hello"}, []string{"  HELLO  "}, nil},
		{"empty current", nil, []string{"a"}, nil},
		{"normalize current", []string{"hello", "world"}, []string{"hello"}, []string{"world"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RemoveTags(tagsFromIDs(tt.current), tt.remove)
			th.AssertEqual(t, slices.Sorted(maps.Keys(got)), tt.want, "tags")
		})
	}
}
