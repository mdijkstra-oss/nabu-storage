package fileview

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/domain/entities/code"
	"hermes-relay/internal/domain/entities/file"
	th "hermes-relay/internal/lib/test-helpers"
	"testing"
	"time"
)

func TestFileReducer(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name            string
		initial         *File
		event           *commands.AnyMessage
		expectedBase    file.BaseFile
		expectedContent string
		expectedChunks  []file.Chunk
	}{
		{
			name:    "CreatedFile initializes file with chunks",
			initial: nil,
			event: th.NewDomainEvent(file.EntityName, "file-1", file.CreatedFile, &file.CreatedFilePayload{
				BaseFile: file.BaseFile{
					ProjectID: "project-1",
					Name:      "test.txt",
					Attributes: file.Attributes{
						Title:   "Test",
						Summary: "Summary",
						Time:    testTime,
					},
				},
				Content: "Short test content",
			}),
			expectedBase: file.BaseFile{
				ID:        "file-1",
				ProjectID: "project-1",
				Name:      "test.txt",
				Attributes: file.Attributes{
					Title:   "Test",
					Summary: "Summary",
					Time:    testTime,
				},
			},
			expectedContent: "Short test content",
			expectedChunks: []file.Chunk{
				{
					Content: "Short test content\n",
					Codes:   []file.CodedSection{},
				},
			},
		},
		{
			name:    "AppendCoding adds codes to chunk",
			initial: fileWithChunk("file-1", "project-1", "Climate change impacts global warming.", testTime),
			event: th.NewDomainEvent(file.EntityName, "file-1", file.CodedFile, &file.CodedFilePayload{
				Actions: []file.CodingAction{
					{
						CodeID:   "code-1",
						CodeSlug: "topic:climate",
						Action:   file.AppendCoding,
						ChunkID:  "chunk-id",
						Sections: []file.CodedSectionAttributes{
							{Text: "Climate change", AIReason: "Climate ref"},
						},
					},
				},
			}),
			expectedBase: file.BaseFile{
				ID:         "file-1",
				ProjectID:  "project-1",
				Name:       "test.txt",
				Attributes: file.Attributes{Time: testTime},
			},
			expectedContent: "Climate change impacts global warming.",
			expectedChunks: []file.Chunk{
				{
					Content: "Climate change impacts global warming.\n",
					Codes: []file.CodedSection{
						{StartIndex: 0, EndIndex: 14, CodeID: "code-1", CodeSlug: "topic:climate", CodedSectionAttributes: file.CodedSectionAttributes{Text: "Climate change", AIReason: "Climate ref"}},
					},
				},
			},
		},
		{
			name: "SetCoding replaces existing codes with same CodeID",
			initial: fileWithCodes("file-1", "project-1", "Climate change impacts global warming.", testTime, []file.CodedSection{
				{StartIndex: 0, EndIndex: 14, CodeID: "code-1", CodeSlug: "topic:climate", CodedSectionAttributes: file.CodedSectionAttributes{Text: "Climate change", AIReason: "Old"}},
				{StartIndex: 23, EndIndex: 38, CodeID: "code-1", CodeSlug: "topic:climate", CodedSectionAttributes: file.CodedSectionAttributes{Text: "global warming", AIReason: "Old"}},
			}),
			event: th.NewDomainEvent(file.EntityName, "file-1", file.CodedFile, &file.CodedFilePayload{
				Actions: []file.CodingAction{
					{
						CodeID:   "code-1",
						CodeSlug: "topic:climate",
						Action:   file.SetCoding,
						ChunkID:  "chunk-id",
						Sections: []file.CodedSectionAttributes{
							{Text: "warming", AIReason: "New"},
						},
					},
				},
			}),
			expectedBase: file.BaseFile{
				ID:         "file-1",
				ProjectID:  "project-1",
				Name:       "test.txt",
				Attributes: file.Attributes{Time: testTime},
			},
			expectedContent: "Climate change impacts global warming.",
			expectedChunks: []file.Chunk{
				{
					Content: "Climate change impacts global warming.\n",
					Codes: []file.CodedSection{
						{StartIndex: 30, EndIndex: 38, CodeID: "code-1", CodeSlug: "topic:climate", CodedSectionAttributes: file.CodedSectionAttributes{Text: "warming", AIReason: "New"}},
					},
				},
			},
		},
		{
			name: "RemoveCoding removes codes by CodeID",
			initial: fileWithCodes("file-1", "project-1", "Climate change impacts global warming.", testTime, []file.CodedSection{
				{StartIndex: 0, EndIndex: 14, CodeID: "code-1", CodeSlug: "topic:climate", CodedSectionAttributes: file.CodedSectionAttributes{Text: "Climate change"}},
				{StartIndex: 23, EndIndex: 38, CodeID: "code-2", CodeSlug: "topic:temperature", CodedSectionAttributes: file.CodedSectionAttributes{Text: "global warming"}},
			}),
			event: th.NewDomainEvent(file.EntityName, "file-1", file.CodedFile, &file.CodedFilePayload{
				Actions: []file.CodingAction{
					{
						CodeID:   "code-1",
						CodeSlug: "topic:climate",
						Action:   file.RemoveCoding,
						ChunkID:  "chunk-id",
						Sections: []file.CodedSectionAttributes{{Text: "dummy"}},
					},
				},
			}),
			expectedBase: file.BaseFile{
				ID:         "file-1",
				ProjectID:  "project-1",
				Name:       "test.txt",
				Attributes: file.Attributes{Time: testTime},
			},
			expectedContent: "Climate change impacts global warming.",
			expectedChunks: []file.Chunk{
				{
					Content: "Climate change impacts global warming.\n",
					Codes: []file.CodedSection{
						{StartIndex: 23, EndIndex: 38, CodeID: "code-2", CodeSlug: "topic:temperature", CodedSectionAttributes: file.CodedSectionAttributes{Text: "global warming"}},
					},
				},
			},
		},
		{
			name: "ClearedCoding removes all codes from all chunks",
			initial: fileWithCodes("file-1", "project-1", "Test content", testTime, []file.CodedSection{
				{StartIndex: 0, EndIndex: 4, CodeSlug: "topic:climate", CodedSectionAttributes: file.CodedSectionAttributes{Text: "Test"}},
			}),
			event: th.NewDomainEvent(file.EntityName, "file-1", file.ClearedCoding, nil),
			expectedBase: file.BaseFile{
				ID:         "file-1",
				ProjectID:  "project-1",
				Name:       "test.txt",
				Attributes: file.Attributes{Time: testTime},
			},
			expectedContent: "Test content",
			expectedChunks: []file.Chunk{
				{
					Content: "Test content\n",
					Codes:   []file.CodedSection{},
				},
			},
		},
		{
			name: "DeletedCode removes all codes with matching CodeID",
			initial: fileWithCodes("file-1", "project-1", "Climate change impacts global warming.", testTime, []file.CodedSection{
				{StartIndex: 0, EndIndex: 14, CodeID: "code-1", CodeSlug: "topic:climate", CodedSectionAttributes: file.CodedSectionAttributes{Text: "Climate change"}},
				{StartIndex: 23, EndIndex: 38, CodeID: "code-2", CodeSlug: "topic:temperature", CodedSectionAttributes: file.CodedSectionAttributes{Text: "global warming"}},
			}),
			event: th.NewDomainEvent(code.EntityName, "code-1", code.DeletedCode, &code.DeletedCodePayload{ProjectID: "project-1"}),
			expectedBase: file.BaseFile{
				ID:         "file-1",
				ProjectID:  "project-1",
				Name:       "test.txt",
				Attributes: file.Attributes{Time: testTime},
			},
			expectedContent: "Climate change impacts global warming.",
			expectedChunks: []file.Chunk{
				{
					Content: "Climate change impacts global warming.\n",
					Codes: []file.CodedSection{
						{StartIndex: 23, EndIndex: 38, CodeID: "code-2", CodeSlug: "topic:temperature", CodedSectionAttributes: file.CodedSectionAttributes{Text: "global warming"}},
					},
				},
			},
		},
		{
			name: "UpdatedCode changes slug for all codes with matching CodeID",
			initial: fileWithCodes("file-1", "project-1", "Climate change impacts global warming.", testTime, []file.CodedSection{
				{StartIndex: 0, EndIndex: 14, CodeID: "code-1", CodeSlug: "topic:climate-old", CodedSectionAttributes: file.CodedSectionAttributes{Text: "Climate change"}},
				{StartIndex: 23, EndIndex: 38, CodeID: "code-1", CodeSlug: "topic:climate-old", CodedSectionAttributes: file.CodedSectionAttributes{Text: "global warming"}},
			}),
			event: th.NewDomainEvent(code.EntityName, "code-1", code.UpdatedCode, &code.UpdatedCodePayload{Slug: "topic:climate-new"}),
			expectedBase: file.BaseFile{
				ID:         "file-1",
				ProjectID:  "project-1",
				Name:       "test.txt",
				Attributes: file.Attributes{Time: testTime},
			},
			expectedContent: "Climate change impacts global warming.",
			expectedChunks: []file.Chunk{
				{
					Content: "Climate change impacts global warming.\n",
					Codes: []file.CodedSection{
						{StartIndex: 0, EndIndex: 14, CodeID: "code-1", CodeSlug: "topic:climate-new", CodedSectionAttributes: file.CodedSectionAttributes{Text: "Climate change"}},
						{StartIndex: 23, EndIndex: 38, CodeID: "code-1", CodeSlug: "topic:climate-new", CodedSectionAttributes: file.CodedSectionAttributes{Text: "global warming"}},
					},
				},
			},
		},
		{
			name: "MergedCodes replaces source CodeID with target CodeID",
			initial: fileWithCodes("file-1", "project-1", "Climate change impacts global warming.", testTime, []file.CodedSection{
				{StartIndex: 0, EndIndex: 14, CodeID: "code-1", CodeSlug: "topic:climate", CodedSectionAttributes: file.CodedSectionAttributes{Text: "Climate change"}},
				{StartIndex: 23, EndIndex: 38, CodeID: "code-2", CodeSlug: "topic:temperature", CodedSectionAttributes: file.CodedSectionAttributes{Text: "global warming"}},
			}),
			event: th.NewDomainEvent(code.EntityName, "code-1", code.MergedCodes, &code.MergedCodesPayload{SourceID: "code-1", TargetID: "code-2"}),
			expectedBase: file.BaseFile{
				ID:         "file-1",
				ProjectID:  "project-1",
				Name:       "test.txt",
				Attributes: file.Attributes{Time: testTime},
			},
			expectedContent: "Climate change impacts global warming.",
			expectedChunks: []file.Chunk{
				{
					Content: "Climate change impacts global warming.\n",
					Codes: []file.CodedSection{
						{StartIndex: 0, EndIndex: 14, CodeID: "code-2", CodeSlug: "topic:climate", CodedSectionAttributes: file.CodedSectionAttributes{Text: "Climate change"}},
						{StartIndex: 23, EndIndex: 38, CodeID: "code-2", CodeSlug: "topic:temperature", CodedSectionAttributes: file.CodedSectionAttributes{Text: "global warming"}},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Reducer(tt.initial, tt.event)

			th.AssertEqualIgnoreFields(t, result.BaseFile, tt.expectedBase, "BaseFile", file.BaseFile{}, "ID")
			th.AssertEqualSimple(t, result.Content, tt.expectedContent)
			th.AssertEqualIgnoreFields(t, result.Chunks, tt.expectedChunks, "Chunks", file.Chunk{}, "ID")
		})
	}
}

func fileWithChunk(id, projectID, content string, testTime time.Time) *File {
	return &File{
		BaseFile: file.BaseFile{
			ID:         id,
			ProjectID:  projectID,
			Name:       "test.txt",
			Attributes: file.Attributes{Time: testTime},
		},
		Content: content,
		Chunks: []file.Chunk{
			{
				ID:      "chunk-id",
				Content: content + "\n",
				Codes:   []file.CodedSection{},
			},
		},
	}
}

func fileWithCodes(id, projectID, content string, testTime time.Time, codes []file.CodedSection) *File {
	return &File{
		BaseFile: file.BaseFile{
			ID:         id,
			ProjectID:  projectID,
			Name:       "test.txt",
			Attributes: file.Attributes{Time: testTime},
		},
		Content: content,
		Chunks: []file.Chunk{
			{
				ID:      "chunk-id",
				Content: content + "\n",
				Codes:   codes,
			},
		},
	}
}
