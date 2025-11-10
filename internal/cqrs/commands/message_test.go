package commands

import (
	th "hermes-relay/internal/lib/test-helpers"
	"testing"
)

type TestPayload struct {
	Name     string `json:"name" validate:"required"`
	Count    int    `json:"count" default:"10" validate:"min=1"`
	IsActive bool   `json:"is_active" default:"true"`
}

func TestEnsureValidPayload_WithDefaults(t *testing.T) {
	tests := []struct {
		Name      string
		Input     *AnyMessage
		Expected  *TestPayload
		ExpectErr string
	}{
		{
			Name: "applies defaults to zero values",
			Input: &AnyMessage{
				Payload: map[string]any{
					"name": "test",
				},
			},
			Expected: &TestPayload{
				Name:     "test",
				Count:    10,
				IsActive: true,
			},
			ExpectErr: "",
		},
		{
			Name: "does not override explicit values",
			Input: &AnyMessage{
				Payload: map[string]any{
					"name":      "test",
					"count":     5,
					"is_active": false,
				},
			},
			Expected: &TestPayload{
				Name:     "test",
				Count:    5,
				IsActive: false,
			},
			ExpectErr: "",
		},
		{
			Name: "validates after applying defaults",
			Input: &AnyMessage{
				Payload: map[string]any{
					"name":  "test",
					"count": 0,
				},
			},
			Expected:  nil,
			ExpectErr: "validation failed: Count must be at least 1 characters",
		},
		{
			Name: "fails validation when required field missing",
			Input: &AnyMessage{
				Payload: map[string]any{},
			},
			Expected:  nil,
			ExpectErr: "validation failed: Name is required",
		},
	}

	th.RunFunctionTestsWithError(t, tests, func(msg *AnyMessage) (*TestPayload, error) {
		var payload TestPayload
		if err := EnsureValidPayload(msg, &payload); err != nil {
			return nil, err
		}
		return &payload, nil
	})
}
