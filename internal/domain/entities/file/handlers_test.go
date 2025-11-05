package file

import (
	"hermes-relay/internal/cqrs/commands"
	th "hermes-relay/internal/lib/test-helpers"
	"testing"
	"time"
)

func TestFileRouter(t *testing.T) {
	baseTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		input       *commands.AnyMessage
		expectErr   string
		expectEvent *commands.AnyMessage
	}{
		{
			name: "CreateFile with valid payload",
			input: commands.ToAny(commands.NewCommand[CreateFilePayload, any](CreateFile, CreateFilePayload{
				BaseFile: BaseFile{
					ProjectID: "project-1",
					Name:      "test-file.txt",
					Attributes: Attributes{
						Title:   "Test File",
						Summary: "A test file",
						Time:    baseTime,
					},
				},
				Content: "Test content",
			}, EntityName, "", nil)),
			expectErr: "",
			expectEvent: commands.ToAny(commands.NewDomainEvent[CreatedFilePayload, any](CreatedFile, CreatedFilePayload{
				BaseFile: BaseFile{
					ProjectID: "project-1",
					Name:      "test-file.txt",
					Attributes: Attributes{
						Title:   "Test File",
						Summary: "A test file",
						Time:    baseTime,
					},
				},
				Content: "Test content",
			}, EntityName, "", nil)),
		},
		{
			name: "CreateFile with minimal required fields",
			input: commands.ToAny(commands.NewCommand[CreateFilePayload, any](CreateFile, CreateFilePayload{
				BaseFile: BaseFile{
					ProjectID: "project-1",
					Name:      "minimal.txt",
					Attributes: Attributes{
						Title: "Minimal File",
					},
				},
				Content: "Content",
			}, EntityName, "", nil)),
			expectErr: "",
			expectEvent: commands.ToAny(commands.NewDomainEvent[CreatedFilePayload, any](CreatedFile, CreatedFilePayload{
				BaseFile: BaseFile{
					ProjectID: "project-1",
					Name:      "minimal.txt",
					Attributes: Attributes{
						Title: "Minimal File",
					},
				},
				Content: "Content",
			}, EntityName, "", nil)),
		},
		{
			name: "CodeFile with SetCoding action",
			input: commands.ToAny(commands.NewCommand[CodeFilePayload, any](CodeFile, CodeFilePayload{
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
			}, EntityName, "file-123", nil)),
			expectErr: "",
			expectEvent: commands.ToAny(commands.NewDomainEvent[CodedFilePayload, any](CodedFile, CodedFilePayload{
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
			}, EntityName, "file-123", nil)),
		},
		{
			name: "CodeFile with multiple actions",
			input: commands.ToAny(commands.NewCommand[CodeFilePayload, any](CodeFile, CodeFilePayload{
				Actions: []CodingAction{
					{
						CodeSlug: "topic:first",
						Action:   SetCoding,
						Sections: []CodedSectionAttributes{
							{Text: "First section", AIReason: "First reason"},
						},
						ChunkID: "chunk-1",
					},
					{
						CodeSlug: "topic:second",
						Action:   AppendCoding,
						Sections: []CodedSectionAttributes{
							{Text: "Second section", Comment: "Second comment"},
						},
						ChunkID: "chunk-2",
					},
				},
			}, EntityName, "file-456", nil)),
			expectErr: "",
			expectEvent: commands.ToAny(commands.NewDomainEvent[CodedFilePayload, any](CodedFile, CodedFilePayload{
				Actions: []CodingAction{
					{
						CodeSlug: "topic:first",
						Action:   SetCoding,
						Sections: []CodedSectionAttributes{
							{Text: "First section", AIReason: "First reason"},
						},
						ChunkID: "chunk-1",
					},
					{
						CodeSlug: "topic:second",
						Action:   AppendCoding,
						Sections: []CodedSectionAttributes{
							{Text: "Second section", Comment: "Second comment"},
						},
						ChunkID: "chunk-2",
					},
				},
			}, EntityName, "file-456", nil)),
		},
		{
			name: "CodeFile with AppendCoding action",
			input: commands.ToAny(commands.NewCommand[CodeFilePayload, any](CodeFile, CodeFilePayload{
				Actions: []CodingAction{
					{
						CodeSlug: "topic:append",
						Action:   AppendCoding,
						Sections: []CodedSectionAttributes{
							{Text: "Appending text", AIReason: "Append reason", Comment: "Append comment"},
						},
						ChunkID: "chunk-1",
					},
				},
			}, EntityName, "file-append", nil)),
			expectErr: "",
			expectEvent: commands.ToAny(commands.NewDomainEvent[CodedFilePayload, any](CodedFile, CodedFilePayload{
				Actions: []CodingAction{
					{
						CodeSlug: "topic:append",
						Action:   AppendCoding,
						Sections: []CodedSectionAttributes{
							{Text: "Appending text", AIReason: "Append reason", Comment: "Append comment"},
						},
						ChunkID: "chunk-1",
					},
				},
			}, EntityName, "file-append", nil)),
		},
		{
			name: "CodeFile with RemoveCoding action",
			input: commands.ToAny(commands.NewCommand[CodeFilePayload, any](CodeFile, CodeFilePayload{
				Actions: []CodingAction{
					{
						CodeSlug: "topic:remove",
						Action:   RemoveCoding,
						Sections: []CodedSectionAttributes{
							{Text: "Dummy text"},
						},
						ChunkID: "chunk-1",
					},
				},
			}, EntityName, "file-remove", nil)),
			expectErr: "",
			expectEvent: commands.ToAny(commands.NewDomainEvent[CodedFilePayload, any](CodedFile, CodedFilePayload{
				Actions: []CodingAction{
					{
						CodeSlug: "topic:remove",
						Action:   RemoveCoding,
						Sections: []CodedSectionAttributes{
							{Text: "Dummy text"},
						},
						ChunkID: "chunk-1",
					},
				},
			}, EntityName, "file-remove", nil)),
		},
		{
			name:        "ClearCoding",
			input:       commands.ToAny(commands.NewCommand[any, any](ClearCoding, nil, EntityName, "file-clear", nil)),
			expectErr:   "",
			expectEvent: commands.ToAny(commands.NewDomainEvent[any, any](ClearedCoding, nil, EntityName, "file-clear", nil)),
		},
		{
			name: "MergeCodes with valid payload",
			input: commands.ToAny(commands.NewCommand[MergeCodesPayload, any](MergeCodes, MergeCodesPayload{
				Source: "topic:climate-old",
				Target: "topic:climate",
			}, EntityName, "file-merge", nil)),
			expectErr: "",
			expectEvent: commands.ToAny(commands.NewDomainEvent[MergedCodesPayload, any](MergedCodes, MergedCodesPayload{
				Source: "topic:climate-old",
				Target: "topic:climate",
			}, EntityName, "file-merge", nil)),
		},
		{
			name: "CreateFile with missing Name",
			input: commands.ToAny(commands.NewCommand[CreateFilePayload, any](CreateFile, CreateFilePayload{
				BaseFile: BaseFile{
					ProjectID: "project-1",
					Attributes: Attributes{
						Title: "No Name File",
					},
				},
				Content: "Content",
			}, EntityName, "", nil)),
			expectErr: "validation failed: Name is required",
		},
		{
			name: "CreateFile with missing Title",
			input: commands.ToAny(commands.NewCommand[CreateFilePayload, any](CreateFile, CreateFilePayload{
				BaseFile: BaseFile{
					ProjectID: "project-1",
					Name:      "file.txt",
				},
				Content: "Content",
			}, EntityName, "", nil)),
			expectErr: "validation failed: Title is required",
		},
		{
			name: "CodeFile with empty actions array",
			input: commands.ToAny(commands.NewCommand[CodeFilePayload, any](CodeFile, CodeFilePayload{
				Actions: []CodingAction{},
			}, EntityName, "file-123", nil)),
			expectErr: "validation failed: Actions must be at least 1 characters",
		},
		{
			name: "CodeFile with missing CodeSlug",
			input: commands.ToAny(commands.NewCommand[CodeFilePayload, any](CodeFile, CodeFilePayload{
				Actions: []CodingAction{
					{
						Action:   SetCoding,
						Sections: []CodedSectionAttributes{{Text: "Some text"}},
						ChunkID:  "chunk-1",
					},
				},
			}, EntityName, "file-123", nil)),
			expectErr: "validation failed: CodeSlug is required",
		},
		{
			name: "CodeFile with missing ChunkID",
			input: commands.ToAny(commands.NewCommand[CodeFilePayload, any](CodeFile, CodeFilePayload{
				Actions: []CodingAction{
					{
						CodeSlug: "topic:test",
						Action:   SetCoding,
						Sections: []CodedSectionAttributes{{Text: "Some text"}},
					},
				},
			}, EntityName, "file-123", nil)),
			expectErr: "validation failed: ChunkID is required",
		},
		{
			name: "CodeFile with empty sections array",
			input: commands.ToAny(commands.NewCommand[CodeFilePayload, any](CodeFile, CodeFilePayload{
				Actions: []CodingAction{
					{
						CodeSlug: "topic:test",
						Action:   SetCoding,
						Sections: []CodedSectionAttributes{},
						ChunkID:  "chunk-1",
					},
				},
			}, EntityName, "file-123", nil)),
			expectErr: "validation failed: Sections must be at least 1 characters",
		},
		{
			name: "CodeFile with invalid slug (no colon)",
			input: commands.ToAny(commands.NewCommand[CodeFilePayload, any](CodeFile, CodeFilePayload{
				Actions: []CodingAction{
					{
						CodeSlug: "topicclimate",
						Action:   SetCoding,
						Sections: []CodedSectionAttributes{{Text: "Some text"}},
						ChunkID:  "chunk-1",
					},
				},
			}, EntityName, "file-123", nil)),
			expectErr: "validation failed: CodeSlug must match code slug format (lowercase with colon and optional dashes)",
		},
		{
			name: "CodeFile with invalid slug (uppercase)",
			input: commands.ToAny(commands.NewCommand[CodeFilePayload, any](CodeFile, CodeFilePayload{
				Actions: []CodingAction{
					{
						CodeSlug: "Topic:climate",
						Action:   SetCoding,
						Sections: []CodedSectionAttributes{{Text: "Some text"}},
						ChunkID:  "chunk-1",
					},
				},
			}, EntityName, "file-123", nil)),
			expectErr: "validation failed: CodeSlug must match code slug format (lowercase with colon and optional dashes)",
		},
		{
			name: "MergeCodes with missing Source",
			input: commands.ToAny(commands.NewCommand[MergeCodesPayload, any](MergeCodes, MergeCodesPayload{
				Target: "topic:climate",
			}, EntityName, "file-merge", nil)),
			expectErr: "validation failed: Source is required",
		},
		{
			name: "MergeCodes with missing Target",
			input: commands.ToAny(commands.NewCommand[MergeCodesPayload, any](MergeCodes, MergeCodesPayload{
				Source: "topic:climate-old",
			}, EntityName, "file-merge", nil)),
			expectErr: "validation failed: Target is required",
		},
		{
			name: "MergeCodes with invalid Source slug",
			input: commands.ToAny(commands.NewCommand[MergeCodesPayload, any](MergeCodes, MergeCodesPayload{
				Source: "invalid-slug",
				Target: "topic:climate",
			}, EntityName, "file-merge", nil)),
			expectErr: "validation failed: Source must match code slug format (lowercase with colon and optional dashes)",
		},
		{
			name: "MergeCodes with invalid Target slug",
			input: commands.ToAny(commands.NewCommand[MergeCodesPayload, any](MergeCodes, MergeCodesPayload{
				Source: "topic:climate-old",
				Target: "InvalidSlug",
			}, EntityName, "file-merge", nil)),
			expectErr: "validation failed: Target must match code slug format (lowercase with colon and optional dashes)",
		},
		{
			name: "Wrong entity type returns nil",
			input: commands.ToAny(commands.NewCommand[CreateFilePayload, any](CreateFile, CreateFilePayload{
				BaseFile: BaseFile{
					ProjectID: "project-1",
					Name:      "test.txt",
					Attributes: Attributes{
						Title: "Test",
					},
				},
				Content: "Test",
			}, "DifferentEntity", "", nil)),
			expectErr:   "",
			expectEvent: nil,
		},
		{
			name: "Wrong message type returns nil",
			input: commands.ToAny(commands.NewDomainEvent[CreateFilePayload, any](CreateFile, CreateFilePayload{
				BaseFile: BaseFile{
					ProjectID: "project-1",
					Name:      "test.txt",
					Attributes: Attributes{
						Title: "Test",
					},
				},
				Content: "Test",
			}, EntityName, "", nil)),
			expectErr:   "",
			expectEvent: nil,
		},
		{
			name: "Wrong action returns nil",
			input: commands.ToAny(commands.NewCommand[CreateFilePayload, any]("DifferentAction", CreateFilePayload{
				BaseFile: BaseFile{
					ProjectID: "project-1",
					Name:      "test.txt",
					Attributes: Attributes{
						Title: "Test",
					},
				},
				Content: "Test",
			}, EntityName, "test-aggregate-id", nil)),
			expectErr:   "",
			expectEvent: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Router(tt.input, nil)

			th.AssertError(t, err, tt.expectErr, "error")
			if tt.expectErr == "" {
				th.AssertMessage(t, result, tt.expectEvent, "event")
			}
		})
	}
}
