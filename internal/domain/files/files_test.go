package files

import (
	"os"
	"path/filepath"
	"testing"

	th "hermes-relay/internal/lib/test-helpers"
)

func TestCreate(t *testing.T) {
	baseDir := t.TempDir()
	projectID := "test-project"

	tests := []struct {
		Name      string
		Path      string
		Diff      string
		Expected  string
		ExpectErr string
	}{
		{
			Name:     "creates file with content",
			Path:     "notes.md",
			Diff:     "@@\n+# Hello\n+World",
			Expected: "# Hello\nWorld",
		},
		{
			Name:     "creates empty file",
			Path:     "empty.md",
			Diff:     "",
			Expected: "",
		},
		{
			Name:     "creates file with single line",
			Path:     "single.md",
			Diff:     "@@\n+# Title",
			Expected: "# Title",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			err := Create(baseDir, projectID, tt.Path, tt.Diff)
			th.AssertError(t, err, tt.ExpectErr, "create error")

			if tt.ExpectErr == "" {
				content, err := Read(baseDir, projectID, tt.Path)
				th.AssertError(t, err, "", "read error")
				th.AssertEqual(t, content, tt.Expected, "content")
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	baseDir := t.TempDir()
	projectID := "test-project"

	setupFile := func(path, content string) {
		th.AssertError(t, Write(baseDir, projectID, path, content), "", "setup write")
	}

	tests := []struct {
		Name      string
		Setup     func()
		Path      string
		Diff      string
		Expected  string
		ExpectErr string
	}{
		{
			Name: "updates existing file",
			Setup: func() {
				setupFile("test.md", "hello\nworld")
			},
			Path:     "test.md",
			Diff:     "@@\n-hello\n+goodbye\nworld",
			Expected: "goodbye\nworld",
		},
		{
			Name: "appends to file",
			Setup: func() {
				setupFile("append.md", "# Title")
			},
			Path:     "append.md",
			Diff:     "@@\n+\n+New content",
			Expected: "# Title\n\nNew content",
		},
		{
			Name:      "fails on non-existent file",
			Setup:     func() {},
			Path:      "nonexistent.md",
			Diff:      "@@\n+content",
			ExpectErr: "not found",
		},
		{
			Name: "fails when context not found",
			Setup: func() {
				setupFile("context.md", "actual content")
			},
			Path:      "context.md",
			Diff:      "@@\n-wrong context\n+replacement",
			ExpectErr: "context not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			tt.Setup()
			err := Update(baseDir, projectID, tt.Path, tt.Diff)
			th.AssertError(t, err, tt.ExpectErr, "update error")

			if tt.ExpectErr == "" {
				content, err := Read(baseDir, projectID, tt.Path)
				th.AssertError(t, err, "", "read error")
				th.AssertEqual(t, content, tt.Expected, "content")
			}
		})
	}
}

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

func TestDelete(t *testing.T) {
	tests := []struct {
		Name       string
		SetupFile  bool
		ExpectErr  string
		FileGone   bool
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
