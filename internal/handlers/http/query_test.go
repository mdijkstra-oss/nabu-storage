package http

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

type TestEmbeddedFilter struct {
	SearchText string `query:"searchText"`
	MinValue   *int   `query:"minValue"`
}

type TestQuery struct {
	TestEmbeddedFilter
	ID   string `path:"id" validate:"required"`
	Page int    `query:"page" default:"1"`
}

func TestParseQuery_WithEmbeddedStruct(t *testing.T) {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "test-123")

	req := httptest.NewRequest("GET", "/test/test-123?searchText=hello&page=5", nil)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	query, err := ParseQuery[TestQuery](req)
	if err != nil {
		t.Fatalf("ParseQuery failed: %v", err)
	}

	if query.ID != "test-123" {
		t.Errorf("expected ID 'test-123', got '%s'", query.ID)
	}

	if query.SearchText != "hello" {
		t.Errorf("expected SearchText 'hello', got '%s'", query.SearchText)
	}

	if query.Page != 5 {
		t.Errorf("expected Page 5, got %d", query.Page)
	}
}

func TestParseQuery_AppliesDefaults(t *testing.T) {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "abc")

	req := httptest.NewRequest("GET", "/test", nil)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	query, err := ParseQuery[TestQuery](req)
	if err != nil {
		t.Fatalf("ParseQuery failed: %v", err)
	}

	if query.Page != 1 {
		t.Errorf("expected default Page 1, got %d", query.Page)
	}
}

func TestParseQuery_ValidationFails(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chi.NewRouteContext()))

	_, err := ParseQuery[TestQuery](req)
	if err == nil {
		t.Fatal("expected validation error for missing required ID")
	}
}
