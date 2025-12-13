package handlers

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/domain/entities/code"
	"hermes-relay/internal/domain/entities/file"
	"hermes-relay/internal/domain/entities/project"
	th "hermes-relay/internal/lib/test-helpers"
	"hermes-relay/internal/lib/test-helpers/domain-helpers"
	rh "hermes-relay/internal/lib/test-helpers/router-helpers"
	"hermes-relay/internal/lib/utils"
	"testing"
)

var (
	testProjectID    = utils.NewID()
	testCodeID1      = utils.NewID()
	testCodeID2      = utils.NewID()
	testCodeID3      = utils.NewID()
	testSectionID1   = utils.NewID()
	testSectionID2   = utils.NewID()
	testSectionID3   = utils.NewID()
	testFileShort    = utils.NewID()
	testFileLong     = utils.NewID()
	testFileMemo     = utils.NewID()
	testFileWithCode = utils.NewID()
	shortContent     = "Some text to code here"
	longContent      = "Some text to code here and more words for searching"
	memoContent      = "This is a memo with some editable content here"

	confidenceHigh = file.ConfidenceHigh
	confidenceLow  = file.ConfidenceLow
)

var cmds = []*commands.AnyMessage{
	project.CreatedProjectEvent(testProjectID),
	code.CreatedCodeEvent(testCodeID1, testProjectID, "topic:first"),
	code.CreatedCodeEvent(testCodeID2, testProjectID, "topic:second"),
	code.CreatedCodeEvent(testCodeID3, testProjectID, "topic:third"),
	file.CreatedFileEventWithType(testFileShort, testProjectID, shortContent, file.FileTypeCorpus),
	file.CreatedFileEventWithType(testFileLong, testProjectID, longContent, file.FileTypeCorpus),
	file.CreatedFileEventWithType(testFileMemo, testProjectID, memoContent, file.FileTypeMemo),
	file.CreatedFileWithSectionsEvent(testFileWithCode, testProjectID, longContent, []file.CodedSection{
		{ID: testSectionID1, CodeID: testCodeID1, Text: "Some text to code", Confidence: file.ConfidenceHigh},
		{ID: testSectionID2, CodeID: testCodeID2, Text: "more words for", Confidence: file.ConfidenceMedium},
		{ID: testSectionID3, CodeID: testCodeID3, Text: "words for searching", Confidence: file.ConfidenceLow},
	}),
}

var ignoreGeneratedIDs = []th.IgnoreFieldsOption{
	{Type: file.SectionOp{}, Fields: []string{"ID"}, EnsureValidUUID: true},
}

func TestFileRouter(t *testing.T) {
	tests := []rh.RouterTestCase{
		{
			Name: "CreateFile with valid payload",
			Input: commands.ToAny(commands.NewCommand[file.CreateFilePayload, any](file.CreateFile, file.CreateFilePayload{
				ProjectID: testProjectID,
				Name:      "test-file.txt",
				Content:   "Test content",
			}, file.EntityName, "", domain_helpers.TestActor(), nil)),
			ExpectErr: "",
			ExpectEvent: commands.ToAny(commands.NewDomainEvent[file.CreatedFilePayload, any](file.CreatedFile, file.CreatedFilePayload{
				FileData: file.FileData{
					ProjectID: testProjectID,
					Name:      "test-file.txt",
					Type:      file.FileTypeCorpus,
					Locked:    true,
				},
				Content: "Test content\n",
				Codes:   []file.CodedSection{},
			}, file.EntityName, "", domain_helpers.TestActor(), nil)),
		},
		{
			Name: "CreateFile with minimal required fields",
			Input: commands.ToAny(commands.NewCommand[file.CreateFilePayload, any](file.CreateFile, file.CreateFilePayload{
				ProjectID: testProjectID,
				Name:      "minimal.txt",
				Content:   "Content",
			}, file.EntityName, "", domain_helpers.TestActor(), nil)),
			ExpectErr: "",
			ExpectEvent: commands.ToAny(commands.NewDomainEvent[file.CreatedFilePayload, any](file.CreatedFile, file.CreatedFilePayload{
				FileData: file.FileData{
					ProjectID: testProjectID,
					Name:      "minimal.txt",
					Type:      file.FileTypeCorpus,
					Locked:    true,
				},
				Content: "Content\n",
				Codes:   []file.CodedSection{},
			}, file.EntityName, "", domain_helpers.TestActor(), nil)),
		},
		{
			Name: "AddCodeSections with valid payload",
			Input: commands.ToAny(commands.NewCommand[file.AddCodeSectionsPayload, any](file.AddCodeSections, file.AddCodeSectionsPayload{
				Sections: []file.AddSectionOp{
					{CodeID: testCodeID1, Text: "Some text to code", Reason: "Test reason", Confidence: file.ConfidenceHigh},
				},
			}, file.EntityName, testFileShort, domain_helpers.TestActor(), nil)),
			ExpectErr: "",
			ExpectEvent: commands.ToAny(commands.NewDomainEvent[file.ModifiedCodeSectionsPayload, any](file.ModifiedCodeSections, file.ModifiedCodeSectionsPayload{
				Operations: []file.SectionOp{
					{Op: "add", ID: "generated-id", CodeID: testCodeID1, Text: "Some text to code", Reason: "Test reason", Confidence: &confidenceHigh},
				},
			}, file.EntityName, testFileShort, domain_helpers.TestActor(), nil)),
			IgnoreFields: ignoreGeneratedIDs,
		},
		{
			Name: "UpdateCodeSections with valid payload",
			Input: commands.ToAny(commands.NewCommand[file.UpdateCodeSectionsPayload, any](file.UpdateCodeSections, file.UpdateCodeSectionsPayload{
				Sections: []file.UpdateSectionOp{
					{ID: testSectionID1, Text: "Some text to code", Reason: "Updated reason"},
				},
			}, file.EntityName, testFileWithCode, domain_helpers.TestActor(), nil)),
			ExpectErr: "",
			ExpectEvent: commands.ToAny(commands.NewDomainEvent[file.ModifiedCodeSectionsPayload, any](file.ModifiedCodeSections, file.ModifiedCodeSectionsPayload{
				Operations: []file.SectionOp{
					{Op: "update", ID: testSectionID1, Text: "Some text to code", Reason: "Updated reason"},
				},
			}, file.EntityName, testFileWithCode, domain_helpers.TestActor(), nil)),
		},
		{
			Name: "RemoveCodeSections with valid payload",
			Input: commands.ToAny(commands.NewCommand[file.RemoveCodeSectionsPayload, any](file.RemoveCodeSections, file.RemoveCodeSectionsPayload{
				SectionIDs: []string{testSectionID1, testSectionID2},
			}, file.EntityName, testFileWithCode, domain_helpers.TestActor(), nil)),
			ExpectErr: "",
			ExpectEvent: commands.ToAny(commands.NewDomainEvent[file.ModifiedCodeSectionsPayload, any](file.ModifiedCodeSections, file.ModifiedCodeSectionsPayload{
				Operations: []file.SectionOp{
					{Op: "delete", ID: testSectionID1},
					{Op: "delete", ID: testSectionID2},
				},
			}, file.EntityName, testFileWithCode, domain_helpers.TestActor(), nil)),
		},
		{
			Name:        "ClearCoding",
			Input:       commands.ToAny(commands.NewCommand[any, any](file.ClearCoding, nil, file.EntityName, testFileShort, domain_helpers.TestActor(), nil)),
			ExpectErr:   "",
			ExpectEvent: commands.ToAny(commands.NewDomainEvent[any, any](file.ClearedCoding, nil, file.EntityName, testFileShort, domain_helpers.TestActor(), nil)),
		},
		{
			Name: "UpdateFile with valid payload",
			Input: commands.ToAny(commands.NewCommand[file.UpdateFilePayload, any](file.UpdateFile, file.UpdateFilePayload{
				Name:        "updated-file.txt",
				Description: "Updated description",
			}, file.EntityName, testFileShort, domain_helpers.TestActor(), nil)),
			ExpectErr: "",
			ExpectEvent: commands.ToAny(commands.NewDomainEvent[file.UpdatedFilePayload, any](file.UpdatedFile, file.UpdatedFilePayload{
				Name:        "updated-file.txt",
				Description: "Updated description",
			}, file.EntityName, testFileShort, domain_helpers.TestActor(), nil)),
		},
		{
			Name: "UpdateFile with empty description",
			Input: commands.ToAny(commands.NewCommand[file.UpdateFilePayload, any](file.UpdateFile, file.UpdateFilePayload{
				Name:        "minimal.txt",
				Description: "",
			}, file.EntityName, testFileShort, domain_helpers.TestActor(), nil)),
			ExpectErr: "",
			ExpectEvent: commands.ToAny(commands.NewDomainEvent[file.UpdatedFilePayload, any](file.UpdatedFile, file.UpdatedFilePayload{
				Name:        "minimal.txt",
				Description: "",
			}, file.EntityName, testFileShort, domain_helpers.TestActor(), nil)),
		},
		{
			Name: "UpdateFile with missing Name",
			Input: commands.ToAny(commands.NewCommand[file.UpdateFilePayload, any](file.UpdateFile, file.UpdateFilePayload{
				Description: "Some description",
			}, file.EntityName, testFileShort, domain_helpers.TestActor(), nil)),
			ExpectErr: "validation failed: Name is required",
		},
		{
			Name:        "DeleteFile with valid aggregate ID",
			Input:       commands.ToAny(commands.NewCommand[file.DeleteFilePayload, any](file.DeleteFile, file.DeleteFilePayload{}, file.EntityName, testFileShort, domain_helpers.TestActor(), nil)),
			ExpectErr:   "",
			ExpectEvent: commands.ToAny(commands.NewDomainEvent[any, any](file.DeletedFile, nil, file.EntityName, testFileShort, domain_helpers.TestActor(), nil)),
		},
		{
			Name: "ReplaceFileContent on unlocked file",
			Input: commands.ToAny(commands.NewCommand[file.ReplaceFileContentPayload, any](file.ReplaceFileContent, file.ReplaceFileContentPayload{
				Content: "New memo content",
			}, file.EntityName, testFileMemo, domain_helpers.TestActor(), nil)),
			ExpectErr: "",
			ExpectEvent: commands.ToAny(commands.NewDomainEvent[file.ReplacedFileContentPayload, any](file.ReplacedFileContent, file.ReplacedFileContentPayload{
				Content: "New memo content",
			}, file.EntityName, testFileMemo, domain_helpers.TestActor(), nil)),
		},
		{
			Name: "ReplaceFileContent on locked file fails",
			Input: commands.ToAny(commands.NewCommand[file.ReplaceFileContentPayload, any](file.ReplaceFileContent, file.ReplaceFileContentPayload{
				Content: "New content",
			}, file.EntityName, testFileShort, domain_helpers.TestActor(), nil)),
			ExpectErr: "file is locked",
		},
		{
			Name: "EditFileContent with fuzzy match",
			Input: commands.ToAny(commands.NewCommand[file.EditFileContentPayload, any](file.EditFileContent, file.EditFileContentPayload{
				OldText: "some editable content",
				NewText: "some UPDATED content",
			}, file.EntityName, testFileMemo, domain_helpers.TestActor(), nil)),
			ExpectErr: "",
			ExpectEvent: commands.ToAny(commands.NewDomainEvent[file.ReplacedFileContentPayload, any](file.ReplacedFileContent, file.ReplacedFileContentPayload{
				Content: "This is a memo with some UPDATED content here",
			}, file.EntityName, testFileMemo, domain_helpers.TestActor(), nil)),
		},
		{
			Name: "EditFileContent on locked file fails",
			Input: commands.ToAny(commands.NewCommand[file.EditFileContentPayload, any](file.EditFileContent, file.EditFileContentPayload{
				OldText: "Some text to code",
				NewText: "replacement",
			}, file.EntityName, testFileShort, domain_helpers.TestActor(), nil)),
			ExpectErr: "file is locked",
		},
		{
			Name: "EditFileContent with text not found",
			Input: commands.ToAny(commands.NewCommand[file.EditFileContentPayload, any](file.EditFileContent, file.EditFileContentPayload{
				OldText: "nonexistent text that does not exist",
				NewText: "replacement",
			}, file.EntityName, testFileMemo, domain_helpers.TestActor(), nil)),
			ExpectErr: "old_text",
		},
		{
			Name: "CreateFile with missing Name",
			Input: commands.ToAny(commands.NewCommand[file.CreateFilePayload, any](file.CreateFile, file.CreateFilePayload{
				ProjectID: testProjectID,
				Content:   "Content",
			}, file.EntityName, "", domain_helpers.TestActor(), nil)),
			ExpectErr: "validation failed: Name is required",
		},
		{
			Name: "CreateFile with missing ProjectID",
			Input: commands.ToAny(commands.NewCommand[file.CreateFilePayload, any](file.CreateFile, file.CreateFilePayload{
				Name:    "file.txt",
				Content: "Content",
			}, file.EntityName, "", domain_helpers.TestActor(), nil)),
			ExpectErr: "validation failed: ProjectID is required",
		},
		{
			Name: "AddCodeSections with empty sections array",
			Input: commands.ToAny(commands.NewCommand[file.AddCodeSectionsPayload, any](file.AddCodeSections, file.AddCodeSectionsPayload{
				Sections: []file.AddSectionOp{},
			}, file.EntityName, testFileShort, domain_helpers.TestActor(), nil)),
			ExpectErr: "validation failed: Sections must be at least 1 characters",
		},
		{
			Name: "AddCodeSections with missing CodeID",
			Input: commands.ToAny(commands.NewCommand[file.AddCodeSectionsPayload, any](file.AddCodeSections, file.AddCodeSectionsPayload{
				Sections: []file.AddSectionOp{
					{Text: "Some text", Confidence: file.ConfidenceHigh},
				},
			}, file.EntityName, testFileShort, domain_helpers.TestActor(), nil)),
			ExpectErr: "validation failed: CodeID is required",
		},
		{
			Name: "AddCodeSections with invalid CodeID",
			Input: commands.ToAny(commands.NewCommand[file.AddCodeSectionsPayload, any](file.AddCodeSections, file.AddCodeSectionsPayload{
				Sections: []file.AddSectionOp{
					{CodeID: "not-a-uuid", Text: "Some text", Confidence: file.ConfidenceHigh},
				},
			}, file.EntityName, testFileShort, domain_helpers.TestActor(), nil)),
			ExpectErr: "validation failed: CodeID failed validation (valid_id_or_slug)",
		},
		{
			Name: "AddCodeSections fails with nonexistent code id",
			Input: commands.ToAny(commands.NewCommand[file.AddCodeSectionsPayload, any](file.AddCodeSections, file.AddCodeSectionsPayload{
				Sections: []file.AddSectionOp{
					{CodeID: utils.NewID(), Text: "Some text to code", Confidence: file.ConfidenceHigh},
				},
			}, file.EntityName, testFileShort, domain_helpers.TestActor(), nil)),
			ExpectErr: "validation failed: operations[0] code not found",
		},
		{
			Name: "UpdateCodeSections with missing section ID",
			Input: commands.ToAny(commands.NewCommand[file.UpdateCodeSectionsPayload, any](file.UpdateCodeSections, file.UpdateCodeSectionsPayload{
				Sections: []file.UpdateSectionOp{
					{Text: "Updated text"},
				},
			}, file.EntityName, testFileWithCode, domain_helpers.TestActor(), nil)),
			ExpectErr: "validation failed: ID is required",
		},
		{
			Name: "RemoveCodeSections with empty section IDs",
			Input: commands.ToAny(commands.NewCommand[file.RemoveCodeSectionsPayload, any](file.RemoveCodeSections, file.RemoveCodeSectionsPayload{
				SectionIDs: []string{},
			}, file.EntityName, testFileWithCode, domain_helpers.TestActor(), nil)),
			ExpectErr: "validation failed: SectionIDs must be at least 1 characters",
		},
		{
			Name: "Wrong entity type returns nil",
			Input: commands.ToAny(commands.NewCommand[file.CreateFilePayload, any](file.CreateFile, file.CreateFilePayload{
				ProjectID: testProjectID,
				Name:      "test.txt",
				Content:   "Test",
			}, "DifferentEntity", "", domain_helpers.TestActor(), nil)),
			ExpectErr:   "",
			ExpectEvent: nil,
		},
		{
			Name: "Wrong action returns nil",
			Input: commands.ToAny(commands.NewCommand[file.CreateFilePayload, any]("DifferentAction", file.CreateFilePayload{
				ProjectID: testProjectID,
				Name:      "test.txt",
				Content:   "Test",
			}, file.EntityName, "test-aggregate-id", domain_helpers.TestActor(), nil)),
			ExpectErr:   "",
			ExpectEvent: nil,
		},
		{
			Name: "AddCodeSections partial success - 2 valid, 1 invalid",
			Input: commands.ToAny(commands.NewCommand[file.AddCodeSectionsPayload, any](file.AddCodeSections, file.AddCodeSectionsPayload{
				Sections: []file.AddSectionOp{
					{CodeID: testCodeID1, Text: "Some text to code", Reason: "First reason", Confidence: file.ConfidenceHigh},
					{CodeID: testCodeID2, Text: "xy", Reason: "Too short", Confidence: file.ConfidenceMedium},
					{CodeID: testCodeID3, Text: "more words for searching", Reason: "Third reason", Confidence: file.ConfidenceLow},
				},
			}, file.EntityName, testFileLong, domain_helpers.TestActor(), nil)),
			ExpectErr: "",
			ExpectEvent: commands.ToAny(commands.NewDomainEvent[file.ModifiedCodeSectionsPayload, any](file.ModifiedCodeSections, file.ModifiedCodeSectionsPayload{
				Operations: []file.SectionOp{
					{Op: "add", ID: "generated-id", CodeID: testCodeID1, Text: "Some text to code", Reason: "First reason", Confidence: &confidenceHigh},
					{Op: "add", ID: "generated-id", CodeID: testCodeID3, Text: "more words for searching", Reason: "Third reason", Confidence: &confidenceLow},
				},
				Failures: map[int]string{1: "text too short (1 words, need 3+) - expand selection: \"xy\""},
			}, file.EntityName, testFileLong, domain_helpers.TestActor(), nil)),
			IgnoreFields: ignoreGeneratedIDs,
		},
		{
			Name: "AddCodeSections all fail - returns validation error with details",
			Input: commands.ToAny(commands.NewCommand[file.AddCodeSectionsPayload, any](file.AddCodeSections, file.AddCodeSectionsPayload{
				Sections: []file.AddSectionOp{
					{CodeID: testCodeID1, Text: "xy", Reason: "Too short", Confidence: file.ConfidenceHigh},
				},
			}, file.EntityName, testFileShort, domain_helpers.TestActor(), nil)),
			ExpectErr: "validation failed: operations[0] text too short (1 words, need 3+) - expand selection: \"xy\"",
		},
		{
			Name: "UpdateCodeSections partial success - 2 valid, 1 invalid",
			Input: commands.ToAny(commands.NewCommand[file.UpdateCodeSectionsPayload, any](file.UpdateCodeSections, file.UpdateCodeSectionsPayload{
				Sections: []file.UpdateSectionOp{
					{ID: testSectionID1, Text: "Some text to code", Reason: "First update"},
					{ID: testSectionID2, Text: "xy", Reason: "Too short"},
					{ID: testSectionID3, Text: "more words for searching", Reason: "Third update"},
				},
			}, file.EntityName, testFileWithCode, domain_helpers.TestActor(), nil)),
			ExpectErr: "",
			ExpectEvent: commands.ToAny(commands.NewDomainEvent[file.ModifiedCodeSectionsPayload, any](file.ModifiedCodeSections, file.ModifiedCodeSectionsPayload{
				Operations: []file.SectionOp{
					{Op: "update", ID: testSectionID1, Text: "Some text to code", Reason: "First update"},
					{Op: "update", ID: testSectionID3, Text: "more words for searching", Reason: "Third update"},
				},
				Failures: map[int]string{1: "text too short (1 words, need 3+) - expand selection: \"xy\""},
			}, file.EntityName, testFileWithCode, domain_helpers.TestActor(), nil)),
		},
		{
			Name: "UpdateCodeSections all fail - returns validation error",
			Input: commands.ToAny(commands.NewCommand[file.UpdateCodeSectionsPayload, any](file.UpdateCodeSections, file.UpdateCodeSectionsPayload{
				Sections: []file.UpdateSectionOp{
					{ID: testSectionID1, Text: "xy", Reason: "Too short"},
					{ID: testSectionID2, Text: "ab", Reason: "Also too short"},
				},
			}, file.EntityName, testFileWithCode, domain_helpers.TestActor(), nil)),
			ExpectErr: "validation failed",
		},
		{
			Name: "UpdateCodeSections reassign code by id",
			Input: commands.ToAny(commands.NewCommand[file.UpdateCodeSectionsPayload, any](file.UpdateCodeSections, file.UpdateCodeSectionsPayload{
				Sections: []file.UpdateSectionOp{
					{ID: testSectionID1, CodeID: testCodeID3},
				},
			}, file.EntityName, testFileWithCode, domain_helpers.TestActor(), nil)),
			ExpectErr: "",
			ExpectEvent: commands.ToAny(commands.NewDomainEvent[file.ModifiedCodeSectionsPayload, any](file.ModifiedCodeSections, file.ModifiedCodeSectionsPayload{
				Operations: []file.SectionOp{
					{Op: "update", ID: testSectionID1, CodeID: testCodeID3},
				},
			}, file.EntityName, testFileWithCode, domain_helpers.TestActor(), nil)),
		},
		{
			Name: "UpdateCodeSections fails with nonexistent code id",
			Input: commands.ToAny(commands.NewCommand[file.UpdateCodeSectionsPayload, any](file.UpdateCodeSections, file.UpdateCodeSectionsPayload{
				Sections: []file.UpdateSectionOp{
					{ID: testSectionID1, CodeID: utils.NewID()},
				},
			}, file.EntityName, testFileWithCode, domain_helpers.TestActor(), nil)),
			ExpectErr: "validation failed: operations[0] code not found",
		},
		{
			Name: "UpdateCodeSections fails with nonexistent section id",
			Input: commands.ToAny(commands.NewCommand[file.UpdateCodeSectionsPayload, any](file.UpdateCodeSections, file.UpdateCodeSectionsPayload{
				Sections: []file.UpdateSectionOp{
					{ID: utils.NewID(), Reason: "Updated"},
				},
			}, file.EntityName, testFileWithCode, domain_helpers.TestActor(), nil)),
			ExpectErr: "validation failed: operations[0] section not found",
		},
		{
			Name: "AddCodeSections fails on non-corpus file",
			Input: commands.ToAny(commands.NewCommand[file.AddCodeSectionsPayload, any](file.AddCodeSections, file.AddCodeSectionsPayload{
				Sections: []file.AddSectionOp{
					{CodeID: testCodeID1, Text: "some editable content", Confidence: file.ConfidenceHigh},
				},
			}, file.EntityName, testFileMemo, domain_helpers.TestActor(), nil)),
			ExpectErr: "coding is only allowed on corpus files",
		},
		{
			Name: "UpdateCodeSections fails on non-corpus file",
			Input: commands.ToAny(commands.NewCommand[file.UpdateCodeSectionsPayload, any](file.UpdateCodeSections, file.UpdateCodeSectionsPayload{
				Sections: []file.UpdateSectionOp{
					{ID: testSectionID1, Reason: "Updated"},
				},
			}, file.EntityName, testFileMemo, domain_helpers.TestActor(), nil)),
			ExpectErr: "coding is only allowed on corpus files",
		},
		{
			Name: "RemoveCodeSections fails on non-corpus file",
			Input: commands.ToAny(commands.NewCommand[file.RemoveCodeSectionsPayload, any](file.RemoveCodeSections, file.RemoveCodeSectionsPayload{
				SectionIDs: []string{testSectionID1},
			}, file.EntityName, testFileMemo, domain_helpers.TestActor(), nil)),
			ExpectErr: "coding is only allowed on corpus files",
		},
		{
			Name:      "ClearCoding fails on non-corpus file",
			Input:     commands.ToAny(commands.NewCommand[any, any](file.ClearCoding, nil, file.EntityName, testFileMemo, domain_helpers.TestActor(), nil)),
			ExpectErr: "coding is only allowed on corpus files",
		},
		{
			Name: "RemoveCodeFromFile with valid payload",
			Input: commands.ToAny(commands.NewCommand[file.RemoveCodeFromFilePayload, any](file.RemoveCodeFromFile, file.RemoveCodeFromFilePayload{
				CodeID: testCodeID1,
			}, file.EntityName, testFileWithCode, domain_helpers.TestActor(), nil)),
			ExpectErr: "",
			ExpectEvent: commands.ToAny(commands.NewDomainEvent[file.RemovedCodeFromFilePayload, any](file.RemovedCodeFromFile, file.RemovedCodeFromFilePayload{
				CodeID: testCodeID1,
			}, file.EntityName, testFileWithCode, domain_helpers.TestActor(), nil)),
		},
		{
			Name: "RemoveCodeFromFile fails on non-corpus file",
			Input: commands.ToAny(commands.NewCommand[file.RemoveCodeFromFilePayload, any](file.RemoveCodeFromFile, file.RemoveCodeFromFilePayload{
				CodeID: testCodeID1,
			}, file.EntityName, testFileMemo, domain_helpers.TestActor(), nil)),
			ExpectErr: "coding is only allowed on corpus files",
		},
		{
			Name: "RemoveCodeFromFile with missing CodeID",
			Input: commands.ToAny(commands.NewCommand[file.RemoveCodeFromFilePayload, any](file.RemoveCodeFromFile, file.RemoveCodeFromFilePayload{
				CodeID: "",
			}, file.EntityName, testFileWithCode, domain_helpers.TestActor(), nil)),
			ExpectErr: "validation failed: CodeID is required",
		},
		{
			Name: "AddCodeSections with LLM actor requires reason",
			Input: commands.ToAny(commands.NewCommand[file.AddCodeSectionsPayload, any](file.AddCodeSections, file.AddCodeSectionsPayload{
				Sections: []file.AddSectionOp{
					{CodeID: testCodeID1, Text: "Some text to code", Confidence: file.ConfidenceHigh},
				},
			}, file.EntityName, testFileShort, commands.Actor{UserID: "claude", ActorType: commands.ActorTypeLLM}, nil)),
			ExpectErr: "validation failed: sections[0] reason is required for LLM actor",
		},
		{
			Name: "AddCodeSections with LLM actor and reason succeeds",
			Input: commands.ToAny(commands.NewCommand[file.AddCodeSectionsPayload, any](file.AddCodeSections, file.AddCodeSectionsPayload{
				Sections: []file.AddSectionOp{
					{CodeID: testCodeID1, Text: "Some text to code", Reason: "LLM reasoning", Confidence: file.ConfidenceHigh},
				},
			}, file.EntityName, testFileShort, commands.Actor{UserID: "claude", ActorType: commands.ActorTypeLLM}, nil)),
			ExpectErr: "",
			ExpectEvent: commands.ToAny(commands.NewDomainEvent[file.ModifiedCodeSectionsPayload, any](file.ModifiedCodeSections, file.ModifiedCodeSectionsPayload{
				Operations: []file.SectionOp{
					{Op: "add", ID: "generated-id", CodeID: testCodeID1, Text: "Some text to code", Reason: "LLM reasoning", Confidence: &confidenceHigh},
				},
			}, file.EntityName, testFileShort, commands.Actor{UserID: "claude", ActorType: commands.ActorTypeLLM}, nil)),
			IgnoreFields: ignoreGeneratedIDs,
		},
		{
			Name: "AddCodeSections with human actor without reason succeeds",
			Input: commands.ToAny(commands.NewCommand[file.AddCodeSectionsPayload, any](file.AddCodeSections, file.AddCodeSectionsPayload{
				Sections: []file.AddSectionOp{
					{CodeID: testCodeID1, Text: "Some text to code", Confidence: file.ConfidenceHigh},
				},
			}, file.EntityName, testFileShort, commands.Actor{UserID: "user-123", ActorType: commands.ActorTypeHuman}, nil)),
			ExpectErr: "",
			ExpectEvent: commands.ToAny(commands.NewDomainEvent[file.ModifiedCodeSectionsPayload, any](file.ModifiedCodeSections, file.ModifiedCodeSectionsPayload{
				Operations: []file.SectionOp{
					{Op: "add", ID: "generated-id", CodeID: testCodeID1, Text: "Some text to code", Reason: "", Confidence: &confidenceHigh},
				},
			}, file.EntityName, testFileShort, commands.Actor{UserID: "user-123", ActorType: commands.ActorTypeHuman}, nil)),
			IgnoreFields: ignoreGeneratedIDs,
		},
		{
			Name: "AddCodeSections with slug resolves to code ID",
			Input: commands.ToAny(commands.NewCommand[file.AddCodeSectionsPayload, any](file.AddCodeSections, file.AddCodeSectionsPayload{
				Sections: []file.AddSectionOp{
					{CodeID: "topic:first", Text: "Some text to code", Reason: "Using slug", Confidence: file.ConfidenceHigh},
				},
			}, file.EntityName, testFileShort, domain_helpers.TestActor(), nil)),
			ExpectErr: "",
			ExpectEvent: commands.ToAny(commands.NewDomainEvent[file.ModifiedCodeSectionsPayload, any](file.ModifiedCodeSections, file.ModifiedCodeSectionsPayload{
				Operations: []file.SectionOp{
					{Op: "add", ID: "generated-id", CodeID: testCodeID1, Text: "Some text to code", Reason: "Using slug", Confidence: &confidenceHigh},
				},
			}, file.EntityName, testFileShort, domain_helpers.TestActor(), nil)),
			IgnoreFields: ignoreGeneratedIDs,
		},
		{
			Name: "AddCodeSections with invalid slug fails",
			Input: commands.ToAny(commands.NewCommand[file.AddCodeSectionsPayload, any](file.AddCodeSections, file.AddCodeSectionsPayload{
				Sections: []file.AddSectionOp{
					{CodeID: "topic:nonexistent", Text: "Some text to code", Confidence: file.ConfidenceHigh},
				},
			}, file.EntityName, testFileShort, domain_helpers.TestActor(), nil)),
			ExpectErr: "validation failed: operations[0] code not found: topic:nonexistent",
		},
		{
			Name: "UpdateCodeSections with slug resolves to code ID",
			Input: commands.ToAny(commands.NewCommand[file.UpdateCodeSectionsPayload, any](file.UpdateCodeSections, file.UpdateCodeSectionsPayload{
				Sections: []file.UpdateSectionOp{
					{ID: testSectionID1, CodeID: "topic:second", Reason: "Reassigning with slug"},
				},
			}, file.EntityName, testFileWithCode, domain_helpers.TestActor(), nil)),
			ExpectErr: "",
			ExpectEvent: commands.ToAny(commands.NewDomainEvent[file.ModifiedCodeSectionsPayload, any](file.ModifiedCodeSections, file.ModifiedCodeSectionsPayload{
				Operations: []file.SectionOp{
					{Op: "update", ID: testSectionID1, CodeID: testCodeID2, Reason: "Reassigning with slug"},
				},
			}, file.EntityName, testFileWithCode, domain_helpers.TestActor(), nil)),
		},
	}

	rh.RunRouterTests(t, cmds, tests, NewRouter)
}
