package handlers

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/domain/entities/file"
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
			input: commands.ToAny(commands.NewCommand[file.CreateFilePayload, any](file.CreateFile, file.CreateFilePayload{
				BaseFile: file.BaseFile{
					ProjectID: "project-1",
					Name:      "test-file.txt",
					Attributes: file.Attributes{
						Title:   "Test File",
						Summary: "A test file",
						Time:    baseTime,
					},
				},
				Content: "Test content",
			}, file.EntityName, "", nil)),
			expectErr: "",
			expectEvent: commands.ToAny(commands.NewDomainEvent[file.CreatedFilePayload, any](file.CreatedFile, file.CreatedFilePayload{
				BaseFile: file.BaseFile{
					ProjectID: "project-1",
					Name:      "test-file.txt",
					Attributes: file.Attributes{
						Title:   "Test File",
						Summary: "A test file",
						Time:    baseTime,
					},
				},
				Content: "Test content",
			}, file.EntityName, "", nil)),
		},
		{
			name: "CreateFile with minimal required fields",
			input: commands.ToAny(commands.NewCommand[file.CreateFilePayload, any](file.CreateFile, file.CreateFilePayload{
				BaseFile: file.BaseFile{
					ProjectID: "project-1",
					Name:      "minimal.txt",
					Attributes: file.Attributes{
						Title: "Minimal File",
					},
				},
				Content: "Content",
			}, file.EntityName, "", nil)),
			expectErr: "",
			expectEvent: commands.ToAny(commands.NewDomainEvent[file.CreatedFilePayload, any](file.CreatedFile, file.CreatedFilePayload{
				BaseFile: file.BaseFile{
					ProjectID: "project-1",
					Name:      "minimal.txt",
					Attributes: file.Attributes{
						Title: "Minimal File",
					},
				},
				Content: "Content",
			}, file.EntityName, "", nil)),
		},
		{
			name: "CodeFile with SetCoding action",
			input: commands.ToAny(commands.NewCommand[file.CodeFilePayload, any](file.CodeFile, file.CodeFilePayload{
				Actions: []file.CodingAction{
					{
						CodeID:   "code-1",
						CodeSlug: "topic:test",
						Action:   file.SetCoding,
						Sections: []file.CodedSectionAttributes{
							{
								Text:     "Some text to code",
								AIReason: "Test reason",
								Comment:  "Test comment",
							},
						},
						ChunkID: "chunk-1",
					},
				},
			}, file.EntityName, "file-123", nil)),
			expectErr: "",
			expectEvent: commands.ToAny(commands.NewDomainEvent[file.CodedFilePayload, any](file.CodedFile, file.CodedFilePayload{
				Actions: []file.CodingAction{
					{
						CodeID:   "code-1",
						CodeSlug: "topic:test",
						Action:   file.SetCoding,
						Sections: []file.CodedSectionAttributes{
							{
								Text:     "Some text to code",
								AIReason: "Test reason",
								Comment:  "Test comment",
							},
						},
						ChunkID: "chunk-1",
					},
				},
			}, file.EntityName, "file-123", nil)),
		},
		{
			name: "CodeFile with multiple actions",
			input: commands.ToAny(commands.NewCommand[file.CodeFilePayload, any](file.CodeFile, file.CodeFilePayload{
				Actions: []file.CodingAction{
					{
						CodeID:   "code-1",
						CodeSlug: "topic:first",
						Action:   file.SetCoding,
						Sections: []file.CodedSectionAttributes{
							{Text: "First section", AIReason: "First reason"},
						},
						ChunkID: "chunk-1",
					},
					{
						CodeID:   "code-2",
						CodeSlug: "topic:second",
						Action:   file.AppendCoding,
						Sections: []file.CodedSectionAttributes{
							{Text: "Second section", Comment: "Second comment"},
						},
						ChunkID: "chunk-2",
					},
				},
			}, file.EntityName, "file-456", nil)),
			expectErr: "",
			expectEvent: commands.ToAny(commands.NewDomainEvent[file.CodedFilePayload, any](file.CodedFile, file.CodedFilePayload{
				Actions: []file.CodingAction{
					{
						CodeID:   "code-1",
						CodeSlug: "topic:first",
						Action:   file.SetCoding,
						Sections: []file.CodedSectionAttributes{
							{Text: "First section", AIReason: "First reason"},
						},
						ChunkID: "chunk-1",
					},
					{
						CodeID:   "code-2",
						CodeSlug: "topic:second",
						Action:   file.AppendCoding,
						Sections: []file.CodedSectionAttributes{
							{Text: "Second section", Comment: "Second comment"},
						},
						ChunkID: "chunk-2",
					},
				},
			}, file.EntityName, "file-456", nil)),
		},
		{
			name: "CodeFile with AppendCoding action",
			input: commands.ToAny(commands.NewCommand[file.CodeFilePayload, any](file.CodeFile, file.CodeFilePayload{
				Actions: []file.CodingAction{
					{
						CodeID:   "code-1",
						CodeSlug: "topic:append",
						Action:   file.AppendCoding,
						Sections: []file.CodedSectionAttributes{
							{Text: "Appending text", AIReason: "Append reason", Comment: "Append comment"},
						},
						ChunkID: "chunk-1",
					},
				},
			}, file.EntityName, "file-append", nil)),
			expectErr: "",
			expectEvent: commands.ToAny(commands.NewDomainEvent[file.CodedFilePayload, any](file.CodedFile, file.CodedFilePayload{
				Actions: []file.CodingAction{
					{
						CodeID:   "code-1",
						CodeSlug: "topic:append",
						Action:   file.AppendCoding,
						Sections: []file.CodedSectionAttributes{
							{Text: "Appending text", AIReason: "Append reason", Comment: "Append comment"},
						},
						ChunkID: "chunk-1",
					},
				},
			}, file.EntityName, "file-append", nil)),
		},
		{
			name: "CodeFile with RemoveCoding action",
			input: commands.ToAny(commands.NewCommand[file.CodeFilePayload, any](file.CodeFile, file.CodeFilePayload{
				Actions: []file.CodingAction{
					{
						CodeID:   "code-1",
						CodeSlug: "topic:remove",
						Action:   file.RemoveCoding,
						Sections: []file.CodedSectionAttributes{
							{Text: "Dummy text"},
						},
						ChunkID: "chunk-1",
					},
				},
			}, file.EntityName, "file-remove", nil)),
			expectErr: "",
			expectEvent: commands.ToAny(commands.NewDomainEvent[file.CodedFilePayload, any](file.CodedFile, file.CodedFilePayload{
				Actions: []file.CodingAction{
					{
						CodeID:   "code-1",
						CodeSlug: "topic:remove",
						Action:   file.RemoveCoding,
						Sections: []file.CodedSectionAttributes{
							{Text: "Dummy text"},
						},
						ChunkID: "chunk-1",
					},
				},
			}, file.EntityName, "file-remove", nil)),
		},
		{
			name:        "ClearCoding",
			input:       commands.ToAny(commands.NewCommand[any, any](file.ClearCoding, nil, file.EntityName, "file-clear", nil)),
			expectErr:   "",
			expectEvent: commands.ToAny(commands.NewDomainEvent[any, any](file.ClearedCoding, nil, file.EntityName, "file-clear", nil)),
		},
		{
			name: "CreateFile with missing Name",
			input: commands.ToAny(commands.NewCommand[file.CreateFilePayload, any](file.CreateFile, file.CreateFilePayload{
				BaseFile: file.BaseFile{
					ProjectID: "project-1",
					Attributes: file.Attributes{
						Title: "No Name File",
					},
				},
				Content: "Content",
			}, file.EntityName, "", nil)),
			expectErr: "validation failed: Name is required",
		},
		{
			name: "CreateFile with missing Title",
			input: commands.ToAny(commands.NewCommand[file.CreateFilePayload, any](file.CreateFile, file.CreateFilePayload{
				BaseFile: file.BaseFile{
					ProjectID: "project-1",
					Name:      "file.txt",
				},
				Content: "Content",
			}, file.EntityName, "", nil)),
			expectErr: "validation failed: Title is required",
		},
		{
			name: "CodeFile with empty actions array",
			input: commands.ToAny(commands.NewCommand[file.CodeFilePayload, any](file.CodeFile, file.CodeFilePayload{
				Actions: []file.CodingAction{},
			}, file.EntityName, "file-123", nil)),
			expectErr: "validation failed: Actions must be at least 1 characters",
		},
		{
			name: "CodeFile with missing CodeSlug",
			input: commands.ToAny(commands.NewCommand[file.CodeFilePayload, any](file.CodeFile, file.CodeFilePayload{
				Actions: []file.CodingAction{
					{
						Action:   file.SetCoding,
						Sections: []file.CodedSectionAttributes{{Text: "Some text"}},
						ChunkID:  "chunk-1",
					},
				},
			}, file.EntityName, "file-123", nil)),
			expectErr: "validation failed: CodeSlug is required, CodeID is required",
		},
		{
			name: "CodeFile with missing ChunkID",
			input: commands.ToAny(commands.NewCommand[file.CodeFilePayload, any](file.CodeFile, file.CodeFilePayload{
				Actions: []file.CodingAction{
					{
						CodeSlug: "topic:test",
						Action:   file.SetCoding,
						Sections: []file.CodedSectionAttributes{{Text: "Some text"}},
					},
				},
			}, file.EntityName, "file-123", nil)),
			expectErr: "validation failed: CodeID is required, ChunkID is required",
		},
		{
			name: "CodeFile with empty sections array",
			input: commands.ToAny(commands.NewCommand[file.CodeFilePayload, any](file.CodeFile, file.CodeFilePayload{
				Actions: []file.CodingAction{
					{
						CodeSlug: "topic:test",
						Action:   file.SetCoding,
						Sections: []file.CodedSectionAttributes{},
						ChunkID:  "chunk-1",
					},
				},
			}, file.EntityName, "file-123", nil)),
			expectErr: "validation failed: CodeID is required, Sections must be at least 1 characters",
		},
		{
			name: "CodeFile with invalid slug (no colon)",
			input: commands.ToAny(commands.NewCommand[file.CodeFilePayload, any](file.CodeFile, file.CodeFilePayload{
				Actions: []file.CodingAction{
					{
						CodeSlug: "topicclimate",
						Action:   file.SetCoding,
						Sections: []file.CodedSectionAttributes{{Text: "Some text"}},
						ChunkID:  "chunk-1",
					},
				},
			}, file.EntityName, "file-123", nil)),
			expectErr: "validation failed: CodeSlug must match code slug format (lowercase with colon and optional dashes), CodeID is required",
		},
		{
			name: "CodeFile with invalid slug (uppercase)",
			input: commands.ToAny(commands.NewCommand[file.CodeFilePayload, any](file.CodeFile, file.CodeFilePayload{
				Actions: []file.CodingAction{
					{
						CodeSlug: "Topic:climate",
						Action:   file.SetCoding,
						Sections: []file.CodedSectionAttributes{{Text: "Some text"}},
						ChunkID:  "chunk-1",
					},
				},
			}, file.EntityName, "file-123", nil)),
			expectErr: "validation failed: CodeSlug must match code slug format (lowercase with colon and optional dashes), CodeID is required",
		},
		{
			name: "Wrong entity type returns nil",
			input: commands.ToAny(commands.NewCommand[file.CreateFilePayload, any](file.CreateFile, file.CreateFilePayload{
				BaseFile: file.BaseFile{
					ProjectID: "project-1",
					Name:      "test.txt",
					Attributes: file.Attributes{
						Title: "Test",
					},
				},
				Content: "Test",
			}, "DifferentEntity", "", nil)),
			expectErr:   "",
			expectEvent: nil,
		},
		{
			name: "Wrong action returns nil",
			input: commands.ToAny(commands.NewCommand[file.CreateFilePayload, any]("DifferentAction", file.CreateFilePayload{
				BaseFile: file.BaseFile{
					ProjectID: "project-1",
					Name:      "test.txt",
					Attributes: file.Attributes{
						Title: "Test",
					},
				},
				Content: "Test",
			}, file.EntityName, "test-aggregate-id", nil)),
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
