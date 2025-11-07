package persistence

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/domain/entities/file"
	"hermes-relay/internal/domain/entities/project"
	th "hermes-relay/internal/lib/test-helpers"
	"os"
	"path/filepath"
	"testing"
)

type FileContent struct {
	Path     string
	Contains []string
}

func TestPersistence(t *testing.T) {
	tests := []struct {
		Name      string
		Input     []*commands.AnyMessage
		Expected  []FileContent
		ExpectErr string
	}{
		{
			Name: "Missing aggregate type returns error",
			Input: []*commands.AnyMessage{
				{Type: commands.DomainEvent, Action: "TestAction", AggregateID: "test-1"},
			},
			Expected:  nil,
			ExpectErr: "cannot persist event without aggregate type",
		},
		{
			Name: "Missing aggregate ID returns error",
			Input: []*commands.AnyMessage{
				{Type: commands.DomainEvent, Action: "TestAction", AggregateType: "Test"},
			},
			Expected:  nil,
			ExpectErr: "cannot persist event without aggregate ID",
		},
		{
			Name: "Single event persists and loads",
			Input: []*commands.AnyMessage{
				th.NewDomainEvent(project.EntityName, "project-1", project.CreatedProject, project.CreatedProjectPayload{Name: "Test"}),
			},
			Expected: []FileContent{
				{
					Path:     "Project/project-1.jsonl",
					Contains: []string{`"action":"CreatedProject"`, `"aggregateId":"project-1"`, `"payload":{"name":"Test"}`, `"version":1`},
				},
			},
			ExpectErr: "",
		},
		{
			Name: "Multiple events for same aggregate append to same file",
			Input: []*commands.AnyMessage{
				th.NewDomainEvent(project.EntityName, "project-1", project.CreatedProject, project.CreatedProjectPayload{Name: "Test"}),
				th.NewDomainEvent(project.EntityName, "project-1", "UpdatedProject", project.CreatedProjectPayload{Name: "Updated"}),
			},
			Expected: []FileContent{
				{
					Path:     "Project/project-1.jsonl",
					Contains: []string{`"action":"CreatedProject"`, `"action":"UpdatedProject"`, `"aggregateId":"project-1"`, `"version":1`},
				},
			},
			ExpectErr: "",
		},
		{
			Name: "Multiple aggregates create separate files",
			Input: []*commands.AnyMessage{
				th.NewDomainEvent(project.EntityName, "project-1", project.CreatedProject, project.CreatedProjectPayload{Name: "Test 1"}),
				th.NewDomainEvent(project.EntityName, "project-2", project.CreatedProject, project.CreatedProjectPayload{Name: "Test 2"}),
			},
			Expected: []FileContent{
				{Path: "Project/project-1.jsonl", Contains: []string{`"aggregateId":"project-1"`, `"name":"Test 1"`, `"version":1`}},
				{Path: "Project/project-2.jsonl", Contains: []string{`"aggregateId":"project-2"`, `"name":"Test 2"`, `"version":1`}},
			},
			ExpectErr: "",
		},
		{
			Name: "Multiple aggregate types create separate directories",
			Input: []*commands.AnyMessage{
				th.NewDomainEvent(project.EntityName, "project-1", project.CreatedProject, project.CreatedProjectPayload{Name: "Test"}),
				th.NewDomainEvent(file.EntityName, "file-1", file.CreatedFile, file.CreatedFilePayload{
					BaseFile: file.BaseFile{ProjectID: "project-1", Name: "test.md"},
					Content:  "content",
				}),
			},
			Expected: []FileContent{
				{Path: "File/file-1.jsonl", Contains: []string{`"action":"CreatedFile"`, `"aggregateId":"file-1"`, `"aggregateType":"File"`, `"version":1`}},
				{Path: "Project/project-1.jsonl", Contains: []string{`"aggregateId":"project-1"`, `"aggregateType":"Project"`, `"name":"Test"`, `"version":1`}},
			},
			ExpectErr: "",
		},
		{
			Name: "Many events persist and load correctly",
			Input: []*commands.AnyMessage{
				th.NewDomainEvent(project.EntityName, "project-1", project.CreatedProject, project.CreatedProjectPayload{Name: "Test"}),
				th.NewDomainEvent(file.EntityName, "file-1", file.CreatedFile, file.CreatedFilePayload{
					BaseFile: file.BaseFile{ProjectID: "project-1", Name: "test1.md"},
					Content:  "content1",
				}),
				th.NewDomainEvent(file.EntityName, "file-2", file.CreatedFile, file.CreatedFilePayload{
					BaseFile: file.BaseFile{ProjectID: "project-1", Name: "test2.md"},
					Content:  "content2",
				}),
				th.NewDomainEvent(file.EntityName, "file-1", "UpdatedFile", nil),
				th.NewDomainEvent(project.EntityName, "project-2", project.CreatedProject, project.CreatedProjectPayload{Name: "Test 2"}),
			},
			Expected: []FileContent{
				{Path: "File/file-1.jsonl", Contains: []string{`"action":"CreatedFile"`, `"action":"UpdatedFile"`, `"aggregateId":"file-1"`, `"version":1`}},
				{Path: "File/file-2.jsonl", Contains: []string{`"action":"CreatedFile"`, `"aggregateId":"file-2"`, `"version":1`}},
				{Path: "Project/project-1.jsonl", Contains: []string{`"aggregateId":"project-1"`, `"name":"Test"`, `"version":1`}},
				{Path: "Project/project-2.jsonl", Contains: []string{`"aggregateId":"project-2"`, `"name":"Test 2"`, `"version":1`}},
			},
			ExpectErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			tmpDir := t.TempDir()
			disk := NewDiskPersistence(tmpDir)

			for _, event := range tt.Input {
				if err := disk.PersistEvent(event); err != nil {
					th.AssertError(t, err, tt.ExpectErr, "error")
					return
				}
			}

			if tt.ExpectErr != "" {
				t.Fatalf("expected error %q but got none", tt.ExpectErr)
			}

			for _, expectedFile := range tt.Expected {
				filePath := filepath.Join(tmpDir, expectedFile.Path)
				content, _ := os.ReadFile(filePath)

				for _, expectedContent := range expectedFile.Contains {
					th.AssertContains(t, string(content), expectedContent)
				}
			}
		})
	}
}
