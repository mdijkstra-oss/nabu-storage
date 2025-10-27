package project

import (
	th "hermes-relay/internal/lib/test-helpers"
	"testing"
)

func TestProjectRouter(t *testing.T) {
	tests := []th.RouterTestCase{
		th.CommandToEventCase("CreateProject with valid payload", CreateProject, CreatedProject, CreateProjectPayload{
			Name: "My Research Project",
		}, EntityName, ""),

		th.ValidationErrorCase("CreateProject with missing Name", CreateProject, CreateProjectPayload{}, EntityName, ""),
	}

	// Add common router test cases
	tests = append(tests, th.CommonRouterTestCases(CreateProject, CreateProjectPayload{
		Name: "Test Project",
	}, EntityName)...)

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			th.TestRouter(t, Router, tt)
		})
	}
}
