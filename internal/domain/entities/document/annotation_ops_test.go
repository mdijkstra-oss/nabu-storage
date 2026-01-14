package document

import (
	"testing"

	th "hermes-relay/internal/lib/test-helpers"
)

func ann(id, text, color string) Annotation {
	return Annotation{ID: id, Text: text, Actor: "test", Color: color}
}

func ptr(s string) *string {
	return &s
}

func TestAddAnnotation(t *testing.T) {
	tests := []struct {
		name    string
		current []Annotation
		add     Annotation
		want    []Annotation
	}{
		{
			name:    "add to empty",
			current: nil,
			add:     ann("a1", "text1", "amber"),
			want:    []Annotation{ann("a1", "text1", "amber")},
		},
		{
			name:    "add to existing",
			current: []Annotation{ann("a1", "text1", "amber")},
			add:     ann("a2", "text2", "blue"),
			want:    []Annotation{ann("a1", "text1", "amber"), ann("a2", "text2", "blue")},
		},
		{
			name:    "add third",
			current: []Annotation{ann("a1", "text1", "amber"), ann("a2", "text2", "blue")},
			add:     ann("a3", "text3", "green"),
			want:    []Annotation{ann("a1", "text1", "amber"), ann("a2", "text2", "blue"), ann("a3", "text3", "green")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AddAnnotation(tt.current, tt.add)
			th.AssertEqual(t, got, tt.want, "annotations")
		})
	}
}

func TestRemoveAnnotations(t *testing.T) {
	tests := []struct {
		name      string
		current   []Annotation
		removeIDs []string
		want      []Annotation
	}{
		{
			name:      "remove single",
			current:   []Annotation{ann("a1", "t1", "amber"), ann("a2", "t2", "blue")},
			removeIDs: []string{"a1"},
			want:      []Annotation{ann("a2", "t2", "blue")},
		},
		{
			name:      "remove multiple",
			current:   []Annotation{ann("a1", "t1", "amber"), ann("a2", "t2", "blue"), ann("a3", "t3", "green")},
			removeIDs: []string{"a1", "a3"},
			want:      []Annotation{ann("a2", "t2", "blue")},
		},
		{
			name:      "remove nonexistent",
			current:   []Annotation{ann("a1", "t1", "amber")},
			removeIDs: []string{"z1"},
			want:      []Annotation{ann("a1", "t1", "amber")},
		},
		{
			name:      "remove all",
			current:   []Annotation{ann("a1", "t1", "amber"), ann("a2", "t2", "blue")},
			removeIDs: []string{"a1", "a2"},
			want:      []Annotation{},
		},
		{
			name:      "empty current",
			current:   nil,
			removeIDs: []string{"a1"},
			want:      []Annotation{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RemoveAnnotations(tt.current, tt.removeIDs)
			th.AssertEqual(t, got, tt.want, "annotations")
		})
	}
}

func TestUpdateAnnotationProps(t *testing.T) {
	tests := []struct {
		name    string
		current []Annotation
		ids     []string
		props   AnnotationPropsUpdate
		want    []Annotation
	}{
		{
			name:    "update color",
			current: []Annotation{ann("a1", "t1", "amber")},
			ids:     []string{"a1"},
			props:   AnnotationPropsUpdate{Color: ptr("blue")},
			want:    []Annotation{ann("a1", "t1", "blue")},
		},
		{
			name:    "update reason",
			current: []Annotation{{ID: "a1", Text: "t1", Actor: "test", Color: "amber", Reason: "old"}},
			ids:     []string{"a1"},
			props:   AnnotationPropsUpdate{Reason: ptr("new reason")},
			want:    []Annotation{{ID: "a1", Text: "t1", Actor: "test", Color: "amber", Reason: "new reason"}},
		},
		{
			name:    "update payload",
			current: []Annotation{ann("a1", "t1", "amber")},
			ids:     []string{"a1"},
			props:   AnnotationPropsUpdate{Payload: &CodingPayload{Type: "coding", CodeID: "c1", Confidence: ConfidenceHigh}},
			want:    []Annotation{{ID: "a1", Text: "t1", Actor: "test", Color: "amber", Payload: &CodingPayload{Type: "coding", CodeID: "c1", Confidence: ConfidenceHigh}}},
		},
		{
			name:    "update multiple annotations",
			current: []Annotation{ann("a1", "t1", "amber"), ann("a2", "t2", "blue")},
			ids:     []string{"a1", "a2"},
			props:   AnnotationPropsUpdate{Color: ptr("green")},
			want:    []Annotation{ann("a1", "t1", "green"), ann("a2", "t2", "green")},
		},
		{
			name:    "update subset",
			current: []Annotation{ann("a1", "t1", "amber"), ann("a2", "t2", "blue"), ann("a3", "t3", "green")},
			ids:     []string{"a2"},
			props:   AnnotationPropsUpdate{Color: ptr("red")},
			want:    []Annotation{ann("a1", "t1", "amber"), ann("a2", "t2", "red"), ann("a3", "t3", "green")},
		},
		{
			name:    "update nonexistent",
			current: []Annotation{ann("a1", "t1", "amber")},
			ids:     []string{"z1"},
			props:   AnnotationPropsUpdate{Color: ptr("blue")},
			want:    []Annotation{ann("a1", "t1", "amber")},
		},
		{
			name:    "no props changed",
			current: []Annotation{ann("a1", "t1", "amber")},
			ids:     []string{"a1"},
			props:   AnnotationPropsUpdate{},
			want:    []Annotation{ann("a1", "t1", "amber")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UpdateAnnotationProps(tt.current, tt.ids, tt.props)
			th.AssertEqual(t, got, tt.want, "annotations")
		})
	}
}
