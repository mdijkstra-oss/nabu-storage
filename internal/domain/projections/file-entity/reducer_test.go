package fileview

import (
	"fmt"
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/domain/entities/code"
	"hermes-relay/internal/domain/entities/file"
	th "hermes-relay/internal/lib/test-helpers"
	"hermes-relay/internal/lib/test-helpers/domain-helpers"
	"hermes-relay/internal/lib/test-helpers/reducer-helpers"
	"testing"
)

func testActor() commands.Actor {
	return domain_helpers.TestActor()
}

func buildFile(id string, overrides file.FileData, chunks []file.Chunk) *File {
	f := file.BuildTestFile(id, overrides)
	if chunks != nil {
		f.Chunks = chunks
	}
	return &f
}

func createTestFile(projectID string) *File {
	return buildFile("file-1", file.FileData{ProjectID: projectID}, []file.Chunk{{ID: "1", Content: "Test content\n", Codes: []file.CodedSection{}}})
}

func TestFileReducer(t *testing.T) {
	testTime := th.DefaultTestTime()

	tests := []reducer_helpers.ReducerTestCase[*File]{
		{
			Name:    "CreatedFile initializes file with chunks from payload",
			Initial: nil,
			Event: domain_helpers.NewDomainEvent(file.EntityName, "file-1", file.CreatedFile, &file.CreatedFilePayload{
				FileData: file.FileData{
					ProjectID: "project-1",
					Name:      "test.txt",
					Type:      file.FileTypeSource,
					Locked:    true,
				},
				Chunks: []file.Chunk{
					{
						ID:      "1",
						Content: "Short test content",
						Codes:   []file.CodedSection{},
					},
				},
			}),
			Expected: buildFile("file-1", file.FileData{
				ProjectID: "project-1",
				Name:      "test.txt",
				Type:      file.FileTypeSource,
				Locked:    true,
			}, []file.Chunk{
				{
					ID:      "1",
					Content: "Short test content",
					Codes:   []file.CodedSection{},
				},
			}),
		},
		{
			Name: "UpdatedFile updates name and description",
			Initial: buildFile("file-1", file.FileData{
				Name:        "old-name.txt",
				Description: "Old description",
			}, []file.Chunk{{ID: "1", Content: "Test content\n", Codes: []file.CodedSection{}}}),
			Event: domain_helpers.NewDomainEvent(file.EntityName, "file-1", file.UpdatedFile, &file.UpdatedFilePayload{
				Name:        "new-name.txt",
				Description: "New description",
			}),
			Expected: buildFile("file-1", file.FileData{
				Name:        "new-name.txt",
				Description: "New description",
			}, []file.Chunk{{ID: "1", Content: "Test content\n", Codes: []file.CodedSection{}}}),
		},
		{
			Name: "UpdatedFile with empty description preserves existing description",
			Initial: buildFile("file-1", file.FileData{
				Name:        "interview-transcript.txt",
				Description: "Interview with participant 023 discussing their experience with the pandemic response",
			}, []file.Chunk{{ID: "1", Content: "Full interview content here\n", Codes: []file.CodedSection{}}}),
			Event: domain_helpers.NewDomainEvent(file.EntityName, "file-1", file.UpdatedFile, &file.UpdatedFilePayload{
				Name:        "interview-023.txt",
				Description: "",
			}),
			Expected: buildFile("file-1", file.FileData{
				Name:        "interview-023.txt",
				Description: "Interview with participant 023 discussing their experience with the pandemic response",
			}, []file.Chunk{{ID: "1", Content: "Full interview content here\n", Codes: []file.CodedSection{}}}),
		},
		{
			Name:     "UpdatedFile on nil state returns nil",
			Initial:  nil,
			Event:    domain_helpers.NewDomainEvent(file.EntityName, "file-1", file.UpdatedFile, &file.UpdatedFilePayload{Name: "test.txt"}),
			Expected: nil,
		},
		{
			Name: "AddedCodeSections adds sections to chunk with LastActor",
			Initial: buildFile("file-1", file.FileData{}, []file.Chunk{
				{
					ID:      "1",
					Content: "Climate change impacts global warming.\n",
					Codes:   []file.CodedSection{},
				},
			}),
			Event: domain_helpers.NewDomainEvent(file.EntityName, "file-1", file.AddedCodeSections, &file.AddedCodeSectionsPayload{
				ChunkID: "1",
				Sections: []file.AddedSection{
					{ID: "section-id-1", CodeID: "code-1", CodeSlug: "topic:climate", Text: "Climate change impacts", Reason: "Climate ref"},
				},
			}),
			Expected: buildFile("file-1", file.FileData{}, []file.Chunk{
				{
					ID:      "1",
					Content: "Climate change impacts global warming.\n",
					Codes: []file.CodedSection{
						{ID: "generated-id-1", CodeID: "code-1", CodeSlug: "topic:climate", Text: "Climate change impacts", Reason: "Climate ref", LastActor: testActor()},
					},
				},
			}),
		},
		{
			Name: "UpdatedCodeSections updates text reason and LastActor",
			Initial: buildFile("file-1", file.FileData{}, []file.Chunk{
				{
					ID:      "1",
					Content: "Climate change impacts global warming.\n",
					Codes: []file.CodedSection{
						{ID: "section-1", CodeID: "code-1", CodeSlug: "topic:climate", Text: "Climate change impacts", Reason: "Old"},
						{ID: "section-2", CodeID: "code-1", CodeSlug: "topic:climate", Text: "impacts global warming", Reason: "Old"},
					},
				},
			}),
			Event: domain_helpers.NewDomainEvent(file.EntityName, "file-1", file.UpdatedCodeSections, &file.UpdateCodeSectionsPayload{
				ChunkID: "1",
				Sections: []file.UpdateSectionOp{
					{ID: "section-1", Text: "Climate change impacts", Reason: "New reason"},
				},
			}),
			Expected: buildFile("file-1", file.FileData{}, []file.Chunk{
				{
					ID:      "1",
					Content: "Climate change impacts global warming.\n",
					Codes: []file.CodedSection{
						{ID: "section-1", CodeID: "code-1", CodeSlug: "topic:climate", Text: "Climate change impacts", Reason: "New reason", LastActor: testActor()},
						{ID: "section-2", CodeID: "code-1", CodeSlug: "topic:climate", Text: "impacts global warming", Reason: "Old"},
					},
				},
			}),
		},
		{
			Name: "RemovedCodeSections removes sections by ID",
			Initial: buildFile("file-1", file.FileData{}, []file.Chunk{
				{
					ID:      "1",
					Content: "Climate change impacts global warming.\n",
					Codes: []file.CodedSection{
						{ID: "section-1", CodeID: "code-1", CodeSlug: "topic:climate", Text: "Climate change impacts"},
						{ID: "section-2", CodeID: "code-2", CodeSlug: "topic:temperature", Text: "impacts global warming"},
					},
				},
			}),
			Event: domain_helpers.NewDomainEvent(file.EntityName, "file-1", file.RemovedCodeSections, &file.RemoveCodeSectionsPayload{
				ChunkID:    "1",
				SectionIDs: []string{"section-1"},
			}),
			Expected: buildFile("file-1", file.FileData{}, []file.Chunk{
				{
					ID:      "1",
					Content: "Climate change impacts global warming.\n",
					Codes: []file.CodedSection{
						{ID: "section-2", CodeID: "code-2", CodeSlug: "topic:temperature", Text: "impacts global warming"},
					},
				},
			}),
		},
		{
			Name: "RemovedCodeSections removes middle section with same code",
			Initial: buildFile("file-1", file.FileData{}, []file.Chunk{
				{
					ID:      "1",
					Content: "Climate change impacts global warming and causes temperature rise.\n",
					Codes: []file.CodedSection{
						{ID: "section-1", CodeID: "code-1", CodeSlug: "topic:climate", Text: "Climate change impacts"},
						{ID: "section-2", CodeID: "code-1", CodeSlug: "topic:climate", Text: "impacts global warming"},
						{ID: "section-3", CodeID: "code-1", CodeSlug: "topic:climate", Text: "causes temperature rise"},
					},
				},
			}),
			Event: domain_helpers.NewDomainEvent(file.EntityName, "file-1", file.RemovedCodeSections, &file.RemoveCodeSectionsPayload{
				ChunkID:    "1",
				SectionIDs: []string{"section-2"},
			}),
			Expected: buildFile("file-1", file.FileData{}, []file.Chunk{
				{
					ID:      "1",
					Content: "Climate change impacts global warming and causes temperature rise.\n",
					Codes: []file.CodedSection{
						{ID: "section-1", CodeID: "code-1", CodeSlug: "topic:climate", Text: "Climate change impacts"},
						{ID: "section-3", CodeID: "code-1", CodeSlug: "topic:climate", Text: "causes temperature rise"},
					},
				},
			}),
		},
		{
			Name: "RemovedCodeSections removes last section with same code",
			Initial: buildFile("file-1", file.FileData{}, []file.Chunk{
				{
					ID:      "1",
					Content: "Climate change impacts global warming and causes temperature rise.\n",
					Codes: []file.CodedSection{
						{ID: "section-1", CodeID: "code-1", CodeSlug: "topic:climate", Text: "Climate change impacts"},
						{ID: "section-2", CodeID: "code-1", CodeSlug: "topic:climate", Text: "impacts global warming"},
						{ID: "section-3", CodeID: "code-1", CodeSlug: "topic:climate", Text: "causes temperature rise"},
					},
				},
			}),
			Event: domain_helpers.NewDomainEvent(file.EntityName, "file-1", file.RemovedCodeSections, &file.RemoveCodeSectionsPayload{
				ChunkID:    "1",
				SectionIDs: []string{"section-3"},
			}),
			Expected: buildFile("file-1", file.FileData{}, []file.Chunk{
				{
					ID:      "1",
					Content: "Climate change impacts global warming and causes temperature rise.\n",
					Codes: []file.CodedSection{
						{ID: "section-1", CodeID: "code-1", CodeSlug: "topic:climate", Text: "Climate change impacts"},
						{ID: "section-2", CodeID: "code-1", CodeSlug: "topic:climate", Text: "impacts global warming"},
					},
				},
			}),
		},
		{
			Name: "RemovedCodeSections removes multiple specific sections",
			Initial: buildFile("file-1", file.FileData{}, []file.Chunk{
				{
					ID:      "1",
					Content: "Climate change impacts global warming and causes temperature rise.\n",
					Codes: []file.CodedSection{
						{ID: "section-1", CodeID: "code-1", CodeSlug: "topic:climate", Text: "Climate change impacts"},
						{ID: "section-2", CodeID: "code-1", CodeSlug: "topic:climate", Text: "impacts global warming"},
						{ID: "section-3", CodeID: "code-2", CodeSlug: "topic:temperature", Text: "global warming"},
						{ID: "section-4", CodeID: "code-1", CodeSlug: "topic:climate", Text: "causes temperature rise"},
					},
				},
			}),
			Event: domain_helpers.NewDomainEvent(file.EntityName, "file-1", file.RemovedCodeSections, &file.RemoveCodeSectionsPayload{
				ChunkID:    "1",
				SectionIDs: []string{"section-2", "section-4"},
			}),
			Expected: buildFile("file-1", file.FileData{}, []file.Chunk{
				{
					ID:      "1",
					Content: "Climate change impacts global warming and causes temperature rise.\n",
					Codes: []file.CodedSection{
						{ID: "section-1", CodeID: "code-1", CodeSlug: "topic:climate", Text: "Climate change impacts"},
						{ID: "section-3", CodeID: "code-2", CodeSlug: "topic:temperature", Text: "global warming"},
					},
				},
			}),
		},
		{
			Name: "ClearedCoding removes all codes from all chunks",
			Initial: buildFile("file-1", file.FileData{}, []file.Chunk{
				{
					ID:      "1",
					Content: "Test content\n",
					Codes: []file.CodedSection{
						{ID: "section-1", CodeSlug: "topic:climate", Text: "Test"},
					},
				},
			}),
			Event: domain_helpers.NewDomainEvent(file.EntityName, "file-1", file.ClearedCoding, nil),
			Expected: buildFile("file-1", file.FileData{}, []file.Chunk{
				{
					ID:      "1",
					Content: "Test content\n",
					Codes:   []file.CodedSection{},
				},
			}),
		},
		{
			Name: "DeletedCode removes all codes with matching CodeID",
			Initial: buildFile("file-1", file.FileData{}, []file.Chunk{
				{
					ID:      "1",
					Content: "Climate change impacts global warming.\n",
					Codes: []file.CodedSection{
						{ID: "section-1", CodeID: "code-1", CodeSlug: "topic:climate", Text: "Climate change impacts"},
						{ID: "section-2", CodeID: "code-2", CodeSlug: "topic:temperature", Text: "impacts global warming"},
					},
				},
			}),
			Event: domain_helpers.NewDomainEvent(code.EntityName, "code-1", code.DeletedCode, nil),
			Expected: buildFile("file-1", file.FileData{}, []file.Chunk{
				{
					ID:      "1",
					Content: "Climate change impacts global warming.\n",
					Codes: []file.CodedSection{
						{ID: "section-2", CodeID: "code-2", CodeSlug: "topic:temperature", Text: "impacts global warming"},
					},
				},
			}),
		},
		{
			Name: "UpdatedCode changes slug for all codes with matching CodeID",
			Initial: buildFile("file-1", file.FileData{}, []file.Chunk{
				{
					ID:      "1",
					Content: "Climate change impacts global warming.\n",
					Codes: []file.CodedSection{
						{ID: "section-1", CodeID: "code-1", CodeSlug: "topic:climate-old", Text: "Climate change impacts"},
						{ID: "section-2", CodeID: "code-1", CodeSlug: "topic:climate-old", Text: "impacts global warming"},
					},
				},
			}),
			Event: domain_helpers.NewDomainEvent(code.EntityName, "code-1", code.UpdatedCode, &code.UpdatedCodePayload{Slug: "topic:climate-new"}),
			Expected: buildFile("file-1", file.FileData{}, []file.Chunk{
				{
					ID:      "1",
					Content: "Climate change impacts global warming.\n",
					Codes: []file.CodedSection{
						{ID: "section-1", CodeID: "code-1", CodeSlug: "topic:climate-new", Text: "Climate change impacts"},
						{ID: "section-2", CodeID: "code-1", CodeSlug: "topic:climate-new", Text: "impacts global warming"},
					},
				},
			}),
		},
		{
			Name: "MergedCodes replaces source CodeID with target CodeID",
			Initial: buildFile("file-1", file.FileData{}, []file.Chunk{
				{
					ID:      "1",
					Content: "Climate change impacts global warming.\n",
					Codes: []file.CodedSection{
						{ID: "section-1", CodeID: "code-1", CodeSlug: "topic:climate", Text: "Climate change impacts"},
						{ID: "section-2", CodeID: "code-2", CodeSlug: "topic:temperature", Text: "impacts global warming"},
					},
				},
			}),
			Event: domain_helpers.NewDomainEvent(code.EntityName, "code-1", code.MergedCodes, &code.MergedCodesPayload{SourceID: "code-1", TargetID: "code-2"}),
			Expected: buildFile("file-1", file.FileData{}, []file.Chunk{
				{
					ID:      "1",
					Content: "Climate change impacts global warming.\n",
					Codes: []file.CodedSection{
						{ID: "section-1", CodeID: "code-2", CodeSlug: "topic:climate", Text: "Climate change impacts"},
						{ID: "section-2", CodeID: "code-2", CodeSlug: "topic:temperature", Text: "impacts global warming"},
					},
				},
			}),
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

	// Normalize Time and ID fields for comparison
	normalize := func(f *File) *File {
		if f != nil {
			f.Time = testTime
			counter := 0
			for i := range f.Chunks {
				for j := range f.Chunks[i].Codes {
					id := f.Chunks[i].Codes[j].ID
					var keepID bool
					if id == "" {
						keepID = true
					} else if len(id) >= 9 && id[:8] == "section-" && id[8] >= '0' && id[8] <= '9' {
						keepID = true
					} else if len(id) >= 14 && id[:13] == "generated-id-" {
						keepID = true
					}

					if !keepID {
						counter++
						f.Chunks[i].Codes[j].ID = fmt.Sprintf("generated-id-%d", counter)
					}
				}
			}
		}
		return f
	}

	reducer_helpers.RunReducerTests(t, combinedTests, Reducer, normalize)
}
