package fileview

import (
	"hermes-relay/internal/cqrs"
	"hermes-relay/internal/domain/entities/file"
	th "hermes-relay/internal/lib/test-helpers"
	"testing"
	"time"
)

const EntityName = "File"

// newFileDomainEvent creates a domain event for file entity testing
func newFileDomainEvent(aggregateID string, action cqrs.Action, payload any) *cqrs.AnyMessage {
	return cqrs.ToAny(cqrs.NewDomainEvent(action, payload, EntityName, aggregateID, (*cqrs.AnyMessage)(nil)))
}

// createTestFile creates a file with the given content for testing
func createTestFile(aggregateID, name, title, summary, content string, testTime time.Time) *File {
	return Reducer(nil, newFileDomainEvent(aggregateID, file.CreatedFile, &file.CreatedFilePayload{
		BaseFile: file.BaseFile{
			Name: name,
			Attributes: file.Attributes{
				Title:   title,
				Summary: summary,
				Time:    testTime,
			},
		},
		Content: content,
	}))
}

// assertChunks compares only the chunks of a file
func assertChunks(t *testing.T, got *File, wantChunks []file.Chunk, msg string) {
	t.Helper()
	th.AssertEqualIgnoreFields(t, got.Chunks, wantChunks, msg, file.Chunk{}, "ID")
}

func TestFileReducer_CreatedFile(t *testing.T) {
	aggregateID := "file-123"
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	content := "Short test content"

	state := createTestFile(aggregateID, "test.txt", "Test", "Summary", content, testTime)

	th.AssertEqualIgnoreFields(t, state, &File{
		BaseFile: file.BaseFile{
			ID:   aggregateID,
			Name: "test.txt",
			Attributes: file.Attributes{
				Title:   "Test",
				Summary: "Summary",
				Time:    testTime,
			},
		},
		Content: content,
		Chunks: []file.Chunk{
			{
				Content: content + "\n",
				Codes:   []file.CodedSection{},
			},
		},
	}, "After create", file.Chunk{}, "ID")
}

func TestFileReducer_Coding(t *testing.T) {
	aggregateID := "file-coding"
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	content := "Climate change impacts global warming. Rising temperatures affect ecosystems."

	state := createTestFile(aggregateID, "coding.txt", "Coding Test", "Testing all coding operations", content, testTime)

	if state == nil || len(state.Chunks) == 0 {
		t.Fatal("Failed to create initial state with chunks")
	}

	chunkID := state.Chunks[0].ID

	// Step 1: AppendCoding - add initial codes
	state = Reducer(state, newFileDomainEvent(aggregateID, file.CodedFile, &file.CodedFilePayload{
		Actions: []file.CodingAction{
			{
				CodeSlug: "topic:climate",
				Action:   file.AppendCoding,
				ChunkID:  chunkID,
				Sections: []file.CodedSectionAttributes{
					{Text: "Climate change", AIReason: "Climate reference 1"},
					{Text: "global warming", AIReason: "Climate reference 2"},
				},
			},
		},
	}))

	assertChunks(t, state, []file.Chunk{
		{
			Content: content + "\n",
			Codes: []file.CodedSection{
				{StartIndex: 0, EndIndex: 14, CodeSlug: "topic:climate", CodedSectionAttributes: file.CodedSectionAttributes{Text: "Climate change", AIReason: "Climate reference 1"}},
				{StartIndex: 23, EndIndex: 38, CodeSlug: "topic:climate", CodedSectionAttributes: file.CodedSectionAttributes{Text: "global warming", AIReason: "Climate reference 2"}},
			},
		},
	}, "After append two climate codes")

	// Step 2: AppendCoding - add different code slug
	state = Reducer(state, newFileDomainEvent(aggregateID, file.CodedFile, &file.CodedFilePayload{
		Actions: []file.CodingAction{
			{
				CodeSlug: "topic:temperature",
				Action:   file.AppendCoding,
				ChunkID:  chunkID,
				Sections: []file.CodedSectionAttributes{
					{Text: "Rising temperatures", AIReason: "Temperature topic"},
				},
			},
		},
	}))

	th.AssertEqual(t, len(state.Chunks[0].Codes), 3, "Should have 3 codes total")

	// Step 3: SetCoding - replace one slug's codes
	state = Reducer(state, newFileDomainEvent(aggregateID, file.CodedFile, &file.CodedFilePayload{
		Actions: []file.CodingAction{
			{
				CodeSlug: "topic:climate",
				Action:   file.SetCoding,
				ChunkID:  chunkID,
				Sections: []file.CodedSectionAttributes{
					{Text: "ecosystems", AIReason: "New climate reference"},
				},
			},
		},
	}))

	assertChunks(t, state, []file.Chunk{
		{
			Content: content + "\n",
			Codes: []file.CodedSection{
				{StartIndex: 39, EndIndex: 58, CodeSlug: "topic:temperature", CodedSectionAttributes: file.CodedSectionAttributes{Text: "Rising temperatures", AIReason: "Temperature topic"}},
				{StartIndex: 66, EndIndex: 77, CodeSlug: "topic:climate", CodedSectionAttributes: file.CodedSectionAttributes{Text: "ecosystems", AIReason: "New climate reference"}},
			},
		},
	}, "After SetCoding climate (should replace 2 with 1, keep temperature)")

	// Step 4: RemoveCoding - remove all codes with specific slug
	state = Reducer(state, newFileDomainEvent(aggregateID, file.CodedFile, &file.CodedFilePayload{
		Actions: []file.CodingAction{
			{
				CodeSlug: "topic:temperature",
				Action:   file.RemoveCoding,
				ChunkID:  chunkID,
				Sections: []file.CodedSectionAttributes{{Text: "dummy"}}, // Required by validation but not used
			},
		},
	}))

	assertChunks(t, state, []file.Chunk{
		{
			Content: content + "\n",
			Codes: []file.CodedSection{
				{StartIndex: 66, EndIndex: 77, CodeSlug: "topic:climate", CodedSectionAttributes: file.CodedSectionAttributes{Text: "ecosystems", AIReason: "New climate reference"}},
			},
		},
	}, "After RemoveCoding temperature (should remove temperature code, keep climate)")

	// Clear coding
	state = Reducer(state, newFileDomainEvent(aggregateID, file.ClearedCoding, nil))

	assertChunks(t, state, []file.Chunk{
		{
			Content: content + "\n",
			Codes:   []file.CodedSection{},
		},
	}, "After clearing all codes")
}

func TestFileReducer_MergeCodes(t *testing.T) {
	aggregateID := "file-merge"
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	content := "Climate change impacts global warming. Rising temperatures affect ecosystems. More climate data here."

	state := createTestFile(aggregateID, "merge.txt", "Merge Test", "Testing merge operation", content, testTime)

	if state == nil || len(state.Chunks) == 0 {
		t.Fatal("Failed to create initial state with chunks")
	}

	chunkID := state.Chunks[0].ID

	// Add codes with different slugs across the file
	state = Reducer(state, newFileDomainEvent(aggregateID, file.CodedFile, &file.CodedFilePayload{
		Actions: []file.CodingAction{
			{
				CodeSlug: "topic:climate-old",
				Action:   file.AppendCoding,
				ChunkID:  chunkID,
				Sections: []file.CodedSectionAttributes{
					{Text: "Climate change", AIReason: "Old climate slug 1"},
					{Text: "climate data", AIReason: "Old climate slug 2"},
				},
			},
			{
				CodeSlug: "topic:temperature",
				Action:   file.AppendCoding,
				ChunkID:  chunkID,
				Sections: []file.CodedSectionAttributes{
					{Text: "Rising temperatures", AIReason: "Temperature topic"},
				},
			},
		},
	}))

	th.AssertEqual(t, len(state.Chunks[0].Codes), 3, "Should have 3 codes before merge")

	// Merge topic:climate-old into topic:climate
	state = Reducer(state, newFileDomainEvent(aggregateID, file.MergedCodes, &file.MergedCodesPayload{
		Source: "topic:climate-old",
		Target: "topic:climate",
	}))

	assertChunks(t, state, []file.Chunk{
		{
			Content: content + "\n",
			Codes: []file.CodedSection{
				{StartIndex: 0, EndIndex: 14, CodeSlug: "topic:climate", CodedSectionAttributes: file.CodedSectionAttributes{Text: "Climate change", AIReason: "Old climate slug 1"}},
				{StartIndex: 83, EndIndex: 95, CodeSlug: "topic:climate", CodedSectionAttributes: file.CodedSectionAttributes{Text: "climate data", AIReason: "Old climate slug 2"}},
				{StartIndex: 39, EndIndex: 58, CodeSlug: "topic:temperature", CodedSectionAttributes: file.CodedSectionAttributes{Text: "Rising temperatures", AIReason: "Temperature topic"}},
			},
		},
	}, "After merging climate-old to climate (should change slug but keep other attributes)")
}
