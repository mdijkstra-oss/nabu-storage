package persistence

import (
	"encoding/json"
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/dispatch"
	"hermes-relay/internal/domain/entities/file"
	"hermes-relay/internal/domain/entities/project"
	th "hermes-relay/internal/lib/test-helpers"
	"hermes-relay/internal/lib/test-helpers/domain-helpers"
	"os"
	"path/filepath"
	"sort"
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
				domain_helpers.NewDomainEvent(project.EntityName, "project-1", project.CreatedProject, project.CreatedProjectPayload{Name: "Test"}),
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
				domain_helpers.NewDomainEvent(project.EntityName, "project-1", project.CreatedProject, project.CreatedProjectPayload{Name: "Test"}),
				domain_helpers.NewDomainEvent(project.EntityName, "project-1", "UpdatedProject", project.CreatedProjectPayload{Name: "Updated"}),
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
				domain_helpers.NewDomainEvent(project.EntityName, "project-1", project.CreatedProject, project.CreatedProjectPayload{Name: "Test 1"}),
				domain_helpers.NewDomainEvent(project.EntityName, "project-2", project.CreatedProject, project.CreatedProjectPayload{Name: "Test 2"}),
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
				domain_helpers.NewDomainEvent(project.EntityName, "project-1", project.CreatedProject, project.CreatedProjectPayload{Name: "Test"}),
				domain_helpers.NewDomainEvent(file.EntityName, "file-1", file.CreatedFile, file.CreatedFilePayload{
					CreateFilePayload: file.CreateFilePayload{ProjectID: "project-1", Name: "test.md", Content: "content"},
					Type:              file.FileTypeSource,
					Locked:            true,
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
				domain_helpers.NewDomainEvent(project.EntityName, "project-1", project.CreatedProject, project.CreatedProjectPayload{Name: "Test"}),
				domain_helpers.NewDomainEvent(file.EntityName, "file-1", file.CreatedFile, file.CreatedFilePayload{
					CreateFilePayload: file.CreateFilePayload{ProjectID: "project-1", Name: "test1.md", Content: "content1"},
					Type:              file.FileTypeSource,
					Locked:            true,
				}),
				domain_helpers.NewDomainEvent(file.EntityName, "file-2", file.CreatedFile, file.CreatedFilePayload{
					CreateFilePayload: file.CreateFilePayload{ProjectID: "project-1", Name: "test2.md", Content: "content2"},
					Type:              file.FileTypeSource,
					Locked:            true,
				}),
				domain_helpers.NewDomainEvent(file.EntityName, "file-1", "UpdatedFile", nil),
				domain_helpers.NewDomainEvent(project.EntityName, "project-2", project.CreatedProject, project.CreatedProjectPayload{Name: "Test 2"}),
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
			apply := disk.Apply()

			for _, event := range tt.Input {
				if err := apply(event); err != nil {
					th.AssertError(t, err, tt.ExpectErr, "error")
					return
				}
			}

			if tt.ExpectErr != "" {
				t.Fatalf("expected error %q but got none", tt.ExpectErr)
			}

			publisher := dispatch.NewInMemoryPublisher()
			var replayed []*commands.AnyMessage
			publisher.Subscribe(func(msg *commands.AnyMessage, _ dispatch.PublishFunc) (*commands.AnyMessage, error) {
				replayed = append(replayed, msg)
				return nil, nil
			})

			_ = disk.ReplayAllEvents(publisher)

			if len(replayed) != len(tt.Input) {
				t.Fatalf("expected %d replayed events, got %d", len(tt.Input), len(replayed))
			}

			sortEvents := func(events []*commands.AnyMessage) {
				sort.Slice(events, func(i, j int) bool {
					if events[i].AggregateType != events[j].AggregateType {
						return events[i].AggregateType < events[j].AggregateType
					}
					if events[i].AggregateID != events[j].AggregateID {
						return events[i].AggregateID < events[j].AggregateID
					}
					return events[i].Timestamp.Before(events[j].Timestamp)
				})
			}

			sortedInput := make([]*commands.AnyMessage, len(tt.Input))
			copy(sortedInput, tt.Input)
			sortEvents(sortedInput)
			sortEvents(replayed)

			for i := range sortedInput {
				compareEvents(t, sortedInput[i], replayed[i])
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

func compareEvents(t *testing.T, expected, actual *commands.AnyMessage) {
	expectedJSON, _ := json.Marshal(expected)
	actualJSON, _ := json.Marshal(actual)

	var expectedMap, actualMap map[string]any
	_ = json.Unmarshal(expectedJSON, &expectedMap)
	_ = json.Unmarshal(actualJSON, &actualMap)

	delete(expectedMap, "id")
	delete(actualMap, "id")
	delete(expectedMap, "Timestamp")
	delete(actualMap, "Timestamp")

	expectedNorm, _ := json.Marshal(expectedMap)
	actualNorm, _ := json.Marshal(actualMap)

	if string(expectedNorm) != string(actualNorm) {
		t.Errorf("events do not match:\nexpected: %s\nactual: %s", expectedNorm, actualNorm)
	}
}
