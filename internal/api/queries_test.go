package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"nabu-storage/internal/domain"
	th "nabu-storage/internal/lib/testutil"
)

func seedProjects(t *testing.T, count int) string {
	t.Helper()
	baseDir := t.TempDir()
	for i := range count {
		id := fmt.Sprintf("%08d-1111-4111-8111-111111111111", i)
		dir := filepath.Join(baseDir, id)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}
	return baseDir
}

func requestProjects(t *testing.T, baseDir, query string) Page[domain.Project] {
	t.Helper()
	recorder := httptest.NewRecorder()
	ProjectsHandler(baseDir)(recorder, httptest.NewRequest(http.MethodGet, "/queries/projects"+query, nil))

	th.AssertEqual(t, recorder.Code, http.StatusOK, "status")
	th.AssertEqual(t, recorder.Header().Get("Content-Type"), "application/json", "content type")

	var page Page[domain.Project]
	if err := json.Unmarshal(recorder.Body.Bytes(), &page); err != nil {
		t.Fatalf("decode body %q: %v", recorder.Body.String(), err)
	}
	return page
}

func TestProjectsHandler(t *testing.T) {
	t.Run("no projects is an empty page, not null", func(t *testing.T) {
		page := requestProjects(t, t.TempDir(), "")
		th.AssertEqual(t, page.Items, []domain.Project{}, "items")
		th.AssertEqual(t, page.Total, 0, "total")
	})

	t.Run("total counts every project, items only the page", func(t *testing.T) {
		page := requestProjects(t, seedProjects(t, 5), "?page=1&page_size=2")
		th.AssertEqual(t, len(page.Items), 2, "items")
		th.AssertEqual(t, page.Total, 5, "total")
		th.AssertEqual(t, page.PageSize, 2, "page size")
	})

	t.Run("the last page is short", func(t *testing.T) {
		page := requestProjects(t, seedProjects(t, 5), "?page=3&page_size=2")
		th.AssertEqual(t, len(page.Items), 1, "items")
		th.AssertEqual(t, page.Page, 3, "page")
	})

	t.Run("a page past the end is empty", func(t *testing.T) {
		page := requestProjects(t, seedProjects(t, 5), "?page=99&page_size=2")
		th.AssertEqual(t, page.Items, []domain.Project{}, "items")
		th.AssertEqual(t, page.Total, 5, "total")
	})

	t.Run("unparseable and non-positive parameters fall back to defaults", func(t *testing.T) {
		for _, query := range []string{"?page=x&page_size=y", "?page=0&page_size=-1", ""} {
			page := requestProjects(t, seedProjects(t, 1), query)
			th.AssertEqual(t, page.Page, 1, "page for "+query)
			th.AssertEqual(t, page.PageSize, defaultPageSize, "page size for "+query)
		}
	})

	t.Run("page size is capped", func(t *testing.T) {
		page := requestProjects(t, seedProjects(t, 1), "?page_size=100000")
		th.AssertEqual(t, page.PageSize, maxPageSize, "page size")
	})
}
