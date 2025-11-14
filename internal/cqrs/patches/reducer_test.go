package patches

import (
	"hermes-relay/internal/domain/entities/project"
	th "hermes-relay/internal/lib/test-helpers"
	"testing"
)

type decidePatchInput struct {
	Before   *project.Project
	After    *project.Project
	IsActive bool
}

func TestDecidePatch(t *testing.T) {
	tests := []struct {
		Name      string
		Input     decidePatchInput
		Expected  string
		ExpectErr string
	}{
		{
			Name: "Inactive project returns none",
			Input: decidePatchInput{
				Before:   &project.Project{ID: "p1", Name: "Old"},
				After:    &project.Project{ID: "p1", Name: "New"},
				IsActive: false,
			},
			Expected: ActionTypeNone,
		},
		{
			Name: "After is nil returns none",
			Input: decidePatchInput{
				Before:   &project.Project{ID: "p1", Name: "Old"},
				After:    nil,
				IsActive: true,
			},
			Expected: ActionTypeNone,
		},
		{
			Name: "Before nil and after exists returns snapshot",
			Input: decidePatchInput{
				Before:   nil,
				After:    &project.Project{ID: "p1", Name: "New", Description: "Test"},
				IsActive: true,
			},
			Expected: ActionTypeSnapshot,
		},
		{
			Name: "Before and after both exist returns patch",
			Input: decidePatchInput{
				Before:   &project.Project{ID: "p1", Name: "Old", Description: "Old desc"},
				After:    &project.Project{ID: "p1", Name: "New", Description: "Old desc"},
				IsActive: true,
			},
			Expected: ActionTypePatch,
		},
		{
			Name: "No changes still returns patch with empty diff",
			Input: decidePatchInput{
				Before:   &project.Project{ID: "p1", Name: "Same", Description: "Same"},
				After:    &project.Project{ID: "p1", Name: "Same", Description: "Same"},
				IsActive: true,
			},
			Expected: ActionTypePatch,
		},
		{
			Name: "Both before and after nil returns none",
			Input: decidePatchInput{
				Before:   nil,
				After:    nil,
				IsActive: true,
			},
			Expected: ActionTypeNone,
		},
	}

	th.RunFunctionTestsWithError(t, tests, func(input decidePatchInput) (PatchAction, error) {
		return DecidePatch(input.Before, input.After, input.IsActive)
	}, func(action PatchAction) string {
		return action.Type
	})
}

func TestDecidePatchSnapshotContent(t *testing.T) {
	proj := &project.Project{ID: "p1", Name: "Test", Description: "Snapshot test"}
	action, err := DecidePatch(nil, proj, true)

	th.AssertError(t, err, "", "error")
	th.AssertEqual(t, action.Type, ActionTypeSnapshot, "action type")
	th.AssertEqual(t, action.Snapshot, proj, "snapshot content")
}

func TestDecidePatchPatchContent(t *testing.T) {
	before := &project.Project{ID: "p1", Name: "Old", Description: "Test"}
	after := &project.Project{ID: "p1", Name: "New", Description: "Test"}

	action, err := DecidePatch(before, after, true)

	th.AssertError(t, err, "", "error")
	th.AssertEqual(t, action.Type, ActionTypePatch, "action type")

	if len(action.Patch) == 0 {
		t.Error("expected patch to have content")
	}
}
