package domain

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	th "nabu-storage/internal/lib/testutil"
)

const (
	alpha = "11111111-1111-4111-8111-111111111111"
	beta  = "22222222-2222-4222-8222-222222222222"
)

func writeProject(t *testing.T, baseDir, id, name string, written time.Time) {
	t.Helper()
	dir := filepath.Join(baseDir, id)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("x"), 0644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
	if err := os.Chtimes(path, written, written); err != nil {
		t.Fatalf("chtimes %s: %v", path, err)
	}
}

func ids(projects []Project) []string {
	names := make([]string, len(projects))
	for i, p := range projects {
		names[i] = p.ID
	}
	return names
}

func TestListProjects(t *testing.T) {
	t.Run("missing baseDir lists nothing", func(t *testing.T) {
		projects, err := ListProjects(filepath.Join(t.TempDir(), "absent"))
		th.AssertError(t, err, "", "error")
		th.AssertEqual(t, len(projects), 0, "count")
	})

	t.Run("newest write comes first", func(t *testing.T) {
		baseDir := t.TempDir()
		writeProject(t, baseDir, alpha, "old.md", time.Now().Add(-time.Hour))
		writeProject(t, baseDir, beta, "new.md", time.Now())

		projects, err := ListProjects(baseDir)
		th.AssertError(t, err, "", "error")
		th.AssertEqual(t, ids(projects), []string{beta, alpha}, "order")
	})

	t.Run("rewriting a file reorders, though the directory is untouched", func(t *testing.T) {
		baseDir := t.TempDir()
		writeProject(t, baseDir, alpha, "a.md", time.Now().Add(-time.Hour))
		writeProject(t, baseDir, beta, "b.md", time.Now().Add(-2*time.Hour))

		writeProject(t, baseDir, beta, "b.md", time.Now())

		projects, err := ListProjects(baseDir)
		th.AssertError(t, err, "", "error")
		th.AssertEqual(t, ids(projects), []string{beta, alpha}, "order")
	})

	t.Run("non-UUID directories and loose files are skipped", func(t *testing.T) {
		baseDir := t.TempDir()
		writeProject(t, baseDir, alpha, "a.md", time.Now())
		if err := os.MkdirAll(filepath.Join(baseDir, "not-a-uuid"), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(baseDir, "stray.md"), []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}

		projects, err := ListProjects(baseDir)
		th.AssertError(t, err, "", "error")
		th.AssertEqual(t, ids(projects), []string{alpha}, "ids")
	})

	t.Run("a UUID spelling is normalised", func(t *testing.T) {
		baseDir := t.TempDir()
		writeProject(t, baseDir, "{"+alpha+"}", "a.md", time.Now())

		projects, err := ListProjects(baseDir)
		th.AssertError(t, err, "", "error")
		th.AssertEqual(t, ids(projects), []string{alpha}, "ids")
	})
}
