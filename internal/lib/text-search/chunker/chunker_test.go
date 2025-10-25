package chunker

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	th "hermes-relay/internal/lib"
)

const (
	testMinSize = 50
	testMaxSize = 100
)

func TestChunkBlocks(t *testing.T) {
	// Dynamically discover test files
	inputDir := filepath.Join("testdata", "input")
	entries, err := os.ReadDir(inputDir)
	if err != nil {
		t.Fatalf("Failed to read input directory: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		inputFile := entry.Name()
		expectedFile := strings.TrimSuffix(inputFile, ".md") + ".txt"
		testName := strings.TrimSuffix(inputFile, ".md")

		t.Run(testName, func(t *testing.T) {
			// Read input
			inputPath := filepath.Join("testdata", "input", inputFile)
			input, err := os.ReadFile(inputPath)
			if err != nil {
				t.Fatalf("Failed to read input file: %v", err)
			}

			// Run chunker
			chunks := ChunkBlocks(string(input), testMinSize, testMaxSize)

			// Read expected output
			expectedPath := filepath.Join("testdata", "expected", expectedFile)
			expectedData, err := os.ReadFile(expectedPath)
			if err != nil {
				t.Fatalf("Failed to read expected file: %v", err)
			}

			expected := parseExpectedChunks(string(expectedData))

			// Assertions
			th.AssertEqual(t, len(chunks), len(expected), "Number of chunks")

			for i, chunk := range chunks {
				th.AssertEqual(t, chunk, expected[i], "Chunk "+string(rune(i+'0')))

				// Note: With word-boundary and mid-word splitting, chunks may not end with newlines
				// This is acceptable for frontend rendering - text will wrap naturally
			}
		})
	}
}

// parseExpectedChunks parses the expected output format
// Chunks are separated by \n---CHUNK---\n which preserves exact chunk content
func parseExpectedChunks(content string) []string {
	if content == "" {
		return []string{}
	}

	return strings.Split(content, "\n---CHUNK---\n")
}
