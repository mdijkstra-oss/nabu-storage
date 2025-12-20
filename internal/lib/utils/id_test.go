package utils

import (
	"testing"
)

func TestValidPrefixedID(t *testing.T) {
	type input struct {
		prefix string
		id     string
	}
	cases := map[input]bool{
		{"project", "project_550e8400-e29b-41d4-a716-446655440000"}:       true,
		{"project", "document_550e8400-e29b-41d4-a716-446655440000"}:      false,
		{"project", "550e8400-e29b-41d4-a716-446655440000"}:               false,
		{"project", "project_not-a-uuid"}:                                 false,
		{"project", ""}:                                                   false,
		{"project", "project_"}:                                           false,
		{"project", "proj_ect_550e8400-e29b-41d4-a716-446655440000"}:      false,
		{"annotation", "annotation_550e8400-e29b-41d4-a716-446655440000"}: true,
		{"annotation", "code_550e8400-e29b-41d4-a716-446655440000"}:       false,
	}
	for in, expected := range cases {
		if got := ValidPrefixedID(in.prefix, in.id); got != expected {
			t.Errorf("ValidPrefixedID(%q, %q) = %v, want %v", in.prefix, in.id, got, expected)
		}
	}
}

func TestValidateID(t *testing.T) {
	type input struct {
		prefix string
		id     string
	}
	cases := map[input]bool{
		{"project", "project_550e8400-e29b-41d4-a716-446655440000"}:  true,
		{"project", "550e8400-e29b-41d4-a716-446655440000"}:          true,
		{"project", "document_550e8400-e29b-41d4-a716-446655440000"}: false,
		{"project", "not-a-uuid"}:                                    false,
		{"project", ""}:                                              false,
		{"project", "project_"}:                                      false,
	}
	for in, expected := range cases {
		if got := ValidateID(in.prefix, in.id); got != expected {
			t.Errorf("ValidateID(%q, %q) = %v, want %v", in.prefix, in.id, got, expected)
		}
	}
}

func TestNormalizeID(t *testing.T) {
	type input struct {
		prefix string
		id     string
	}
	cases := map[input]string{
		{"project", "550e8400-e29b-41d4-a716-446655440000"}:          "project_550e8400-e29b-41d4-a716-446655440000",
		{"project", "project_550e8400-e29b-41d4-a716-446655440000"}:  "project_550e8400-e29b-41d4-a716-446655440000",
		{"project", "document_550e8400-e29b-41d4-a716-446655440000"}: "document_550e8400-e29b-41d4-a716-446655440000",
		{"project", ""}:         "",
		{"project", "not-uuid"}: "not-uuid",
	}
	for in, expected := range cases {
		if got := NormalizeID(in.prefix, in.id); got != expected {
			t.Errorf("NormalizeID(%q, %q) = %q, want %q", in.prefix, in.id, got, expected)
		}
	}
}

func TestNormalizeAggregateID(t *testing.T) {
	type input struct {
		aggregateType string
		id            string
	}
	cases := map[input]string{
		{"Project", "550e8400-e29b-41d4-a716-446655440000"}:         "project_550e8400-e29b-41d4-a716-446655440000",
		{"Document", "550e8400-e29b-41d4-a716-446655440000"}:        "document_550e8400-e29b-41d4-a716-446655440000",
		{"Annotation", "550e8400-e29b-41d4-a716-446655440000"}:      "annotation_550e8400-e29b-41d4-a716-446655440000",
		{"Code", "550e8400-e29b-41d4-a716-446655440000"}:            "code_550e8400-e29b-41d4-a716-446655440000",
		{"Unknown", "550e8400-e29b-41d4-a716-446655440000"}:         "550e8400-e29b-41d4-a716-446655440000",
		{"Project", "project_550e8400-e29b-41d4-a716-446655440000"}: "project_550e8400-e29b-41d4-a716-446655440000",
	}
	for in, expected := range cases {
		if got := NormalizeAggregateID(in.aggregateType, in.id); got != expected {
			t.Errorf("NormalizeAggregateID(%q, %q) = %q, want %q", in.aggregateType, in.id, got, expected)
		}
	}
}

func TestDeterministicBlockID(t *testing.T) {
	id1 := DeterministicBlockID("doc_123", 0)
	id2 := DeterministicBlockID("doc_123", 0)
	id3 := DeterministicBlockID("doc_123", 1)
	id4 := DeterministicBlockID("doc_456", 0)

	if id1 != id2 {
		t.Errorf("same input should produce same output: %q != %q", id1, id2)
	}
	if id1 == id3 {
		t.Errorf("different index should produce different output")
	}
	if id1 == id4 {
		t.Errorf("different doc should produce different output")
	}
	if id1[:6] != "block_" {
		t.Errorf("should have block_ prefix, got %q", id1[:6])
	}
}
