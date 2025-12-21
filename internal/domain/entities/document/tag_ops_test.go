package document

import (
	"testing"

	th "hermes-relay/internal/lib/test-helpers"
)

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
		{"normalize current", []string{"Hello"}, []string{"hello"}, []string{"hello"}},
		{"normalize current dedup", []string{"Hello", "WORLD"}, []string{"world"}, []string{"hello", "world"}},
		{"skip empty in current", []string{"   ", "a"}, []string{"b"}, []string{"a", "b"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AddTags(tt.current, tt.add)
			th.AssertEqual(t, got, tt.want, "tags")
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
		{"remove all", []string{"a", "b"}, []string{"a", "b"}, []string{}},
		{"normalize on remove", []string{"hello"}, []string{"  HELLO  "}, []string{}},
		{"empty current", nil, []string{"a"}, []string{}},
		{"normalize current", []string{"Hello", "WORLD"}, []string{"hello"}, []string{"world"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RemoveTags(tt.current, tt.remove)
			th.AssertEqual(t, got, tt.want, "tags")
		})
	}
}
