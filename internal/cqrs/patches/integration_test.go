package patches

import (
	"hermes-relay/internal/domain/entities/file"
	"hermes-relay/internal/domain/entities/project"
	th "hermes-relay/internal/lib/test-helpers"
	"testing"
)

func buildProject(files map[string]file.File) *project.Project {
	proj := project.BuildTestProject("proj-1", project.ProjectData{})
	proj.Files = files
	return &proj
}

func buildFile(id, name, content string, codes []file.CodedSection) file.File {
	f := file.BuildTestFile(id, file.FileData{ProjectID: "proj-1", Name: name})
	f.Content = content
	f.Codes = codes
	return f
}

type patchTestCase struct {
	Name            string
	Before          *project.Project
	After           *project.Project
	IsActive        bool
	ExpectedType    string
	ExpectNullPatch bool
}

func TestDecidePatchWithStateChanges(t *testing.T) {
	tests := []patchTestCase{
		{
			Name: "Removing coded sections generates patch",
			Before: buildProject(map[string]file.File{
				"file-1": buildFile("file-1", "test.txt", "content", []file.CodedSection{
					file.BuildTestCodedSection("s1", "c1", "text1"),
					file.BuildTestCodedSection("s2", "c2", "text2"),
					file.BuildTestCodedSection("s3", "c3", "text3"),
				}),
			}),
			After: buildProject(map[string]file.File{
				"file-1": buildFile("file-1", "test.txt", "content", []file.CodedSection{
					file.BuildTestCodedSection("s3", "c3", "text3"),
				}),
			}),
			IsActive:        true,
			ExpectedType:    ActionTypePatch,
			ExpectNullPatch: false,
		},
		{
			Name: "Adding coded sections generates patch",
			Before: buildProject(map[string]file.File{
				"file-1": buildFile("file-1", "test.txt", "content", []file.CodedSection{}),
			}),
			After: buildProject(map[string]file.File{
				"file-1": buildFile("file-1", "test.txt", "content", []file.CodedSection{
					file.BuildTestCodedSection("s1", "c1", "new"),
				}),
			}),
			IsActive:        true,
			ExpectedType:    ActionTypePatch,
			ExpectNullPatch: false,
		},
		{
			Name: "Changing file name generates patch",
			Before: buildProject(map[string]file.File{
				"file-1": buildFile("file-1", "old.txt", "", []file.CodedSection{}),
			}),
			After: buildProject(map[string]file.File{
				"file-1": buildFile("file-1", "new.txt", "", []file.CodedSection{}),
			}),
			IsActive:        true,
			ExpectedType:    ActionTypePatch,
			ExpectNullPatch: false,
		},
		{
			Name: "Changing file content generates patch",
			Before: buildProject(map[string]file.File{
				"file-1": buildFile("file-1", "test.txt", "old content", []file.CodedSection{}),
			}),
			After: buildProject(map[string]file.File{
				"file-1": buildFile("file-1", "test.txt", "new content", []file.CodedSection{}),
			}),
			IsActive:        true,
			ExpectedType:    ActionTypePatch,
			ExpectNullPatch: false,
		},
		{
			Name:            "Identical projects return none",
			Before:          buildProject(map[string]file.File{}),
			After:           buildProject(map[string]file.File{}),
			IsActive:        true,
			ExpectedType:    ActionTypeNone,
			ExpectNullPatch: true,
		},
		{
			Name:            "Nil before returns snapshot",
			Before:          nil,
			After:           buildProject(map[string]file.File{}),
			IsActive:        true,
			ExpectedType:    ActionTypeSnapshot,
			ExpectNullPatch: true,
		},
		{
			Name:            "Nil after returns none",
			Before:          buildProject(map[string]file.File{}),
			After:           nil,
			IsActive:        true,
			ExpectedType:    ActionTypeNone,
			ExpectNullPatch: true,
		},
		{
			Name:            "Inactive project returns none",
			Before:          buildProject(map[string]file.File{}),
			After:           buildProject(map[string]file.File{"file-1": buildFile("file-1", "new.txt", "", []file.CodedSection{})}),
			IsActive:        false,
			ExpectedType:    ActionTypeNone,
			ExpectNullPatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			assertPatchAction(t, tt)
		})
	}
}

func assertPatchAction(t *testing.T, tc patchTestCase) {
	action, err := DecidePatch(tc.Before, tc.After, tc.IsActive)

	th.AssertError(t, err, "", "decide patch")
	th.AssertEqual(t, action.Type, tc.ExpectedType, "action type")
	assertPatchData(t, tc, action)
}

func assertPatchData(t *testing.T, tc patchTestCase, action PatchAction) {
	if tc.ExpectNullPatch {
		assertNullPatch(t, action)
		return
	}

	if tc.ExpectedType == ActionTypePatch {
		assertNonNullPatch(t, action)
	}
}

func assertNullPatch(t *testing.T, action PatchAction) {
	if action.Patch != nil {
		t.Errorf("expected nil patch, got %s", string(action.Patch))
	}
}

func assertNonNullPatch(t *testing.T, action PatchAction) {
	if len(action.Patch) == 0 {
		t.Error("expected non-empty patch")
	}
	if string(action.Patch) == "null" {
		t.Error("expected non-null patch but got null")
	}
}
