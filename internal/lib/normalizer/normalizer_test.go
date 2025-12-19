package normalizer

import (
	test_helpers "hermes-relay/internal/lib/test-helpers"
	"testing"
)

func TestNormalize(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		expect    any
		expectErr string
	}{
		{
			name: "trim spaces",
			input: &struct {
				Name string `normalize:"trim"`
			}{Name: "  John Doe  "},
			expect: &struct {
				Name string `normalize:"trim"`
			}{Name: "John Doe"},
		},
		{
			name: "collapse spaces",
			input: &struct {
				Name string `normalize:"collapse"`
			}{Name: "John    Doe"},
			expect: &struct {
				Name string `normalize:"collapse"`
			}{Name: "John Doe"},
		},
		{
			name: "kebab case",
			input: &struct {
				Slug string `normalize:"kebab"`
			}{Slug: "Hello World!"},
			expect: &struct {
				Slug string `normalize:"kebab"`
			}{Slug: "hello-world"},
		},
		{
			name: "chained normalizers",
			input: &struct {
				Slug string `normalize:"trim,kebab"`
			}{Slug: "  Hello World!  "},
			expect: &struct {
				Slug string `normalize:"trim,kebab"`
			}{Slug: "hello-world"},
		},
		{
			name: "skip empty values",
			input: &struct {
				Name string `normalize:"trim"`
			}{Name: ""},
			expect: &struct {
				Name string `normalize:"trim"`
			}{Name: ""},
		},
		{
			name:   "no normalize tag ignored",
			input:  &struct{ Name string }{Name: "  John Doe  "},
			expect: &struct{ Name string }{Name: "  John Doe  "},
		},
		{
			name: "nested struct",
			input: &struct {
				Name    string `normalize:"trim"`
				Address struct {
					City string `normalize:"trim,kebab"`
				}
			}{
				Name: "  John  ",
				Address: struct {
					City string `normalize:"trim,kebab"`
				}{City: "  New York  "},
			},
			expect: &struct {
				Name    string `normalize:"trim"`
				Address struct {
					City string `normalize:"trim,kebab"`
				}
			}{
				Name: "John",
				Address: struct {
					City string `normalize:"trim,kebab"`
				}{City: "new-york"},
			},
		},
		{
			name: "pointer field",
			input: &struct {
				Name    string `normalize:"trim"`
				Address *struct {
					City string `normalize:"trim"`
				}
			}{
				Name: "  John  ",
				Address: &struct {
					City string `normalize:"trim"`
				}{City: "  Amsterdam  "},
			},
			expect: &struct {
				Name    string `normalize:"trim"`
				Address *struct {
					City string `normalize:"trim"`
				}
			}{
				Name: "John",
				Address: &struct {
					City string `normalize:"trim"`
				}{City: "Amsterdam"},
			},
		},
		{
			name: "nil pointer skipped",
			input: &struct {
				Name    string `normalize:"trim"`
				Address *struct {
					City string `normalize:"trim"`
				}
			}{Name: "  John  ", Address: nil},
			expect: &struct {
				Name    string `normalize:"trim"`
				Address *struct {
					City string `normalize:"trim"`
				}
			}{Name: "John", Address: nil},
		},
		{
			name: "slice of structs",
			input: &struct {
				Tags []struct {
					Name string `normalize:"trim,kebab"`
				}
			}{
				Tags: []struct {
					Name string `normalize:"trim,kebab"`
				}{
					{Name: "  Go Lang  "},
					{Name: "  Testing 123  "},
				},
			},
			expect: &struct {
				Tags []struct {
					Name string `normalize:"trim,kebab"`
				}
			}{
				Tags: []struct {
					Name string `normalize:"trim,kebab"`
				}{
					{Name: "go-lang"},
					{Name: "testing-123"},
				},
			},
		},
		{
			name:      "non-pointer error",
			input:     struct{ Name string }{},
			expectErr: "must pass pointer to struct",
		},
		{
			name: "unknown normalizer error",
			input: &struct {
				Name string `normalize:"unknown"`
			}{Name: "test"},
			expectErr: "unknown normalizer: unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Normalize(tt.input)

			test_helpers.AssertError(t, err, tt.expectErr, "error")
			if tt.expectErr == "" {
				test_helpers.AssertEqual(t, tt.input, tt.expect, "normalized result")
			}
		})
	}
}

func TestNormalizeValue(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		normalizers []normFunc
		expect      string
	}{
		{
			name:        "trim",
			input:       "  hello  ",
			normalizers: []normFunc{Trim},
			expect:      "hello",
		},
		{
			name:        "chained",
			input:       "  Hello    World!  ",
			normalizers: []normFunc{Trim, Collapse, Kebab},
			expect:      "hello-world",
		},
		{
			name:        "no normalizers",
			input:       "  hello  ",
			normalizers: []normFunc{},
			expect:      "  hello  ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeValue(tt.input, tt.normalizers...)
			test_helpers.AssertEqual(t, result, tt.expect, "normalized value")
		})
	}
}

func TestNormalizeIDFields(t *testing.T) {
	tests := []struct {
		name   string
		input  any
		expect any
	}{
		{
			name: "project_id normalizes plain uuid",
			input: &struct {
				ProjectID string `normalize:"project_id"`
			}{ProjectID: "550e8400-e29b-41d4-a716-446655440000"},
			expect: &struct {
				ProjectID string `normalize:"project_id"`
			}{ProjectID: "project_550e8400-e29b-41d4-a716-446655440000"},
		},
		{
			name: "project_id keeps already prefixed",
			input: &struct {
				ProjectID string `normalize:"project_id"`
			}{ProjectID: "project_550e8400-e29b-41d4-a716-446655440000"},
			expect: &struct {
				ProjectID string `normalize:"project_id"`
			}{ProjectID: "project_550e8400-e29b-41d4-a716-446655440000"},
		},
		{
			name: "document_id normalizes plain uuid",
			input: &struct {
				DocumentID string `normalize:"document_id"`
			}{DocumentID: "550e8400-e29b-41d4-a716-446655440000"},
			expect: &struct {
				DocumentID string `normalize:"document_id"`
			}{DocumentID: "document_550e8400-e29b-41d4-a716-446655440000"},
		},
		{
			name: "document_id keeps already prefixed",
			input: &struct {
				DocumentID string `normalize:"document_id"`
			}{DocumentID: "document_550e8400-e29b-41d4-a716-446655440000"},
			expect: &struct {
				DocumentID string `normalize:"document_id"`
			}{DocumentID: "document_550e8400-e29b-41d4-a716-446655440000"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Normalize(tt.input)
			test_helpers.AssertError(t, err, "", "error")
			test_helpers.AssertEqual(t, tt.input, tt.expect, "normalized result")
		})
	}
}
