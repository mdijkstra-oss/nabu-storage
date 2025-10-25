package textsearch

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestGenerateExpectedOutputs regenerates expected output files for chunker tests.
//
// SAFETY: Requires explicit GENERATE argument to prevent accidental regeneration.
//
// Usage:
//   # Regenerate a single file (recommended for iterating on one test case)
//   GENERATE=word-boundary.md go test -run TestGenerateExpectedOutputs
//
//   # Regenerate all files (⚠️  WARNING: You MUST manually verify each file after!)
//   GENERATE=all go test -run TestGenerateExpectedOutputs
//
// The test will skip if no GENERATE argument is provided.
func TestGenerateExpectedOutputs(t *testing.T) {
	generateArg := os.Getenv("GENERATE")
	if generateArg == "" {
		t.Skip("Skipped: GENERATE argument required. Use:\n  GENERATE=all go test -run TestGenerateExpectedOutputs\n  GENERATE=filename.md go test -run TestGenerateExpectedOutputs")
	}

	inputDir := filepath.Join("testdata", "input")
	var filesToGenerate []string

	if generateArg == "all" {
		t.Log("⚠️  WARNING: Regenerating ALL expected outputs. You MUST manually verify each file!")

		// Discover all test files
		entries, err := os.ReadDir(inputDir)
		if err != nil {
			t.Fatalf("Failed to read input directory: %v", err)
		}

		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
				filesToGenerate = append(filesToGenerate, entry.Name())
			}
		}
	} else {
		// Generate single file
		if !strings.HasSuffix(generateArg, ".md") {
			t.Fatalf("File must end with .md, got: %s", generateArg)
		}

		// Verify file exists
		inputPath := filepath.Join(inputDir, generateArg)
		if _, err := os.Stat(inputPath); os.IsNotExist(err) {
			t.Fatalf("File does not exist: %s", inputPath)
		}

		filesToGenerate = append(filesToGenerate, generateArg)
	}

	for _, inputFile := range filesToGenerate {
		expectedFile := strings.TrimSuffix(inputFile, ".md") + ".txt"

		// Read input
		inputPath := filepath.Join("testdata", "input", inputFile)
		input, err := os.ReadFile(inputPath)
		if err != nil {
			t.Fatalf("Failed to read input file: %v", err)
		}

		// Run chunker with test parameters
		chunks := ChunkBlocks(string(input), testMinSize, testMaxSize)

		// Format output with chunk delimiter on its own line
		// Using \n---CHUNK---\n preserves exact chunk content when split
		output := strings.Join(chunks, "\n---CHUNK---\n")

		// Write expected output
		expectedPath := filepath.Join("testdata", "expected", expectedFile)
		err = os.WriteFile(expectedPath, []byte(output), 0644)
		if err != nil {
			t.Fatalf("Failed to write expected file: %v", err)
		}

		t.Logf("Generated %s (%d chunks)", expectedFile, len(chunks))
	}
}
