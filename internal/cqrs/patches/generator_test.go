package patches

import (
	"encoding/json"
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
			Expected: `{"name":"new"}`,
		},
		{
			Name: "Add field value",
			Input: patchTestInput{
				Before: testStruct{Name: "test", Count: 0},
				After:  testStruct{Name: "test", Count: 5},
			},
			Expected: `{"count":5}`,
		},
		{
			Name: "Add array field",
			Input: patchTestInput{
				Before: testStruct{Name: "test", Count: 1},
				After:  testStruct{Name: "test", Count: 1, Tags: []string{"a", "b"}},
			},
			Expected: `{"tags":["a","b"]}`,
		},
		{
			Name: "Multiple changes",
			Input: patchTestInput{
				Before: testStruct{Name: "old", Count: 1},
				After:  testStruct{Name: "new", Count: 2, Tags: []string{"x"}},
			},
			Expected: `{"count":2,"name":"new","tags":["x"]}`,
		},
		{
			Name: "No changes produces empty patch",
			Input: patchTestInput{
				Before: testStruct{Name: "same", Count: 1},
				After:  testStruct{Name: "same", Count: 1},
			},
			Expected: `{}`,
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
