package utils_test

import (
	th "hermes-relay/internal/lib/test-helpers"
	"hermes-relay/internal/lib/utils"
	"testing"
)

type targetStruct struct {
	ID     string
	Name   string
	Value  int
	Active bool
}

type updateStruct struct {
	ID     string
	Name   string
	Value  int
	Active bool
}

type applyPartialUpdateInput struct {
	Current     targetStruct
	Updates     updateStruct
	ExcludeKeys []string
}

func TestApplyPartialUpdate(t *testing.T) {
	tests := []struct {
		Name     string
		Input    applyPartialUpdateInput
		Expected targetStruct
	}{
		{
			Name: "applies non-zero fields and excludes ID by default",
			Input: applyPartialUpdateInput{
				Current:     targetStruct{ID: "original-id", Name: "original", Value: 10, Active: false},
				Updates:     updateStruct{ID: "new-id", Name: "updated", Value: 20},
				ExcludeKeys: nil,
			},
			Expected: targetStruct{ID: "original-id", Name: "updated", Value: 20, Active: false},
		},
		{
			Name: "excludes ID by default even when not specified",
			Input: applyPartialUpdateInput{
				Current:     targetStruct{ID: "original-id", Name: "original", Value: 10},
				Updates:     updateStruct{ID: "attempted-overwrite", Name: "updated", Value: 20},
				ExcludeKeys: nil,
			},
			Expected: targetStruct{ID: "original-id", Name: "updated", Value: 20},
		},
		{
			Name: "excludes additional keys on top of default ID",
			Input: applyPartialUpdateInput{
				Current:     targetStruct{ID: "original-id", Name: "original", Value: 10, Active: true},
				Updates:     updateStruct{ID: "new-id", Name: "updated", Value: 20, Active: false},
				ExcludeKeys: []string{"Active"},
			},
			Expected: targetStruct{ID: "original-id", Name: "updated", Value: 20, Active: true},
		},
		{
			Name: "zero values in updates are not applied",
			Input: applyPartialUpdateInput{
				Current:     targetStruct{ID: "original-id", Name: "original", Value: 10},
				Updates:     updateStruct{Name: "", Value: 0},
				ExcludeKeys: nil,
			},
			Expected: targetStruct{ID: "original-id", Name: "original", Value: 10},
		},
		{
			Name: "excluded key with zero value still excluded",
			Input: applyPartialUpdateInput{
				Current:     targetStruct{ID: "original-id", Name: "original"},
				Updates:     updateStruct{ID: "", Name: "updated"},
				ExcludeKeys: nil,
			},
			Expected: targetStruct{ID: "original-id", Name: "updated"},
		},
	}

	th.RunFunctionTests(t, tests, func(input applyPartialUpdateInput) targetStruct {
		return utils.ApplyPartialUpdate(input.Current, input.Updates, input.ExcludeKeys...)
	})
}
