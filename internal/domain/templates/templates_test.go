package templates

import (
	"hermes-relay/internal/domain/entities/file"
	"strings"
	"testing"
)

func TestDefaultFiles(t *testing.T) {
	files := DefaultFiles()

	if len(files) == 0 {
		t.Error("DefaultFiles() should return at least one file")
	}

	codebookFound := false
	for _, f := range files {
		if f.Type == file.FileTypeCodebook {
			codebookFound = true
			if f.Name == "" {
				t.Error("Codebook file should have a name")
			}
			if !strings.Contains(f.Content, "Coding Preferences") {
				t.Error("Codebook content should contain 'Coding Preferences'")
			}
		}
	}

	if !codebookFound {
		t.Error("DefaultFiles() should include a codebook file")
	}
}
