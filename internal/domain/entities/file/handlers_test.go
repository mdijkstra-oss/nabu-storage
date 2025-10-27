package file

import (
	th "hermes-relay/internal/lib/test-helpers"
	"testing"
	"time"
)

func TestFileRouter(t *testing.T) {
	baseTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	// Test helper for creating valid CreateFileData
	validCreateFile := func() CreateFileData {
		return CreateFileData{
			BaseFile: BaseFile{
				Name: "test-file.txt",
				Attributes: Attributes{
					Title:   "Test File",
					Summary: "A test file",
					Time:    baseTime,
				},
			},
			Content: "Test content",
		}
	}

	// Test helper for creating valid CodeFileData
	validCodeFile := func() CodeFileData {
		return CodeFileData{
			Actions: []CodingAction{
				{
					CodeSlug: "topic:test",
					Action:   SetCoding,
					Sections: []CodedSectionAttributes{
						{
							Text:     "Some text to code",
							AIReason: "Test reason",
							Comment:  "Test comment",
						},
					},
					ChunkID: "chunk-1",
				},
			},
		}
	}

	// Specific test cases
	tests := []th.RouterTestCase{
		th.CommandToEventCase("CreateFile with valid payload", Create, CreatedFile, validCreateFile(), EntityName, ""),

		th.CommandToEventCase("CreateFile with minimal required fields", Create, CreatedFile, CreateFileData{
			BaseFile: BaseFile{
				Name: "minimal.txt",
				Attributes: Attributes{
					Title: "Minimal File",
				},
			},
			Content: "Content",
		}, EntityName, ""),

		th.CommandToEventCase("CodeFile with valid payload", CodeFile, CodedFile, validCodeFile(), EntityName, "file-123"),

		th.CommandToEventCase("CodeFile with multiple actions", CodeFile, CodedFile, CodeFileData{
			Actions: []CodingAction{
				{
					CodeSlug: "topic:first",
					Action:   SetCoding,
					Sections: []CodedSectionAttributes{
						{
							Text:     "First section",
							AIReason: "First reason",
						},
					},
					ChunkID: "chunk-1",
				},
				{
					CodeSlug: "topic:second",
					Action:   AppendCoding,
					Sections: []CodedSectionAttributes{
						{
							Text:    "Second section",
							Comment: "Second comment",
						},
					},
					ChunkID: "chunk-2",
				},
			},
		}, EntityName, "file-456"),

		th.CommandToEventCase("CodeFile with partial attributes (text only)", CodeFile, CodedFile, CodeFileData{
			Actions: []CodingAction{
				{
					CodeSlug: "topic:minimal",
					Action:   SetCoding,
					Sections: []CodedSectionAttributes{
						{
							Text: "Just text",
						},
					},
					ChunkID: "chunk-1",
				},
			},
		}, EntityName, "file-789"),

		th.CommandToEventCase("CodeFile with AppendCoding action", CodeFile, CodedFile, CodeFileData{
			Actions: []CodingAction{
				{
					CodeSlug: "topic:append",
					Action:   AppendCoding,
					Sections: []CodedSectionAttributes{
						{
							Text:     "Appending text",
							AIReason: "Append reason",
							Comment:  "Append comment",
						},
					},
					ChunkID: "chunk-1",
				},
			},
		}, EntityName, "file-append"),

		th.CommandToEventCase("CodeFile with RemoveCoding action", CodeFile, CodedFile, CodeFileData{
			Actions: []CodingAction{
				{
					CodeSlug: "topic:remove",
					Action:   RemoveCoding,
					Sections: []CodedSectionAttributes{
						{
							Text: "Dummy text",
						},
					},
					ChunkID: "chunk-1",
				},
			},
		}, EntityName, "file-remove"),

		th.CommandToEventCase[any]("ClearCoding", ClearCoding, ClearedCoding, nil, EntityName, "file-clear"),

		th.CommandToEventCase("MergeCodes with valid payload", MergeCodes, MergedCodes, MergeCodesData{
			Source: "topic:climate-old",
			Target: "topic:climate",
		}, EntityName, "file-merge"),

		th.ValidationErrorCase("CreateFile with missing Name", Create, CreateFilePayload{
			BaseFile: BaseFile{
				Attributes: Attributes{
					Title: "No Name File",
				},
			},
			Content: "Content",
		}, EntityName, ""),

		th.ValidationErrorCase("CreateFile with missing Title", Create, CreateFilePayload{
			BaseFile: BaseFile{
				Name: "file.txt",
			},
			Content: "Content",
		}, EntityName, ""),

		th.ValidationErrorCase("CodeFile with empty actions array", CodeFile, CodeFilePayload{
			Actions: []CodingAction{},
		}, EntityName, "file-123"),

		th.ValidationErrorCase("CodeFile with missing CodeSlug", CodeFile, CodeFilePayload{
			Actions: []CodingAction{
				{
					Action: SetCoding,
					Sections: []CodedSectionAttributes{
						{Text: "Some text"},
					},
					ChunkID: "chunk-1",
				},
			},
		}, EntityName, "file-123"),

		th.ValidationErrorCase("CodeFile with missing ChunkID", CodeFile, CodeFilePayload{
			Actions: []CodingAction{
				{
					CodeSlug: "topic:test",
					Action:   SetCoding,
					Sections: []CodedSectionAttributes{
						{Text: "Some text"},
					},
				},
			},
		}, EntityName, "file-123"),

		th.ValidationErrorCase("CodeFile with empty sections array", CodeFile, CodeFilePayload{
			Actions: []CodingAction{
				{
					CodeSlug: "topic:test",
					Action:   SetCoding,
					Sections: []CodedSectionAttributes{},
					ChunkID:  "chunk-1",
				},
			},
		}, EntityName, "file-123"),

		th.ValidationErrorCase("CodeFile with invalid slug (no colon)", CodeFile, CodeFilePayload{
			Actions: []CodingAction{
				{
					CodeSlug: "topicclimate",
					Action:   SetCoding,
					Sections: []CodedSectionAttributes{
						{Text: "Some text"},
					},
					ChunkID: "chunk-1",
				},
			},
		}, EntityName, "file-123"),

		th.ValidationErrorCase("CodeFile with invalid slug (uppercase)", CodeFile, CodeFilePayload{
			Actions: []CodingAction{
				{
					CodeSlug: "Topic:climate",
					Action:   SetCoding,
					Sections: []CodedSectionAttributes{
						{Text: "Some text"},
					},
					ChunkID: "chunk-1",
				},
			},
		}, EntityName, "file-123"),

		th.ValidationErrorCase("CodeFile with invalid slug (empty before colon)", CodeFile, CodeFilePayload{
			Actions: []CodingAction{
				{
					CodeSlug: ":climate",
					Action:   SetCoding,
					Sections: []CodedSectionAttributes{
						{Text: "Some text"},
					},
					ChunkID: "chunk-1",
				},
			},
		}, EntityName, "file-123"),

		th.ValidationErrorCase("MergeCodes with missing Source", MergeCodes, MergeCodesPayload{
			Target: "topic:climate",
		}, EntityName, "file-merge"),

		th.ValidationErrorCase("MergeCodes with missing Target", MergeCodes, MergeCodesPayload{
			Source: "topic:climate-old",
		}, EntityName, "file-merge"),

		th.ValidationErrorCase("MergeCodes with invalid Source slug", MergeCodes, MergeCodesPayload{
			Source: "invalid-slug",
			Target: "topic:climate",
		}, EntityName, "file-merge"),

		th.ValidationErrorCase("MergeCodes with invalid Target slug", MergeCodes, MergeCodesPayload{
			Source: "topic:climate-old",
			Target: "InvalidSlug",
		}, EntityName, "file-merge"),
	}

	// Add common router test cases (wrong entity, wrong message type, wrong action)
	tests = append(tests, th.CommonRouterTestCases(Create, validCreateFile(), EntityName)...)

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			th.TestRouter(t, Router, tt)
		})
	}
}
