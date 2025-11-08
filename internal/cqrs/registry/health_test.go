package registry_test

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/domain/entities/file"
	"hermes-relay/internal/domain/entities/project"
	registry_helpers "hermes-relay/internal/lib/test-helpers/registry-helpers"
	test_helpers "hermes-relay/internal/lib/test-helpers"
	"testing"
	"time"
)

func TestProjectHealthTracking(t *testing.T) {
	t.Run("Corrupt event marks project unhealthy but other projects continue", func(t *testing.T) {
		// Setup: Create 2 projects using test helper
		setupEvents := []*commands.AnyMessage{
			commands.ToAny(commands.NewDomainEvent[project.CreatedProjectPayload, any](
				project.CreatedProject,
				project.CreatedProjectPayload{Name: "Project 1"},
				project.EntityName,
				"project-1",
				nil,
			)),
			commands.ToAny(commands.NewDomainEvent[project.CreatedProjectPayload, any](
				project.CreatedProject,
				project.CreatedProjectPayload{Name: "Project 2"},
				project.EntityName,
				"project-2",
				nil,
			)),
		}

		registry := registry_helpers.NewTestRegistry(setupEvents)

		// Verify both projects are healthy
		view1 := registry.GetProject("project-1")
		view2 := registry.GetProject("project-2")
		test_helpers.AssertEqual(t, view1.IsHealthy(), true, "project 1 should be healthy initially")
		test_helpers.AssertEqual(t, view2.IsHealthy(), true, "project 2 should be healthy initially")

		// Act: Send corrupt event to project 1 (invalid payload that can't unmarshal)
		corruptEvent := &commands.AnyMessage{
			Action:        file.CreatedFile,
			Type:          commands.DomainEvent,
			AggregateType: file.EntityName,
			AggregateID:   "file-1",
			Payload: map[string]any{
				"project_id": "project-1",
				"name":       123, // Wrong type - should be string
				"content":    []int{1, 2, 3}, // Wrong type - should be string
			},
			Timestamp: time.Now(),
		}

		registry_helpers.ApplyTestEvent(registry, corruptEvent)

		// Assert: Project 1 is now unhealthy
		test_helpers.AssertEqual(t, view1.IsHealthy(), false, "project 1 should be unhealthy after corrupt event")

		// Assert: Project 2 is still healthy
		test_helpers.AssertEqual(t, view2.IsHealthy(), true, "project 2 should still be healthy")

		// Act: Send valid events to both projects
		validEvent1 := commands.ToAny(commands.NewDomainEvent[file.CreatedFilePayload, any](
			file.CreatedFile,
			file.CreatedFilePayload{
				CreateFilePayload: file.CreateFilePayload{
					ProjectID: "project-1",
					Name:      "test.md",
					Content:   "content",
				},
				Type:   file.FileTypeSource,
				Locked: true,
			},
			file.EntityName,
			"file-2",
			nil,
		))

		validEvent2 := commands.ToAny(commands.NewDomainEvent[file.CreatedFilePayload, any](
			file.CreatedFile,
			file.CreatedFilePayload{
				CreateFilePayload: file.CreateFilePayload{
					ProjectID: "project-2",
					Name:      "test2.md",
					Content:   "content",
				},
				Type:   file.FileTypeSource,
				Locked: true,
			},
			file.EntityName,
			"file-3",
			nil,
		))

		registry_helpers.ApplyTestEvent(registry, validEvent1)
		registry_helpers.ApplyTestEvent(registry, validEvent2)

		// Assert: Project 1 did NOT process the new event (still unhealthy, no new files)
		files1 := view1.FileStore.GetAll()
		test_helpers.AssertEqual(t, len(files1), 0, "project 1 should not have processed any files")

		// Assert: Project 2 DID process the event (still healthy, has the file)
		files2 := view2.FileStore.GetAll()
		test_helpers.AssertEqual(t, len(files2), 1, "project 2 should have processed the file")
		test_helpers.AssertEqual(t, files2[0].Name, "test2.md", "project 2 should have the correct file")
	})
}
