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
				Before:   func() *project.Project { p := project.BuildTestProject("p1", project.ProjectData{Name: "Old"}); return &p }(),
				After:    func() *project.Project { p := project.BuildTestProject("p1", project.ProjectData{Name: "New"}); return &p }(),
				IsActive: false,
			},
			Expected: ActionTypeNone,
		},
		{
			Name: "After is nil returns none",
			Input: decidePatchInput{
				Before:   func() *project.Project { p := project.BuildTestProject("p1", project.ProjectData{Name: "Old"}); return &p }(),
				After:    nil,
				IsActive: true,
			},
			Expected: ActionTypeNone,
		},
		{
			Name: "Before nil and after exists returns snapshot",
			Input: decidePatchInput{
				Before:   nil,
				After:    func() *project.Project { p := project.BuildTestProject("p1", project.ProjectData{Name: "New", Description: "Test"}); return &p }(),
				IsActive: true,
			},
			Expected: ActionTypeSnapshot,
		},
		{
			Name: "Before and after both exist returns patch",
			Input: decidePatchInput{
				Before:   func() *project.Project { p := project.BuildTestProject("p1", project.ProjectData{Name: "Old", Description: "Old desc"}); return &p }(),
				After:    func() *project.Project { p := project.BuildTestProject("p1", project.ProjectData{Name: "New", Description: "Old desc"}); return &p }(),
				IsActive: true,
			},
			Expected: ActionTypePatch,
		},
		{
			Name: "No changes returns none",
			Input: decidePatchInput{
				Before:   func() *project.Project { p := project.BuildTestProject("p1", project.ProjectData{Name: "Same", Description: "Same"}); return &p }(),
				After:    func() *project.Project { p := project.BuildTestProject("p1", project.ProjectData{Name: "Same", Description: "Same"}); return &p }(),
				IsActive: true,
			},
			Expected: ActionTypeNone,
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
	proj := func() *project.Project { p := project.BuildTestProject("p1", project.ProjectData{Name: "Test", Description: "Snapshot test"}); return &p }()
	action, err := DecidePatch(nil, proj, true)

	th.AssertError(t, err, "", "error")
	th.AssertEqual(t, action.Type, ActionTypeSnapshot, "action type")
	th.AssertEqual(t, action.Snapshot, proj, "snapshot content")
}

func TestDecidePatchPatchContent(t *testing.T) {
	before := func() *project.Project { p := project.BuildTestProject("p1", project.ProjectData{Name: "Old", Description: "Test"}); return &p }()
	after := func() *project.Project { p := project.BuildTestProject("p1", project.ProjectData{Name: "New", Description: "Test"}); return &p }()

	action, err := DecidePatch(before, after, true)

	th.AssertError(t, err, "", "error")
	th.AssertEqual(t, action.Type, ActionTypePatch, "action type")

	if len(action.Patch) == 0 {
		t.Error("expected patch to have content")
	}
}
