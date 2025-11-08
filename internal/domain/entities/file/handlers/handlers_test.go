package handlers

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/domain/entities/file"
	rh "hermes-relay/internal/lib/test-helpers/registry-helpers"
	"testing"
)

var cmds = []*commands.AnyMessage{}

func TestFileRouter(t *testing.T) {
	tests := []rh.RouterTestCase{
		{
			Name: "CreateFile with valid payload",
			Input: commands.ToAny(commands.NewCommand[file.CreateFilePayload, any](file.CreateFile, file.CreateFilePayload{
				ProjectID: "project-1",
				Name:      "test-file.txt",
				Content:   "Test content",
			}, file.EntityName, "", nil)),
			ExpectErr: "",
			ExpectEvent: commands.ToAny(commands.NewDomainEvent[file.CreatedFilePayload, any](file.CreatedFile, file.CreatedFilePayload{
				CreateFilePayload: file.CreateFilePayload{
					ProjectID: "project-1",
					Name:      "test-file.txt",
					Content:   "Test content",
				},
				Type:   file.FileTypeSource,
				Locked: true,
			}, file.EntityName, "", nil)),
		},
		{
			Name: "CreateFile with minimal required fields",
			Input: commands.ToAny(commands.NewCommand[file.CreateFilePayload, any](file.CreateFile, file.CreateFilePayload{
				ProjectID: "project-1",
				Name:      "minimal.txt",
				Content:   "Content",
			}, file.EntityName, "", nil)),
			ExpectErr: "",
			ExpectEvent: commands.ToAny(commands.NewDomainEvent[file.CreatedFilePayload, any](file.CreatedFile, file.CreatedFilePayload{
				CreateFilePayload: file.CreateFilePayload{
					ProjectID: "project-1",
					Name:      "minimal.txt",
					Content:   "Content",
				},
				Type:   file.FileTypeSource,
				Locked: true,
			}, file.EntityName, "", nil)),
		},
		{
			Name: "CodeFile with SetCoding action",
			Input: commands.ToAny(commands.NewCommand[file.CodeFilePayload, any](file.CodeFile, file.CodeFilePayload{
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
			ExpectErr: "",
			ExpectEvent: commands.ToAny(commands.NewDomainEvent[file.CodedFilePayload, any](file.CodedFile, file.CodedFilePayload{
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
			Name: "CodeFile with multiple actions",
			Input: commands.ToAny(commands.NewCommand[file.CodeFilePayload, any](file.CodeFile, file.CodeFilePayload{
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
			ExpectErr: "",
			ExpectEvent: commands.ToAny(commands.NewDomainEvent[file.CodedFilePayload, any](file.CodedFile, file.CodedFilePayload{
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
			Name: "CodeFile with AppendCoding action",
			Input: commands.ToAny(commands.NewCommand[file.CodeFilePayload, any](file.CodeFile, file.CodeFilePayload{
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
			ExpectErr: "",
			ExpectEvent: commands.ToAny(commands.NewDomainEvent[file.CodedFilePayload, any](file.CodedFile, file.CodedFilePayload{
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
			Name: "CodeFile with RemoveCoding action",
			Input: commands.ToAny(commands.NewCommand[file.CodeFilePayload, any](file.CodeFile, file.CodeFilePayload{
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
			ExpectErr: "",
			ExpectEvent: commands.ToAny(commands.NewDomainEvent[file.CodedFilePayload, any](file.CodedFile, file.CodedFilePayload{
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
			Name:        "ClearCoding",
			Input:       commands.ToAny(commands.NewCommand[any, any](file.ClearCoding, nil, file.EntityName, "file-clear", nil)),
			ExpectErr:   "",
			ExpectEvent: commands.ToAny(commands.NewDomainEvent[any, any](file.ClearedCoding, nil, file.EntityName, "file-clear", nil)),
		},
		{
			Name: "CreateFile with missing Name",
			Input: commands.ToAny(commands.NewCommand[file.CreateFilePayload, any](file.CreateFile, file.CreateFilePayload{
				ProjectID: "project-1",
				Content:   "Content",
			}, file.EntityName, "", nil)),
			ExpectErr: "validation failed: Name is required",
		},
		{
			Name: "CreateFile with missing ProjectID",
			Input: commands.ToAny(commands.NewCommand[file.CreateFilePayload, any](file.CreateFile, file.CreateFilePayload{
				Name:    "file.txt",
				Content: "Content",
			}, file.EntityName, "", nil)),
			ExpectErr: "validation failed: ProjectID is required",
		},
		{
			Name: "CodeFile with empty actions array",
			Input: commands.ToAny(commands.NewCommand[file.CodeFilePayload, any](file.CodeFile, file.CodeFilePayload{
				Actions: []file.CodingAction{},
			}, file.EntityName, "file-123", nil)),
			ExpectErr: "validation failed: Actions must be at least 1 characters",
		},
		{
			Name: "CodeFile with missing CodeSlug",
			Input: commands.ToAny(commands.NewCommand[file.CodeFilePayload, any](file.CodeFile, file.CodeFilePayload{
				Actions: []file.CodingAction{
					{
						Action:   file.SetCoding,
						Sections: []file.CodedSectionAttributes{{Text: "Some text"}},
						ChunkID:  "chunk-1",
					},
				},
			}, file.EntityName, "file-123", nil)),
			ExpectErr: "validation failed: CodeSlug is required, CodeID is required",
		},
		{
			Name: "CodeFile with missing ChunkID",
			Input: commands.ToAny(commands.NewCommand[file.CodeFilePayload, any](file.CodeFile, file.CodeFilePayload{
				Actions: []file.CodingAction{
					{
						CodeSlug: "topic:test",
						Action:   file.SetCoding,
						Sections: []file.CodedSectionAttributes{{Text: "Some text"}},
					},
				},
			}, file.EntityName, "file-123", nil)),
			ExpectErr: "validation failed: CodeID is required, ChunkID is required",
		},
		{
			Name: "CodeFile with empty sections array",
			Input: commands.ToAny(commands.NewCommand[file.CodeFilePayload, any](file.CodeFile, file.CodeFilePayload{
				Actions: []file.CodingAction{
					{
						CodeSlug: "topic:test",
						Action:   file.SetCoding,
						Sections: []file.CodedSectionAttributes{},
						ChunkID:  "chunk-1",
					},
				},
			}, file.EntityName, "file-123", nil)),
			ExpectErr: "validation failed: CodeID is required, Sections must be at least 1 characters",
		},
		{
			Name: "CodeFile with invalid slug (no colon)",
			Input: commands.ToAny(commands.NewCommand[file.CodeFilePayload, any](file.CodeFile, file.CodeFilePayload{
				Actions: []file.CodingAction{
					{
						CodeSlug: "topicclimate",
						Action:   file.SetCoding,
						Sections: []file.CodedSectionAttributes{{Text: "Some text"}},
						ChunkID:  "chunk-1",
					},
				},
			}, file.EntityName, "file-123", nil)),
			ExpectErr: "validation failed: CodeSlug must match code slug format (lowercase with colon and optional dashes), CodeID is required",
		},
		{
			Name: "CodeFile with invalid slug (uppercase)",
			Input: commands.ToAny(commands.NewCommand[file.CodeFilePayload, any](file.CodeFile, file.CodeFilePayload{
				Actions: []file.CodingAction{
					{
						CodeSlug: "Topic:climate",
						Action:   file.SetCoding,
						Sections: []file.CodedSectionAttributes{{Text: "Some text"}},
						ChunkID:  "chunk-1",
					},
				},
			}, file.EntityName, "file-123", nil)),
			ExpectErr: "validation failed: CodeSlug must match code slug format (lowercase with colon and optional dashes), CodeID is required",
		},
		{
			Name: "Wrong entity type returns nil",
			Input: commands.ToAny(commands.NewCommand[file.CreateFilePayload, any](file.CreateFile, file.CreateFilePayload{
				ProjectID: "project-1",
				Name:      "test.txt",
				Content:   "Test",
			}, "DifferentEntity", "", nil)),
			ExpectErr:   "",
			ExpectEvent: nil,
		},
		{
			Name: "Wrong action returns nil",
			Input: commands.ToAny(commands.NewCommand[file.CreateFilePayload, any]("DifferentAction", file.CreateFilePayload{
				ProjectID: "project-1",
				Name:      "test.txt",
				Content:   "Test",
			}, file.EntityName, "test-aggregate-id", nil)),
			ExpectErr:   "",
			ExpectEvent: nil,
		},
	}

	rh.RunRouterTests(t, cmds, tests, NewRouter)
}
