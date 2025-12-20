package templates

import (
	"strings"
	"testing"
)

func TestDefaultDocuments(t *testing.T) {
	docs := DefaultDocuments()

	if len(docs) == 0 {
		t.Error("DefaultDocuments() should return at least one document")
	}

	codebookFound := false
	for _, d := range docs {
		if d.Name == "Codebook" {
			codebookFound = true
			if !strings.Contains(d.Content, "Coding Preferences") {
				t.Error("Codebook content should contain 'Coding Preferences'")
			}
		}
	}

	if !codebookFound {
		t.Error("DefaultDocuments() should include a codebook document")
	}
}
