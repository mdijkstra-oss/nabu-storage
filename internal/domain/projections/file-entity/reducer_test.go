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

var (
	confidenceHigh = file.ConfidenceHigh
)

func buildFile(id string, overrides file.FileData, content string, codes []file.CodedSection) *File {
	f := file.BuildTestFile(id, overrides)
	f.Content = content
	f.Codes = codes
	return &f
}

func createTestFile(projectID string) *File {
	return buildFile("file-1", file.FileData{ProjectID: projectID}, "Test content\n", []file.CodedSection{})
}

func TestFileReducer(t *testing.T) {
	testTime := th.DefaultTestTime()

	tests := []reducer_helpers.ReducerTestCase[*File]{
		{
			Name:    "CreatedFile initializes file with content from payload",
			Initial: nil,
			Event: domain_helpers.NewDomainEvent(file.EntityName, "file-1", file.CreatedFile, &file.CreatedFilePayload{
				FileData: file.FileData{
					ProjectID: "project-1",
					Name:      "test.txt",
					Type:      file.FileTypeCorpus,
					Locked:    true,
				},
				Content: "Short test content",
				Codes:   []file.CodedSection{},
			}),
			Expected: buildFile("file-1", file.FileData{
				ProjectID: "project-1",
				Name:      "test.txt",
				Type:      file.FileTypeCorpus,
				Locked:    true,
			}, "Short test content", []file.CodedSection{}),
		},
		{
			Name: "UpdatedFile updates name and description",
			Initial: buildFile("file-1", file.FileData{
				Name:        "old-name.txt",
				Description: "Old description",
			}, "Test content\n", []file.CodedSection{}),
			Event: domain_helpers.NewDomainEvent(file.EntityName, "file-1", file.UpdatedFile, &file.UpdatedFilePayload{
				Name:        "new-name.txt",
				Description: "New description",
			}),
			Expected: buildFile("file-1", file.FileData{
				Name:        "new-name.txt",
				Description: "New description",
			}, "Test content\n", []file.CodedSection{}),
		},
		{
			Name: "UpdatedFile with empty description preserves existing description",
			Initial: buildFile("file-1", file.FileData{
				Name:        "interview-transcript.txt",
				Description: "Interview with participant 023 discussing their experience with the pandemic response",
			}, "Full interview content here\n", []file.CodedSection{}),
			Event: domain_helpers.NewDomainEvent(file.EntityName, "file-1", file.UpdatedFile, &file.UpdatedFilePayload{
				Name:        "interview-023.txt",
				Description: "",
			}),
			Expected: buildFile("file-1", file.FileData{
				Name:        "interview-023.txt",
				Description: "Interview with participant 023 discussing their experience with the pandemic response",
			}, "Full interview content here\n", []file.CodedSection{}),
		},
		{
			Name:     "UpdatedFile on nil state returns nil",
			Initial:  nil,
			Event:    domain_helpers.NewDomainEvent(file.EntityName, "file-1", file.UpdatedFile, &file.UpdatedFilePayload{Name: "test.txt"}),
			Expected: nil,
		},
		{
			Name:    "AddedCodeSections adds sections to file with LastActor",
			Initial: buildFile("file-1", file.FileData{}, "Climate change impacts global warming.\n", []file.CodedSection{}),
			Event: domain_helpers.NewDomainEvent(file.EntityName, "file-1", file.ModifiedCodeSections, &file.ModifiedCodeSectionsPayload{
				Operations: []file.SectionOp{
					{Op: "add", ID: "section-id-1", CodeID: "code-1", Text: "Climate change impacts", Reason: "Climate ref", Confidence: &confidenceHigh},
				},
			}),
			Expected: buildFile("file-1", file.FileData{}, "Climate change impacts global warming.\n", []file.CodedSection{
				{ID: "generated-id-1", CodeID: "code-1", Text: "Climate change impacts", Reason: "Climate ref", Confidence: file.ConfidenceHigh, LastActor: testActor()},
			}),
		},
		{
			Name: "UpdatedCodeSections updates text reason and LastActor",
			Initial: buildFile("file-1", file.FileData{}, "Climate change impacts global warming.\n", []file.CodedSection{
				{ID: "section-1", CodeID: "code-1", Text: "Climate change impacts", Reason: "Old"},
				{ID: "section-2", CodeID: "code-1", Text: "impacts global warming", Reason: "Old"},
			}),
			Event: domain_helpers.NewDomainEvent(file.EntityName, "file-1", file.ModifiedCodeSections, &file.ModifiedCodeSectionsPayload{
				Operations: []file.SectionOp{
					{Op: "update", ID: "section-1", Text: "Climate change impacts", Reason: "New reason"},
				},
			}),
			Expected: buildFile("file-1", file.FileData{}, "Climate change impacts global warming.\n", []file.CodedSection{
				{ID: "section-1", CodeID: "code-1", Text: "Climate change impacts", Reason: "New reason", LastActor: testActor()},
				{ID: "section-2", CodeID: "code-1", Text: "impacts global warming", Reason: "Old"},
			}),
		},
		{
			Name: "RemovedCodeSections removes sections by ID",
			Initial: buildFile("file-1", file.FileData{}, "Climate change impacts global warming.\n", []file.CodedSection{
				{ID: "section-1", CodeID: "code-1", Text: "Climate change impacts"},
				{ID: "section-2", CodeID: "code-2", Text: "impacts global warming"},
			}),
			Event: domain_helpers.NewDomainEvent(file.EntityName, "file-1", file.ModifiedCodeSections, &file.ModifiedCodeSectionsPayload{
				Operations: []file.SectionOp{
					{Op: "delete", ID: "section-1"},
				},
			}),
			Expected: buildFile("file-1", file.FileData{}, "Climate change impacts global warming.\n", []file.CodedSection{
				{ID: "section-2", CodeID: "code-2", Text: "impacts global warming"},
			}),
		},
		{
			Name: "RemovedCodeSections removes middle section with same code",
			Initial: buildFile("file-1", file.FileData{}, "Climate change impacts global warming and causes temperature rise.\n", []file.CodedSection{
				{ID: "section-1", CodeID: "code-1", Text: "Climate change impacts"},
				{ID: "section-2", CodeID: "code-1", Text: "impacts global warming"},
				{ID: "section-3", CodeID: "code-1", Text: "causes temperature rise"},
			}),
			Event: domain_helpers.NewDomainEvent(file.EntityName, "file-1", file.ModifiedCodeSections, &file.ModifiedCodeSectionsPayload{
				Operations: []file.SectionOp{
					{Op: "delete", ID: "section-2"},
				},
			}),
			Expected: buildFile("file-1", file.FileData{}, "Climate change impacts global warming and causes temperature rise.\n", []file.CodedSection{
				{ID: "section-1", CodeID: "code-1", Text: "Climate change impacts"},
				{ID: "section-3", CodeID: "code-1", Text: "causes temperature rise"},
			}),
		},
		{
			Name: "RemovedCodeSections removes last section with same code",
			Initial: buildFile("file-1", file.FileData{}, "Climate change impacts global warming and causes temperature rise.\n", []file.CodedSection{
				{ID: "section-1", CodeID: "code-1", Text: "Climate change impacts"},
				{ID: "section-2", CodeID: "code-1", Text: "impacts global warming"},
				{ID: "section-3", CodeID: "code-1", Text: "causes temperature rise"},
			}),
			Event: domain_helpers.NewDomainEvent(file.EntityName, "file-1", file.ModifiedCodeSections, &file.ModifiedCodeSectionsPayload{
				Operations: []file.SectionOp{
					{Op: "delete", ID: "section-3"},
				},
			}),
			Expected: buildFile("file-1", file.FileData{}, "Climate change impacts global warming and causes temperature rise.\n", []file.CodedSection{
				{ID: "section-1", CodeID: "code-1", Text: "Climate change impacts"},
				{ID: "section-2", CodeID: "code-1", Text: "impacts global warming"},
			}),
		},
		{
			Name: "RemovedCodeSections removes multiple specific sections",
			Initial: buildFile("file-1", file.FileData{}, "Climate change impacts global warming and causes temperature rise.\n", []file.CodedSection{
				{ID: "section-1", CodeID: "code-1", Text: "Climate change impacts"},
				{ID: "section-2", CodeID: "code-1", Text: "impacts global warming"},
				{ID: "section-3", CodeID: "code-2", Text: "global warming"},
				{ID: "section-4", CodeID: "code-1", Text: "causes temperature rise"},
			}),
			Event: domain_helpers.NewDomainEvent(file.EntityName, "file-1", file.ModifiedCodeSections, &file.ModifiedCodeSectionsPayload{
				Operations: []file.SectionOp{
					{Op: "delete", ID: "section-2"},
					{Op: "delete", ID: "section-4"},
				},
			}),
			Expected: buildFile("file-1", file.FileData{}, "Climate change impacts global warming and causes temperature rise.\n", []file.CodedSection{
				{ID: "section-1", CodeID: "code-1", Text: "Climate change impacts"},
				{ID: "section-3", CodeID: "code-2", Text: "global warming"},
			}),
		},
		{
			Name:    "ClearedCoding removes all codes from file",
			Initial: buildFile("file-1", file.FileData{}, "Test content\n", []file.CodedSection{
				{ID: "section-1", CodeID: "code-1", Text: "Test"},
			}),
			Event: domain_helpers.NewDomainEvent(file.EntityName, "file-1", file.ClearedCoding, nil),
			Expected: buildFile("file-1", file.FileData{}, "Test content\n", []file.CodedSection{}),
		},
		{
			Name: "DeletedCode removes all codes with matching CodeID",
			Initial: buildFile("file-1", file.FileData{}, "Climate change impacts global warming.\n", []file.CodedSection{
				{ID: "section-1", CodeID: "code-1", Text: "Climate change impacts"},
				{ID: "section-2", CodeID: "code-2", Text: "impacts global warming"},
			}),
			Event: domain_helpers.NewDomainEvent(code.EntityName, "code-1", code.DeletedCode, nil),
			Expected: buildFile("file-1", file.FileData{}, "Climate change impacts global warming.\n", []file.CodedSection{
				{ID: "section-2", CodeID: "code-2", Text: "impacts global warming"},
			}),
		},
		{
			Name: "MergedCodes replaces source CodeID with target CodeID",
			Initial: buildFile("file-1", file.FileData{}, "Climate change impacts global warming.\n", []file.CodedSection{
				{ID: "section-1", CodeID: "code-1", Text: "Climate change impacts"},
				{ID: "section-2", CodeID: "code-2", Text: "impacts global warming"},
			}),
			Event: domain_helpers.NewDomainEvent(code.EntityName, "code-1", code.MergedCodes, &code.MergedCodesPayload{SourceID: "code-1", TargetID: "code-2"}),
			Expected: buildFile("file-1", file.FileData{}, "Climate change impacts global warming.\n", []file.CodedSection{
				{ID: "section-1", CodeID: "code-2", Text: "Climate change impacts"},
				{ID: "section-2", CodeID: "code-2", Text: "impacts global warming"},
			}),
		},
		{
			Name: "RemovedCodeFromFile removes all sections with matching CodeID from file",
			Initial: buildFile("file-1", file.FileData{}, "Climate change impacts global warming.\n", []file.CodedSection{
				{ID: "section-1", CodeID: "code-1", Text: "Climate change impacts"},
				{ID: "section-2", CodeID: "code-2", Text: "impacts global warming"},
				{ID: "section-3", CodeID: "code-1", Text: "global warming"},
			}),
			Event: domain_helpers.NewDomainEvent(file.EntityName, "file-1", file.RemovedCodeFromFile, &file.RemovedCodeFromFilePayload{CodeID: "code-1"}),
			Expected: buildFile("file-1", file.FileData{}, "Climate change impacts global warming.\n", []file.CodedSection{
				{ID: "section-2", CodeID: "code-2", Text: "impacts global warming"},
			}),
		},
		{
			Name: "ClearedCodeApplications removes all sections with matching CodeID across all files",
			Initial: buildFile("file-1", file.FileData{}, "Climate change impacts global warming.\n", []file.CodedSection{
				{ID: "section-1", CodeID: "code-1", Text: "Climate change impacts"},
				{ID: "section-2", CodeID: "code-2", Text: "impacts global warming"},
			}),
			Event: domain_helpers.NewDomainEvent(code.EntityName, "code-1", code.ClearedCodeApplications, nil),
			Expected: buildFile("file-1", file.FileData{}, "Climate change impacts global warming.\n", []file.CodedSection{
				{ID: "section-2", CodeID: "code-2", Text: "impacts global warming"},
			}),
		},
		{
			Name: "RecodedAll replaces source CodeID with target CodeID keeping both codes",
			Initial: buildFile("file-1", file.FileData{}, "Climate change impacts global warming.\n", []file.CodedSection{
				{ID: "section-1", CodeID: "code-1", Text: "Climate change impacts"},
				{ID: "section-2", CodeID: "code-2", Text: "impacts global warming"},
			}),
			Event: domain_helpers.NewDomainEvent(code.EntityName, "code-1", code.RecodedAll, &code.RecodedAllPayload{TargetCodeID: "code-2"}),
			Expected: buildFile("file-1", file.FileData{}, "Climate change impacts global warming.\n", []file.CodedSection{
				{ID: "section-1", CodeID: "code-2", Text: "Climate change impacts"},
				{ID: "section-2", CodeID: "code-2", Text: "impacts global warming"},
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

	normalize := func(f *File) *File {
		if f != nil {
			f.Time = testTime
			counter := 0
			for j := range f.Codes {
				id := f.Codes[j].ID
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
					f.Codes[j].ID = fmt.Sprintf("generated-id-%d", counter)
				}
			}
		}
		return f
	}

	reducer_helpers.RunReducerTests(t, combinedTests, Reducer, normalize)
}
