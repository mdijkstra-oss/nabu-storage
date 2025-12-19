package utils

import (
	"strings"
	"testing"
)

func TestValidPrefixedID(t *testing.T) {
	tests := []struct {
		name     string
		prefix   string
		id       string
		expected bool
	}{
		{"valid project id", "project", "project_550e8400-e29b-41d4-a716-446655440000", true},
		{"valid document id", "document", "document_550e8400-e29b-41d4-a716-446655440000", true},
		{"wrong prefix", "project", "document_550e8400-e29b-41d4-a716-446655440000", false},
		{"no prefix", "project", "550e8400-e29b-41d4-a716-446655440000", false},
		{"invalid uuid", "project", "project_not-a-uuid", false},
		{"empty string", "project", "", false},
		{"prefix only", "project", "project_", false},
		{"underscore in wrong place", "project", "proj_ect_550e8400-e29b-41d4-a716-446655440000", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidPrefixedID(tt.prefix, tt.id)
			if result != tt.expected {
				t.Errorf("ValidPrefixedID(%q, %q) = %v, want %v", tt.prefix, tt.id, result, tt.expected)
			}
		})
	}
}

func TestValidateID(t *testing.T) {
	tests := []struct {
		name     string
		prefix   string
		id       string
		expected bool
	}{
		{"prefixed id with correct prefix", "project", "project_550e8400-e29b-41d4-a716-446655440000", true},
		{"prefixed id with wrong prefix", "project", "document_550e8400-e29b-41d4-a716-446655440000", false},
		{"plain uuid is valid", "project", "550e8400-e29b-41d4-a716-446655440000", true},
		{"invalid uuid", "project", "not-a-uuid", false},
		{"empty string", "project", "", false},
		{"prefix only", "project", "project_", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateID(tt.prefix, tt.id)
			if result != tt.expected {
				t.Errorf("ValidateID(%q, %q) = %v, want %v", tt.prefix, tt.id, result, tt.expected)
			}
		})
	}
}

func TestNormalizeID(t *testing.T) {
	tests := []struct {
		name     string
		prefix   string
		id       string
		expected string
	}{
		{"plain uuid gets prefixed", "project", "550e8400-e29b-41d4-a716-446655440000", "project_550e8400-e29b-41d4-a716-446655440000"},
		{"already prefixed stays same", "project", "project_550e8400-e29b-41d4-a716-446655440000", "project_550e8400-e29b-41d4-a716-446655440000"},
		{"wrong prefix stays same", "project", "document_550e8400-e29b-41d4-a716-446655440000", "document_550e8400-e29b-41d4-a716-446655440000"},
		{"empty string stays empty", "project", "", ""},
		{"invalid id stays same", "project", "not-a-uuid", "not-a-uuid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeID(tt.prefix, tt.id)
			if result != tt.expected {
				t.Errorf("NormalizeID(%q, %q) = %q, want %q", tt.prefix, tt.id, result, tt.expected)
			}
		})
	}
}

func TestNormalizeAggregateID(t *testing.T) {
	tests := []struct {
		name          string
		aggregateType string
		id            string
		expected      string
	}{
		{"Project plain uuid", "Project", "550e8400-e29b-41d4-a716-446655440000", "project_550e8400-e29b-41d4-a716-446655440000"},
		{"Document plain uuid", "Document", "550e8400-e29b-41d4-a716-446655440000", "document_550e8400-e29b-41d4-a716-446655440000"},
		{"unknown type returns unchanged", "Unknown", "550e8400-e29b-41d4-a716-446655440000", "550e8400-e29b-41d4-a716-446655440000"},
		{"already prefixed stays same", "Project", "project_550e8400-e29b-41d4-a716-446655440000", "project_550e8400-e29b-41d4-a716-446655440000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeAggregateID(tt.aggregateType, tt.id)
			if result != tt.expected {
				t.Errorf("NormalizeAggregateID(%q, %q) = %q, want %q", tt.aggregateType, tt.id, result, tt.expected)
			}
		})
	}
}

func TestValidProjectID(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		expected bool
	}{
		{"prefixed project id", "project_550e8400-e29b-41d4-a716-446655440000", true},
		{"plain uuid is valid", "550e8400-e29b-41d4-a716-446655440000", true},
		{"document prefix is invalid", "document_550e8400-e29b-41d4-a716-446655440000", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidProjectID(tt.id)
			if result != tt.expected {
				t.Errorf("ValidProjectID(%q) = %v, want %v", tt.id, result, tt.expected)
			}
		})
	}
}

func TestValidDocumentID(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		expected bool
	}{
		{"prefixed document id", "document_550e8400-e29b-41d4-a716-446655440000", true},
		{"plain uuid is valid", "550e8400-e29b-41d4-a716-446655440000", true},
		{"project prefix is invalid", "project_550e8400-e29b-41d4-a716-446655440000", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidDocumentID(tt.id)
			if result != tt.expected {
				t.Errorf("ValidDocumentID(%q) = %v, want %v", tt.id, result, tt.expected)
			}
		})
	}
}

func TestNewProjectID(t *testing.T) {
	id := NewProjectID()
	if !strings.HasPrefix(id, "project_") {
		t.Errorf("NewProjectID() = %q, want prefix 'project_'", id)
	}
	if !ValidProjectID(id) {
		t.Errorf("NewProjectID() generated invalid ID: %q", id)
	}
}

func TestNewDocumentID(t *testing.T) {
	id := NewDocumentID()
	if !strings.HasPrefix(id, "document_") {
		t.Errorf("NewDocumentID() = %q, want prefix 'document_'", id)
	}
	if !ValidDocumentID(id) {
		t.Errorf("NewDocumentID() generated invalid ID: %q", id)
	}
}
