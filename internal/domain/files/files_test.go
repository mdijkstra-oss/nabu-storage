package files

import (
	"os"
	"path/filepath"
	"testing"

	th "nabu-storage/internal/lib/testutil"
)

func TestList(t *testing.T) {
	tests := []struct {
		Name          string
		SetupFiles    []string
		HiddenFiles   []string
		ExpectedCount int
	}{
		{
			Name:          "returns empty for no files",
			SetupFiles:    []string{},
			HiddenFiles:   []string{},
			ExpectedCount: 0,
		},
		{
			Name:          "lists visible files only",
			SetupFiles:    []string{"visible.md", "another.md"},
			HiddenFiles:   []string{".hidden"},
			ExpectedCount: 2,
		},
		{
			Name:          "excludes all hidden files",
			SetupFiles:    []string{"doc.md"},
			HiddenFiles:   []string{".git", ".env"},
			ExpectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			baseDir := t.TempDir()
			projectID := "test-project"
			dir := ProjectDir(baseDir, projectID)
			_ = os.MkdirAll(dir, 0755)

			for _, f := range tt.SetupFiles {
				_ = os.WriteFile(filepath.Join(dir, f), []byte(""), 0644)
			}
			for _, f := range tt.HiddenFiles {
				_ = os.WriteFile(filepath.Join(dir, f), []byte(""), 0644)
			}

			files, err := List(baseDir, projectID)
			th.AssertError(t, err, "", "list error")
			th.AssertEqual(t, len(files), tt.ExpectedCount, "file count")
		})
	}
}

func TestSeedRequiredFiles(t *testing.T) {
	tests := []struct {
		Name          string
		ExistingFiles map[string]string
		ExpectSeeded  []string
		ExpectKept    map[string]string
	}{
		{
			Name:          "seeds all missing files",
			ExistingFiles: map[string]string{},
			ExpectSeeded:  []string{"preferences.md", "settings.hidden.md"},
		},
		{
			Name:          "skips existing files",
			ExistingFiles: map[string]string{"preferences.md": "# My Prefs"},
			ExpectSeeded:  []string{"settings.hidden.md"},
			ExpectKept:    map[string]string{"preferences.md": "# My Prefs"},
		},
		{
			Name: "seeds nothing when all exist",
			ExistingFiles: map[string]string{
				"preferences.md":     "# Custom",
				"settings.hidden.md": "# Custom Settings",
			},
			ExpectSeeded: []string{},
			ExpectKept: map[string]string{
				"preferences.md":     "# Custom",
				"settings.hidden.md": "# Custom Settings",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			baseDir := t.TempDir()
			projectID := "test-project"

			for path, content := range tt.ExistingFiles {
				th.AssertError(t, Write(baseDir, projectID, path, content), "", "setup write")
			}

			SeedRequiredFiles(baseDir, projectID)

			for _, path := range tt.ExpectSeeded {
				th.AssertEqual(t, Exists(baseDir, projectID, path), true, "seeded "+path)
			}

			for path, expected := range tt.ExpectKept {
				content, err := Read(baseDir, projectID, path)
				th.AssertError(t, err, "", "read "+path)
				th.AssertEqual(t, content, expected, "preserved "+path)
			}
		})
	}
}

func TestSortForInitialSend(t *testing.T) {
	tests := []struct {
		Name     string
		Input    []string
		Expected []string
	}{
		{
			Name:     "empty list",
			Input:    []string{},
			Expected: []string{},
		},
		{
			Name:     "alphabetical order",
			Input:    []string{"notes.md", "codes.hidden.md", "settings.hidden.md", "journal.md"},
			Expected: []string{"codes.hidden.md", "journal.md", "notes.md", "settings.hidden.md"},
		},
		{
			Name:     "already sorted",
			Input:    []string{"a.md", "b.md"},
			Expected: []string{"a.md", "b.md"},
		},
		{
			Name:     "mixed regular and hidden interleaved",
			Input:    []string{"tags.hidden.md", "a.md", "settings.hidden.md", "searches.hidden.md"},
			Expected: []string{"a.md", "searches.hidden.md", "settings.hidden.md", "tags.hidden.md"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			result := SortForInitialSend(tt.Input)
			th.AssertEqual(t, len(result), len(tt.Expected), "length")
			for i, name := range tt.Expected {
				th.AssertEqual(t, result[i], name, "index "+string(rune('0'+i)))
			}
		})
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		Name      string
		SetupFile bool
		ExpectErr string
		FileGone  bool
	}{
		{
			Name:      "deletes existing file",
			SetupFile: true,
			ExpectErr: "",
			FileGone:  true,
		},
		{
			Name:      "fails on non-existent file",
			SetupFile: false,
			ExpectErr: "no such file",
			FileGone:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			baseDir := t.TempDir()
			projectID := "test-project"
			path := "target.md"

			if tt.SetupFile {
				_ = Write(baseDir, projectID, path, "content")
			}

			err := Delete(baseDir, projectID, path)
			th.AssertError(t, err, tt.ExpectErr, "delete error")
			th.AssertEqual(t, Exists(baseDir, projectID, path), !tt.FileGone, "file exists")
		})
	}
}
