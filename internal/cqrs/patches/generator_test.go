package patches

import (
	"encoding/json"
	jsonpatch "github.com/evanphx/json-patch/v5"
	"hermes-relay/internal/domain/entities/project"
	th "hermes-relay/internal/lib/test-helpers"
	"hermes-relay/internal/lib/utils"
	"testing"
)

type testStruct struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
	Tags  []string `json:"tags,omitempty"`
}

type patchTestInput struct {
	Before testStruct
	After  testStruct
}

func TestGeneratePatch(t *testing.T) {
	tests := []struct {
		Name      string
		Input     patchTestInput
		Expected  string
		ExpectErr string
	}{
		{
			Name: "Replace field value",
			Input: patchTestInput{
				Before: testStruct{Name: "old", Count: 1},
				After:  testStruct{Name: "new", Count: 1},
			},
			Expected: `[{"op":"replace","path":"/name","value":"new"}]`,
		},
		{
			Name: "Add field value",
			Input: patchTestInput{
				Before: testStruct{Name: "test", Count: 0},
				After:  testStruct{Name: "test", Count: 5},
			},
			Expected: `[{"op":"replace","path":"/count","value":5}]`,
		},
		{
			Name: "Add array field",
			Input: patchTestInput{
				Before: testStruct{Name: "test", Count: 1},
				After:  testStruct{Name: "test", Count: 1, Tags: []string{"a", "b"}},
			},
			Expected: `[{"op":"add","path":"/tags","value":["a","b"]}]`,
		},
		{
			Name: "Multiple changes",
			Input: patchTestInput{
				Before: testStruct{Name: "old", Count: 1},
				After:  testStruct{Name: "new", Count: 2, Tags: []string{"x"}},
			},
			Expected: `[{"op":"replace","path":"/count","value":2},{"op":"replace","path":"/name","value":"new"},{"op":"add","path":"/tags","value":["x"]}]`,
		},
		{
			Name: "No changes produces empty patch",
			Input: patchTestInput{
				Before: testStruct{Name: "same", Count: 1},
				After:  testStruct{Name: "same", Count: 1},
			},
			Expected: `null`,
		},
	}

	th.RunFunctionTestsWithError(t, tests, func(input patchTestInput) ([]byte, error) {
		return GeneratePatch(input.Before, input.After)
	}, func(patch []byte) string {
		var normalized any
		utils.ShouldWork(json.Unmarshal(patch, &normalized))
		result := utils.Should(json.Marshal(normalized))
		return string(result)
	})
}

func TestGeneratePatchRoundTrip(t *testing.T) {
	tests := []struct {
		Name   string
		Before *project.Project
		After  *project.Project
	}{
		{
			Name: "Simple field update",
			Before: &project.Project{
				ID:      "proj-1",
				Healthy: true,
				ProjectData: project.ProjectData{
					Name:        "Old Name",
					Description: "Old Description",
				},
			},
			After: &project.Project{
				ID:      "proj-1",
				Healthy: true,
				ProjectData: project.ProjectData{
					Name:        "New Name",
					Description: "Old Description",
				},
			},
		},
		{
			Name: "Multiple field changes",
			Before: &project.Project{
				ID:      "proj-1",
				Healthy: true,
				ProjectData: project.ProjectData{
					Name:        "Project A",
					Description: "First version",
				},
			},
			After: &project.Project{
				ID:      "proj-1",
				Healthy: true,
				ProjectData: project.ProjectData{
					Name:        "Project B",
					Description: "Second version",
				},
			},
		},
		{
			Name: "Health status change",
			Before: &project.Project{
				ID:      "proj-1",
				Healthy: true,
				ProjectData: project.ProjectData{
					Name:        "Test",
					Description: "Test",
				},
			},
			After: &project.Project{
				ID:      "proj-1",
				Healthy: false,
				ProjectData: project.ProjectData{
					Name:        "Test",
					Description: "Test",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			patch, err := GeneratePatch(tt.Before, tt.After)
			th.AssertError(t, err, "", "generate patch")

			beforeJSON := utils.Should(json.Marshal(tt.Before))
			patchObj, err := jsonpatch.DecodePatch(patch)
			th.AssertError(t, err, "", "decode patch")

			result, err := patchObj.Apply(beforeJSON)
			th.AssertError(t, err, "", "apply patch")

			var patched project.Project
			utils.ShouldWork(json.Unmarshal(result, &patched))

			th.AssertEqual(t, &patched, tt.After, "patched should equal after")
		})
	}
}
