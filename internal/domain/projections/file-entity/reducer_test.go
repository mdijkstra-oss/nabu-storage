package fileview

import (
	"hermes-relay/internal/domain/entities/code"
	"hermes-relay/internal/domain/entities/file"
	"hermes-relay/internal/lib/test-helpers/domain-helpers"
	"hermes-relay/internal/lib/test-helpers/reducer-helpers"
	"testing"
	"time"
)

func createTestFile(projectID string) *File {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	return &File{
		BaseFile: file.BaseFile{
			ID:         "file-1",
			ProjectID:  projectID,
			Name:       "test.txt",
			Attributes: file.Attributes{Time: testTime},
			Healthy:    true,
		},
		Chunks: []file.Chunk{{ID: "1", Content: "Test content\n", Codes: []file.CodedSection{}}},
	}
}

func TestFileReducer(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []reducer_helpers.ReducerTestCase[*File]{
		{
			Name:    "CreatedFile initializes file with chunks from payload",
			Initial: nil,
			Event: domain_helpers.NewDomainEvent(file.EntityName, "file-1", file.CreatedFile, &file.CreatedFilePayload{
				CreateFilePayload: file.CreateFilePayload{
					ProjectID: "project-1",
					Name:      "test.txt",
					Content:   "Short test content",
				},
				Type:   file.FileTypeSource,
				Locked: true,
				Chunks: []file.Chunk{
					{
						ID:      "1",
						Content: "Short test content",
						Codes:   []file.CodedSection{},
					},
				},
			}),
			Expected: &File{
				BaseFile: file.BaseFile{
					ID:        "file-1",
					ProjectID: "project-1",
					Name:      "test.txt",
					Attributes: file.Attributes{
						Title:   "",
						Summary: "",
						Time:    testTime,
						Type:    file.FileTypeSource,
						Locked:  true,
					},
					Healthy: true,
				},
				Chunks: []file.Chunk{
					{
						ID:      "1",
						Content: "Short test content",
						Codes:   []file.CodedSection{},
					},
				},
			},
		},
		{
			Name: "UpdatedFile updates name and description",
			Initial: &File{
				BaseFile: file.BaseFile{
					ID:          "file-1",
					ProjectID:   "project-1",
					Name:        "old-name.txt",
					Description: "Old description",
					Attributes:  file.Attributes{Time: testTime},
					Healthy:     true,
				},
				Chunks: []file.Chunk{{ID: "1", Content: "Test content\n", Codes: []file.CodedSection{}}},
			},
			Event: domain_helpers.NewDomainEvent(file.EntityName, "file-1", file.UpdatedFile, &file.UpdatedFilePayload{
				Name:        "new-name.txt",
				Description: "New description",
			}),
			Expected: &File{
				BaseFile: file.BaseFile{
					ID:          "file-1",
					ProjectID:   "project-1",
					Name:        "new-name.txt",
					Description: "New description",
					Attributes:  file.Attributes{Time: testTime},
					Healthy:     true,
				},
				Chunks: []file.Chunk{{ID: "1", Content: "Test content\n", Codes: []file.CodedSection{}}},
			},
		},
		{
			Name: "UpdatedFile with empty description preserves existing description",
			Initial: &File{
				BaseFile: file.BaseFile{
					ID:          "file-1",
					ProjectID:   "project-1",
					Name:        "interview-transcript.txt",
					Description: "Interview with participant 023 discussing their experience with the pandemic response",
					Attributes:  file.Attributes{Time: testTime},
					Healthy:     true,
				},
				Chunks: []file.Chunk{{ID: "1", Content: "Full interview content here\n", Codes: []file.CodedSection{}}},
			},
			Event: domain_helpers.NewDomainEvent(file.EntityName, "file-1", file.UpdatedFile, &file.UpdatedFilePayload{
				Name:        "interview-023.txt",
				Description: "",
			}),
			Expected: &File{
				BaseFile: file.BaseFile{
					ID:          "file-1",
					ProjectID:   "project-1",
					Name:        "interview-023.txt",
					Description: "Interview with participant 023 discussing their experience with the pandemic response",
					Attributes:  file.Attributes{Time: testTime},
					Healthy:     true,
				},
				Chunks: []file.Chunk{{ID: "1", Content: "Full interview content here\n", Codes: []file.CodedSection{}}},
			},
		},
		{
			Name:     "UpdatedFile on nil state returns nil",
			Initial:  nil,
			Event:    domain_helpers.NewDomainEvent(file.EntityName, "file-1", file.UpdatedFile, &file.UpdatedFilePayload{Name: "test.txt"}),
			Expected: nil,
		},
		{
			Name: "AppendCoding adds codes to chunk",
			Initial: &File{
				BaseFile: file.BaseFile{
					ID:         "file-1",
					ProjectID:  "project-1",
					Name:       "test.txt",
					Attributes: file.Attributes{Time: testTime},
					Healthy:    true,
				},
				Chunks: []file.Chunk{
					{
						ID:      "1",
						Content: "Climate change impacts global warming.\n",
						Codes:   []file.CodedSection{},
					},
				},
			},
			Event: domain_helpers.NewDomainEvent(file.EntityName, "file-1", file.CodedFile, &file.CodedFilePayload{
				Actions: []file.CodingAction{
					{
						CodeID:   "code-1",
						CodeSlug: "topic:climate",
						Action:   file.AppendCoding,
						ChunkID:  "1",
						Sections: []file.CodedSectionAttributes{
							{Text: "Climate change", Reason: "Climate ref"},
						},
					},
				},
			}),
			Expected: &File{
				BaseFile: file.BaseFile{
					ID:         "file-1",
					ProjectID:  "project-1",
					Name:       "test.txt",
					Attributes: file.Attributes{Time: testTime},
					Healthy:    true,
				},
				Chunks: []file.Chunk{
					{
						ID:      "1",
						Content: "Climate change impacts global warming.\n",
						Codes: []file.CodedSection{
							{CodeID: "code-1", CodeSlug: "topic:climate", CodedSectionAttributes: file.CodedSectionAttributes{Text: "Climate change", Reason: "Climate ref"}},
						},
					},
				},
			},
		},
		{
			Name: "SetCoding replaces existing codes with same CodeID",
			Initial: &File{
				BaseFile: file.BaseFile{
					ID:         "file-1",
					ProjectID:  "project-1",
					Name:       "test.txt",
					Attributes: file.Attributes{Time: testTime},
					Healthy:    true,
				},
				Chunks: []file.Chunk{
					{
						ID:      "1",
						Content: "Climate change impacts global warming.\n",
						Codes: []file.CodedSection{
							{CodeID: "code-1", CodeSlug: "topic:climate", CodedSectionAttributes: file.CodedSectionAttributes{Text: "Climate change", Reason: "Old"}},
							{CodeID: "code-1", CodeSlug: "topic:climate", CodedSectionAttributes: file.CodedSectionAttributes{Text: "global warming", Reason: "Old"}},
						},
					},
				},
			},
			Event: domain_helpers.NewDomainEvent(file.EntityName, "file-1", file.CodedFile, &file.CodedFilePayload{
				Actions: []file.CodingAction{
					{
						CodeID:   "code-1",
						CodeSlug: "topic:climate",
						Action:   file.SetCoding,
						ChunkID:  "1",
						Sections: []file.CodedSectionAttributes{
							{Text: "warming", Reason: "New"},
						},
					},
				},
			}),
			Expected: &File{
				BaseFile: file.BaseFile{
					ID:         "file-1",
					ProjectID:  "project-1",
					Name:       "test.txt",
					Attributes: file.Attributes{Time: testTime},
					Healthy:    true,
				},
				Chunks: []file.Chunk{
					{
						ID:      "1",
						Content: "Climate change impacts global warming.\n",
						Codes: []file.CodedSection{
							{CodeID: "code-1", CodeSlug: "topic:climate", CodedSectionAttributes: file.CodedSectionAttributes{Text: "warming", Reason: "New"}},
						},
					},
				},
			},
		},
		{
			Name: "RemoveCoding removes codes by CodeID",
			Initial: &File{
				BaseFile: file.BaseFile{
					ID:         "file-1",
					ProjectID:  "project-1",
					Name:       "test.txt",
					Attributes: file.Attributes{Time: testTime},
					Healthy:    true,
				},
				Chunks: []file.Chunk{
					{
						ID:      "1",
						Content: "Climate change impacts global warming.\n",
						Codes: []file.CodedSection{
							{CodeID: "code-1", CodeSlug: "topic:climate", CodedSectionAttributes: file.CodedSectionAttributes{Text: "Climate change"}},
							{CodeID: "code-2", CodeSlug: "topic:temperature", CodedSectionAttributes: file.CodedSectionAttributes{Text: "global warming"}},
						},
					},
				},
			},
			Event: domain_helpers.NewDomainEvent(file.EntityName, "file-1", file.CodedFile, &file.CodedFilePayload{
				Actions: []file.CodingAction{
					{
						CodeID:   "code-1",
						CodeSlug: "topic:climate",
						Action:   file.RemoveCoding,
						ChunkID:  "1",
						Sections: []file.CodedSectionAttributes{{Text: "dummy"}},
					},
				},
			}),
			Expected: &File{
				BaseFile: file.BaseFile{
					ID:         "file-1",
					ProjectID:  "project-1",
					Name:       "test.txt",
					Attributes: file.Attributes{Time: testTime},
					Healthy:    true,
				},
				Chunks: []file.Chunk{
					{
						ID:      "1",
						Content: "Climate change impacts global warming.\n",
						Codes: []file.CodedSection{
							{CodeID: "code-2", CodeSlug: "topic:temperature", CodedSectionAttributes: file.CodedSectionAttributes{Text: "global warming"}},
						},
					},
				},
			},
		},
		{
			Name: "ClearedCoding removes all codes from all chunks",
			Initial: &File{
				BaseFile: file.BaseFile{
					ID:         "file-1",
					ProjectID:  "project-1",
					Name:       "test.txt",
					Attributes: file.Attributes{Time: testTime},
					Healthy:    true,
				},
				Chunks: []file.Chunk{
					{
						ID:      "1",
						Content: "Test content\n",
						Codes: []file.CodedSection{
							{CodeSlug: "topic:climate", CodedSectionAttributes: file.CodedSectionAttributes{Text: "Test"}},
						},
					},
				},
			},
			Event: domain_helpers.NewDomainEvent(file.EntityName, "file-1", file.ClearedCoding, nil),
			Expected: &File{
				BaseFile: file.BaseFile{
					ID:         "file-1",
					ProjectID:  "project-1",
					Name:       "test.txt",
					Attributes: file.Attributes{Time: testTime},
					Healthy:    true,
				},
				Chunks: []file.Chunk{
					{
						ID:      "1",
						Content: "Test content\n",
						Codes:   []file.CodedSection{},
					},
				},
			},
		},
		{
			Name: "DeletedCode removes all codes with matching CodeID",
			Initial: &File{
				BaseFile: file.BaseFile{
					ID:         "file-1",
					ProjectID:  "project-1",
					Name:       "test.txt",
					Attributes: file.Attributes{Time: testTime},
					Healthy:    true,
				},
				Chunks: []file.Chunk{
					{
						ID:      "1",
						Content: "Climate change impacts global warming.\n",
						Codes: []file.CodedSection{
							{CodeID: "code-1", CodeSlug: "topic:climate", CodedSectionAttributes: file.CodedSectionAttributes{Text: "Climate change"}},
							{CodeID: "code-2", CodeSlug: "topic:temperature", CodedSectionAttributes: file.CodedSectionAttributes{Text: "global warming"}},
						},
					},
				},
			},
			Event: domain_helpers.NewDomainEvent(code.EntityName, "code-1", code.DeletedCode, &code.DeletedCodePayload{ProjectID: "project-1"}),
			Expected: &File{
				BaseFile: file.BaseFile{
					ID:         "file-1",
					ProjectID:  "project-1",
					Name:       "test.txt",
					Attributes: file.Attributes{Time: testTime},
					Healthy:    true,
				},
				Chunks: []file.Chunk{
					{
						ID:      "1",
						Content: "Climate change impacts global warming.\n",
						Codes: []file.CodedSection{
							{CodeID: "code-2", CodeSlug: "topic:temperature", CodedSectionAttributes: file.CodedSectionAttributes{Text: "global warming"}},
						},
					},
				},
			},
		},
		{
			Name: "UpdatedCode changes slug for all codes with matching CodeID",
			Initial: &File{
				BaseFile: file.BaseFile{
					ID:         "file-1",
					ProjectID:  "project-1",
					Name:       "test.txt",
					Attributes: file.Attributes{Time: testTime},
					Healthy:    true,
				},
				Chunks: []file.Chunk{
					{
						ID:      "1",
						Content: "Climate change impacts global warming.\n",
						Codes: []file.CodedSection{
							{CodeID: "code-1", CodeSlug: "topic:climate-old", CodedSectionAttributes: file.CodedSectionAttributes{Text: "Climate change"}},
							{CodeID: "code-1", CodeSlug: "topic:climate-old", CodedSectionAttributes: file.CodedSectionAttributes{Text: "global warming"}},
						},
					},
				},
			},
			Event: domain_helpers.NewDomainEvent(code.EntityName, "code-1", code.UpdatedCode, &code.UpdatedCodePayload{Slug: "topic:climate-new"}),
			Expected: &File{
				BaseFile: file.BaseFile{
					ID:         "file-1",
					ProjectID:  "project-1",
					Name:       "test.txt",
					Attributes: file.Attributes{Time: testTime},
					Healthy:    true,
				},
				Chunks: []file.Chunk{
					{
						ID:      "1",
						Content: "Climate change impacts global warming.\n",
						Codes: []file.CodedSection{
							{CodeID: "code-1", CodeSlug: "topic:climate-new", CodedSectionAttributes: file.CodedSectionAttributes{Text: "Climate change"}},
							{CodeID: "code-1", CodeSlug: "topic:climate-new", CodedSectionAttributes: file.CodedSectionAttributes{Text: "global warming"}},
						},
					},
				},
			},
		},
		{
			Name: "MergedCodes replaces source CodeID with target CodeID",
			Initial: &File{
				BaseFile: file.BaseFile{
					ID:         "file-1",
					ProjectID:  "project-1",
					Name:       "test.txt",
					Attributes: file.Attributes{Time: testTime},
					Healthy:    true,
				},
				Chunks: []file.Chunk{
					{
						ID:      "1",
						Content: "Climate change impacts global warming.\n",
						Codes: []file.CodedSection{
							{CodeID: "code-1", CodeSlug: "topic:climate", CodedSectionAttributes: file.CodedSectionAttributes{Text: "Climate change"}},
							{CodeID: "code-2", CodeSlug: "topic:temperature", CodedSectionAttributes: file.CodedSectionAttributes{Text: "global warming"}},
						},
					},
				},
			},
			Event: domain_helpers.NewDomainEvent(code.EntityName, "code-1", code.MergedCodes, &code.MergedCodesPayload{SourceID: "code-1", TargetID: "code-2"}),
			Expected: &File{
				BaseFile: file.BaseFile{
					ID:         "file-1",
					ProjectID:  "project-1",
					Name:       "test.txt",
					Attributes: file.Attributes{Time: testTime},
					Healthy:    true,
				},
				Chunks: []file.Chunk{
					{
						ID:      "1",
						Content: "Climate change impacts global warming.\n",
						Codes: []file.CodedSection{
							{CodeID: "code-2", CodeSlug: "topic:climate", CodedSectionAttributes: file.CodedSectionAttributes{Text: "Climate change"}},
							{CodeID: "code-2", CodeSlug: "topic:temperature", CodedSectionAttributes: file.CodedSectionAttributes{Text: "global warming"}},
						},
					},
				},
			},
		},
	}

	deletedEntityTests := reducer_helpers.DeletedEntityTests(
		file.EntityName,
		file.DeletedFile,
		func() *File { return createTestFile("project-1") },
	)
	deletedProjectTests := reducer_helpers.DeletedProjectCascadeTests(createTestFile)

	combinedTests := append(tests, deletedEntityTests...)
	combinedTests = append(combinedTests, deletedProjectTests...)

	// Normalize Time field to testTime for comparison
	normalizeTime := func(f *File) *File {
		if f != nil {
			f.Time = testTime
		}
		return f
	}

	reducer_helpers.RunReducerTests(t, combinedTests, Reducer, normalizeTime)
}
