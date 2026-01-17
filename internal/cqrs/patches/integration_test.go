package patches

import (
	"hermes-relay/internal/domain/entities/document"
	"hermes-relay/internal/domain/entities/project"
	th "hermes-relay/internal/lib/test-helpers"
	"testing"
)

func buildProject(documents map[string]document.Document) *project.Project {
	proj := project.BuildTestProject("proj-1", project.ProjectData{})
	proj.Documents = documents
	return &proj
}

func buildDocument(id, name string, content []document.Block, annotations map[string]document.Annotation) document.Document {
	tree := document.FromArray(content)
	d := document.BuildTestDocument(id, document.DocumentData{ProjectID: "proj-1", Name: name})
	d.Blocks = tree.Blocks
	d.HeadID = tree.HeadID
	d.TailID = tree.TailID
	d.Annotations = annotations
	return d
}

func annMap(anns ...document.Annotation) map[string]document.Annotation {
	result := make(map[string]document.Annotation, len(anns))
	for _, a := range anns {
		result[a.ID] = a
	}
	return result
}

func buildTestAnnotation(id, text string) document.Annotation {
	return document.Annotation{
		ID:    id,
		Text:  text,
		Actor: "test",
		Color: "blue",
	}
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
			Name: "Removing annotations generates patch",
			Before: buildProject(map[string]document.Document{
				"doc-1": buildDocument("doc-1", "test.txt", nil, annMap(
					buildTestAnnotation("a1", "text1"),
					buildTestAnnotation("a2", "text2"),
					buildTestAnnotation("a3", "text3"),
				)),
			}),
			After: buildProject(map[string]document.Document{
				"doc-1": buildDocument("doc-1", "test.txt", nil, annMap(
					buildTestAnnotation("a3", "text3"),
				)),
			}),
			IsActive:        true,
			ExpectedType:    ActionTypePatch,
			ExpectNullPatch: false,
		},
		{
			Name: "Adding annotations generates patch",
			Before: buildProject(map[string]document.Document{
				"doc-1": buildDocument("doc-1", "test.txt", nil, map[string]document.Annotation{}),
			}),
			After: buildProject(map[string]document.Document{
				"doc-1": buildDocument("doc-1", "test.txt", nil, annMap(
					buildTestAnnotation("a1", "new"),
				)),
			}),
			IsActive:        true,
			ExpectedType:    ActionTypePatch,
			ExpectNullPatch: false,
		},
		{
			Name: "Changing document name generates patch",
			Before: buildProject(map[string]document.Document{
				"doc-1": buildDocument("doc-1", "old.txt", nil, nil),
			}),
			After: buildProject(map[string]document.Document{
				"doc-1": buildDocument("doc-1", "new.txt", nil, nil),
			}),
			IsActive:        true,
			ExpectedType:    ActionTypePatch,
			ExpectNullPatch: false,
		},
		{
			Name: "Changing document content generates patch",
			Before: buildProject(map[string]document.Document{
				"doc-1": buildDocument("doc-1", "test.txt", []document.Block{{ID: "b1", Type: document.BlockTypeParagraph}}, nil),
			}),
			After: buildProject(map[string]document.Document{
				"doc-1": buildDocument("doc-1", "test.txt", []document.Block{{ID: "b2", Type: document.BlockTypeHeading}}, nil),
			}),
			IsActive:        true,
			ExpectedType:    ActionTypePatch,
			ExpectNullPatch: false,
		},
		{
			Name:            "Identical projects return none",
			Before:          buildProject(map[string]document.Document{}),
			After:           buildProject(map[string]document.Document{}),
			IsActive:        true,
			ExpectedType:    ActionTypeNone,
			ExpectNullPatch: true,
		},
		{
			Name:            "Nil before returns snapshot",
			Before:          nil,
			After:           buildProject(map[string]document.Document{}),
			IsActive:        true,
			ExpectedType:    ActionTypeSnapshot,
			ExpectNullPatch: true,
		},
		{
			Name:            "Nil after returns none",
			Before:          buildProject(map[string]document.Document{}),
			After:           nil,
			IsActive:        true,
			ExpectedType:    ActionTypeNone,
			ExpectNullPatch: true,
		},
		{
			Name:            "Inactive project returns none",
			Before:          buildProject(map[string]document.Document{}),
			After:           buildProject(map[string]document.Document{"doc-1": buildDocument("doc-1", "new.txt", nil, nil)}),
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
