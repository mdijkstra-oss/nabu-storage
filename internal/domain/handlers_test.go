package domain

import (
	"os"
	"path/filepath"
	"testing"

	"hermes-relay/internal/domain/files"
	th "hermes-relay/internal/lib/test-helpers"
)

func TestExecute(t *testing.T) {
	baseDir := t.TempDir()
	projectID := "test-project"

	tests := []struct {
		Name      string
		Setup     func()
		Command   Command
		ExpectErr string
		Verify    func(t *testing.T)
	}{
		{
			Name:  "WriteFile succeeds",
			Setup: func() {},
			Command: Command{
				Action:  WriteFile,
				Path:    "new.md",
				Content: "# Hello",
			},
			Verify: func(t *testing.T) {
				content, _ := files.Read(baseDir, projectID, "new.md")
				th.AssertEqual(t, content, "# Hello", "content")
			},
		},
		{
			Name:  "WriteFile overwrites existing",
			Setup: func() { _ = files.Write(baseDir, projectID, "overwrite.md", "old") },
			Command: Command{
				Action:  WriteFile,
				Path:    "overwrite.md",
				Content: "new",
			},
			Verify: func(t *testing.T) {
				content, _ := files.Read(baseDir, projectID, "overwrite.md")
				th.AssertEqual(t, content, "new", "content")
			},
		},
		{
			Name:      "WriteFile fails without path",
			Setup:     func() {},
			Command:   Command{Action: WriteFile, Content: "x"},
			ExpectErr: "path",
		},
		{
			Name: "DeleteFile succeeds",
			Setup: func() {
				_ = files.Write(baseDir, projectID, "delete.md", "x")
			},
			Command: Command{
				Action: DeleteFile,
				Path:   "delete.md",
			},
			Verify: func(t *testing.T) {
				th.AssertEqual(t, files.Exists(baseDir, projectID, "delete.md"), false, "deleted")
			},
		},
		{
			Name:  "DeleteFile fails if not found",
			Setup: func() {},
			Command: Command{
				Action: DeleteFile,
				Path:   "nope.md",
			},
			ExpectErr: "not found",
		},
		{
			Name: "RenameFile succeeds",
			Setup: func() {
				_ = files.Write(baseDir, projectID, "oldname.md", "content")
			},
			Command: Command{
				Action:  RenameFile,
				Path:    "oldname.md",
				NewPath: "newname.md",
			},
			Verify: func(t *testing.T) {
				th.AssertEqual(t, files.Exists(baseDir, projectID, "oldname.md"), false, "old gone")
				th.AssertEqual(t, files.Exists(baseDir, projectID, "newname.md"), true, "new exists")
				content, _ := files.Read(baseDir, projectID, "newname.md")
				th.AssertEqual(t, content, "content", "content preserved")
			},
		},
		{
			Name:  "RenameFile fails if source not found",
			Setup: func() {},
			Command: Command{
				Action:  RenameFile,
				Path:    "notexist.md",
				NewPath: "new.md",
			},
			ExpectErr: "not found",
		},
		{
			Name: "RenameFile fails if dest exists",
			Setup: func() {
				_ = files.Write(baseDir, projectID, "source.md", "src")
				_ = files.Write(baseDir, projectID, "dest.md", "dst")
			},
			Command: Command{
				Action:  RenameFile,
				Path:    "source.md",
				NewPath: "dest.md",
			},
			ExpectErr: "already exists",
		},
		{
			Name: "RenameFile fails with invalid new path",
			Setup: func() {
				_ = files.Write(baseDir, projectID, "valid.md", "x")
			},
			Command: Command{
				Action:  RenameFile,
				Path:    "valid.md",
				NewPath: "../evil.md",
			},
			ExpectErr: "invalid new path",
		},
		{
			Name:    "Commit succeeds",
			Setup:   func() {},
			Command: Command{Action: Commit},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			tt.Setup()
			err := Execute(&tt.Command, projectID, baseDir)
			th.AssertError(t, err, tt.ExpectErr, "execute")
			if tt.Verify != nil && tt.ExpectErr == "" {
				tt.Verify(t)
			}
		})
	}
}

func TestExecutePanicsOnUnknownAction(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for unknown action")
		}
	}()

	_ = Execute(&Command{Action: "Unknown"}, "proj", "/tmp")
}

func TestPathTraversalPrevention(t *testing.T) {
	baseDir := t.TempDir()
	projectID := "test-project"
	outsideFile := filepath.Join(baseDir, "should-not-exist.txt")

	tests := []struct {
		Name string
		Path string
	}{
		{Name: "parent traversal", Path: "../should-not-exist.txt"},
		{Name: "nested traversal", Path: "foo/../../../should-not-exist.txt"},
		{Name: "absolute-like", Path: "/etc/passwd"},
		{Name: "nested directory", Path: "subdir/file.md"},
		{Name: "backslash nested", Path: "subdir\\file.md"},
		{Name: "hidden file", Path: ".secret"},
		{Name: "empty path", Path: ""},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			err := Execute(&Command{
				Action:  WriteFile,
				Path:    tt.Path,
				Content: "malicious content",
			}, projectID, baseDir)

			th.AssertError(t, err, "invalid path", "should reject")

			if _, statErr := os.Stat(outsideFile); statErr == nil {
				t.Fatal("file was written outside project directory")
			}

			entries, _ := os.ReadDir(baseDir)
			for _, e := range entries {
				if e.Name() != projectID && e.Name() != "should-not-exist.txt" {
					t.Fatalf("unexpected file created: %s", e.Name())
				}
			}
		})
	}
}

func TestPathTraversalPreventionUpdate(t *testing.T) {
	baseDir := t.TempDir()
	projectID := "test-project"

	targetFile := filepath.Join(baseDir, "target.txt")
	_ = os.WriteFile(targetFile, []byte("original"), 0644)

	err := Execute(&Command{
		Action:  WriteFile,
		Path:    "../target.txt",
		Content: "hacked",
	}, projectID, baseDir)

	th.AssertError(t, err, "invalid path", "should reject traversal")

	content, _ := os.ReadFile(targetFile)
	th.AssertEqual(t, string(content), "original", "file should be unchanged")
}

func TestPathTraversalPreventionDelete(t *testing.T) {
	baseDir := t.TempDir()
	projectID := "test-project"

	targetFile := filepath.Join(baseDir, "important.txt")
	_ = os.WriteFile(targetFile, []byte("keep me"), 0644)

	err := Execute(&Command{
		Action: DeleteFile,
		Path:   "../important.txt",
	}, projectID, baseDir)

	th.AssertError(t, err, "invalid path", "should reject traversal")

	if _, statErr := os.Stat(targetFile); os.IsNotExist(statErr) {
		t.Fatal("file was deleted outside project directory")
	}
}

func TestPathTraversalPreventionRename(t *testing.T) {
	baseDir := t.TempDir()
	projectID := "test-project"

	_ = files.Write(baseDir, projectID, "source.md", "content")
	targetFile := filepath.Join(baseDir, "escaped.txt")

	tests := []struct {
		Name    string
		Path    string
		NewPath string
		ErrMsg  string
	}{
		{Name: "traversal in source", Path: "../source.md", NewPath: "dest.md", ErrMsg: "invalid path"},
		{Name: "traversal in dest", Path: "source.md", NewPath: "../escaped.txt", ErrMsg: "invalid new path"},
		{Name: "unicode in dest", Path: "source.md", NewPath: "foo／bar.md", ErrMsg: "invalid new path"},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			err := Execute(&Command{
				Action:  RenameFile,
				Path:    tt.Path,
				NewPath: tt.NewPath,
			}, projectID, baseDir)

			th.AssertError(t, err, tt.ErrMsg, "should reject")

			if _, statErr := os.Stat(targetFile); statErr == nil {
				t.Fatal("file was created outside project directory")
			}
		})
	}
}
