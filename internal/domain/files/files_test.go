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
		CreateProject bool
		ExistingFiles map[string]string
		ExpectSeeded  []string
		ExpectMissing []string
		ExpectKept    map[string]string
	}{
		{
			Name:          "seeds all missing files",
			CreateProject: true,
			ExistingFiles: map[string]string{},
			ExpectSeeded:  []string{"preferences.md", "settings.hidden.md"},
		},
		{
			Name:          "skips existing files",
			CreateProject: true,
			ExistingFiles: map[string]string{"preferences.md": "# My Prefs"},
			ExpectSeeded:  []string{"settings.hidden.md"},
			ExpectKept:    map[string]string{"preferences.md": "# My Prefs"},
		},
		{
			Name:          "seeds nothing when all exist",
			CreateProject: true,
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
		{
			Name:          "creates nothing for an unknown project",
			CreateProject: false,
			ExistingFiles: map[string]string{},
			ExpectMissing: []string{"preferences.md", "settings.hidden.md"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			baseDir := t.TempDir()
			projectID := "test-project"

			if tt.CreateProject {
				th.AssertError(t, EnsureProjectDir(baseDir, projectID), "", "setup project dir")
			}

			for path, content := range tt.ExistingFiles {
				th.AssertError(t, Write(baseDir, projectID, path, content), "", "setup write")
			}

			SeedRequiredFiles(baseDir, projectID)

			th.AssertEqual(t, ProjectExists(baseDir, projectID), tt.CreateProject, "project directory")

			for _, path := range tt.ExpectSeeded {
				th.AssertEqual(t, Exists(baseDir, projectID, path), true, "seeded "+path)
			}

			for _, path := range tt.ExpectMissing {
				th.AssertEqual(t, Exists(baseDir, projectID, path), false, "not seeded "+path)
			}

			for path, expected := range tt.ExpectKept {
				content, err := Read(baseDir, projectID, path)
				th.AssertError(t, err, "", "read "+path)
				th.AssertEqual(t, content, expected, "preserved "+path)
			}
		})
	}
}

func TestListReturnsSortedNames(t *testing.T) {
	baseDir := t.TempDir()
	projectID := "test-project"

	for _, name := range []string{"notes.md", "a.md", "settings.hidden.md", "journal.md"} {
		th.AssertError(t, Write(baseDir, projectID, name, ""), "", "setup write")
	}

	names, err := List(baseDir, projectID)
	th.AssertError(t, err, "", "list error")
	th.AssertEqual(t, names, []string{"a.md", "journal.md", "notes.md", "settings.hidden.md"}, "sorted names")
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
