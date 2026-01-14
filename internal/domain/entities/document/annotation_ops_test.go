package document

import (
	"testing"

	th "hermes-relay/internal/lib/test-helpers"
)

func ann(id, text, color string) Annotation {
	return Annotation{ID: id, Text: text, Actor: "test", Color: color}
}

func annMap(anns ...Annotation) map[string]Annotation {
	result := make(map[string]Annotation, len(anns))
	for _, a := range anns {
		result[a.ID] = a
	}
	return result
}

func ptr(s string) *string {
	return &s
}

func TestAddAnnotation(t *testing.T) {
	tests := []struct {
		name    string
		current map[string]Annotation
		add     Annotation
		want    map[string]Annotation
	}{
		{
			name:    "add to empty",
			current: nil,
			add:     ann("a1", "text1", "amber"),
			want:    annMap(ann("a1", "text1", "amber")),
		},
		{
			name:    "add to existing",
			current: annMap(ann("a1", "text1", "amber")),
			add:     ann("a2", "text2", "blue"),
			want:    annMap(ann("a1", "text1", "amber"), ann("a2", "text2", "blue")),
		},
		{
			name:    "add third",
			current: annMap(ann("a1", "text1", "amber"), ann("a2", "text2", "blue")),
			add:     ann("a3", "text3", "green"),
			want:    annMap(ann("a1", "text1", "amber"), ann("a2", "text2", "blue"), ann("a3", "text3", "green")),
		},
		{
			name:    "skip exact duplicate text",
			current: annMap(ann("a1", "Daniel", "amber")),
			add:     ann("a2", "Daniel", "blue"),
			want:    annMap(ann("a1", "Daniel", "amber")),
		},
		{
			name:    "skip case-insensitive duplicate",
			current: annMap(ann("a1", "Daniel", "amber")),
			add:     ann("a2", "daniel", "blue"),
			want:    annMap(ann("a1", "Daniel", "amber")),
		},
		{
			name:    "skip apostrophe-normalized duplicate",
			current: annMap(ann("a1", "Daniel's", "amber")),
			add:     ann("a2", "Daniels", "blue"),
			want:    annMap(ann("a1", "Daniel's", "amber")),
		},
		{
			name:    "skip underscore-normalized duplicate",
			current: annMap(ann("a1", "hello_world", "amber")),
			add:     ann("a2", "hello world", "blue"),
			want:    annMap(ann("a1", "hello_world", "amber")),
		},
		{
			name:    "add different text",
			current: annMap(ann("a1", "Daniel", "amber")),
			add:     ann("a2", "David", "blue"),
			want:    annMap(ann("a1", "Daniel", "amber"), ann("a2", "David", "blue")),
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
		current   map[string]Annotation
		removeIDs []string
		want      map[string]Annotation
	}{
		{
			name:      "remove single",
			current:   annMap(ann("a1", "t1", "amber"), ann("a2", "t2", "blue")),
			removeIDs: []string{"a1"},
			want:      annMap(ann("a2", "t2", "blue")),
		},
		{
			name:      "remove multiple",
			current:   annMap(ann("a1", "t1", "amber"), ann("a2", "t2", "blue"), ann("a3", "t3", "green")),
			removeIDs: []string{"a1", "a3"},
			want:      annMap(ann("a2", "t2", "blue")),
		},
		{
			name:      "remove nonexistent",
			current:   annMap(ann("a1", "t1", "amber")),
			removeIDs: []string{"z1"},
			want:      annMap(ann("a1", "t1", "amber")),
		},
		{
			name:      "remove all",
			current:   annMap(ann("a1", "t1", "amber"), ann("a2", "t2", "blue")),
			removeIDs: []string{"a1", "a2"},
			want:      map[string]Annotation{},
		},
		{
			name:      "empty current",
			current:   nil,
			removeIDs: []string{"a1"},
			want:      map[string]Annotation{},
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
		current map[string]Annotation
		ids     []string
		props   AnnotationPropsUpdate
		want    map[string]Annotation
	}{
		{
			name:    "update color",
			current: annMap(ann("a1", "t1", "amber")),
			ids:     []string{"a1"},
			props:   AnnotationPropsUpdate{Color: ptr("blue")},
			want:    annMap(ann("a1", "t1", "blue")),
		},
		{
			name:    "update reason",
			current: annMap(Annotation{ID: "a1", Text: "t1", Actor: "test", Color: "amber", Reason: "old"}),
			ids:     []string{"a1"},
			props:   AnnotationPropsUpdate{Reason: ptr("new reason")},
			want:    annMap(Annotation{ID: "a1", Text: "t1", Actor: "test", Color: "amber", Reason: "new reason"}),
		},
		{
			name:    "update payload",
			current: annMap(ann("a1", "t1", "amber")),
			ids:     []string{"a1"},
			props:   AnnotationPropsUpdate{Payload: &CodingPayload{Type: "coding", CodeID: "c1", Confidence: ConfidenceHigh}},
			want:    annMap(Annotation{ID: "a1", Text: "t1", Actor: "test", Color: "amber", Payload: &CodingPayload{Type: "coding", CodeID: "c1", Confidence: ConfidenceHigh}}),
		},
		{
			name:    "update multiple annotations",
			current: annMap(ann("a1", "t1", "amber"), ann("a2", "t2", "blue")),
			ids:     []string{"a1", "a2"},
			props:   AnnotationPropsUpdate{Color: ptr("green")},
			want:    annMap(ann("a1", "t1", "green"), ann("a2", "t2", "green")),
		},
		{
			name:    "update subset",
			current: annMap(ann("a1", "t1", "amber"), ann("a2", "t2", "blue"), ann("a3", "t3", "green")),
			ids:     []string{"a2"},
			props:   AnnotationPropsUpdate{Color: ptr("red")},
			want:    annMap(ann("a1", "t1", "amber"), ann("a2", "t2", "red"), ann("a3", "t3", "green")),
		},
		{
			name:    "update nonexistent",
			current: annMap(ann("a1", "t1", "amber")),
			ids:     []string{"z1"},
			props:   AnnotationPropsUpdate{Color: ptr("blue")},
			want:    annMap(ann("a1", "t1", "amber")),
		},
		{
			name:    "no props changed",
			current: annMap(ann("a1", "t1", "amber")),
			ids:     []string{"a1"},
			props:   AnnotationPropsUpdate{},
			want:    annMap(ann("a1", "t1", "amber")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UpdateAnnotationProps(tt.current, tt.ids, tt.props)
			th.AssertEqual(t, got, tt.want, "annotations")
		})
	}
}
