package domain

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/dispatch"
	"hermes-relay/internal/domain/entities/document"
	"hermes-relay/internal/domain/entities/project"
	"hermes-relay/internal/lib/test-helpers/domain-helpers"
	rh "hermes-relay/internal/lib/test-helpers/router-helpers"
	"hermes-relay/internal/lib/utils"
	"testing"
)

var (
	healthyProjectID  = utils.NewID()
	healthyDocumentID = utils.NewID()
)

var setupCommands = []*commands.AnyMessage{
	commands.ToAny(commands.NewDomainEvent[project.CreatedProjectPayload, any](
		project.CreatedProject,
		project.CreatedProjectPayload{
			Name:        "Healthy Project",
			Description: "Test project",
		},
		project.EntityName,
		healthyProjectID,
		domain_helpers.TestActor(),
		nil,
	)),
	commands.ToAny(commands.NewDomainEvent[document.CreatedDocumentPayload, any](
		document.CreatedDocument,
		document.CreatedDocumentPayload{
			ProjectID: healthyProjectID,
			Name:      "Existing Document",
		},
		document.EntityName,
		healthyDocumentID,
		domain_helpers.TestActor(),
		nil,
	)),
}

func TestCommandRouter(t *testing.T) {
	tests := []rh.RouterTestCase{
		{
			Name: "Project creation succeeds with valid payload",
			Input: commands.ToAny(commands.NewCommand[project.CreateProjectPayload, any](
				project.CreateProject,
				project.CreateProjectPayload{Name: "New Project", Description: "Test"},
				project.EntityName,
				"",
				domain_helpers.TestActor(),
				nil,
			)),
			ExpectErr: "",
			ExpectEvent: commands.ToAny(commands.NewDomainEvent[project.CreatedProjectPayload, any](
				project.CreatedProject,
				project.CreatedProjectPayload{
					Name:        "New Project",
					Description: "Test",
				},
				project.EntityName,
				"",
				domain_helpers.TestActor(),
				nil,
			)),
		},
		{
			Name: "Project creation fails with invalid payload (missing Name)",
			Input: commands.ToAny(commands.NewCommand[project.CreateProjectPayload, any](
				project.CreateProject,
				project.CreateProjectPayload{Name: "", Description: "Test"},
				project.EntityName,
				"",
				domain_helpers.TestActor(),
				nil,
			)),
			ExpectErr: "validation failed: Name is required",
		},
		{
			Name: "Document creation succeeds on healthy project",
			Input: commands.ToAny(commands.NewCommand[document.CreateDocumentPayload, any](
				document.CreateDocument,
				document.CreateDocumentPayload{
					ProjectID: healthyProjectID,
					Name:      "New Document",
				},
				document.EntityName,
				"",
				domain_helpers.TestActor(),
				nil,
			)),
			ExpectErr: "",
			ExpectEvent: commands.ToAny(commands.NewDomainEvent[document.CreatedDocumentPayload, any](
				document.CreatedDocument,
				document.CreatedDocumentPayload{
					ProjectID: healthyProjectID,
					Name:      "New Document",
				},
				document.EntityName,
				"",
				domain_helpers.TestActor(),
				nil,
			)),
		},
		{
			Name: "Document creation fails on non-existent project",
			Input: commands.ToAny(commands.NewCommand[document.CreateDocumentPayload, any](
				document.CreateDocument,
				document.CreateDocumentPayload{
					ProjectID: utils.NewID(),
					Name:      "Orphan Document",
				},
				document.EntityName,
				"",
				domain_helpers.TestActor(),
				nil,
			)),
			ExpectErr: "validation failed: ProjectID not found",
		},
		{
			Name: "Document update succeeds on healthy document in healthy project",
			Input: commands.ToAny(commands.NewCommand[document.UpdateDocumentPayload, any](
				document.UpdateDocument,
				document.UpdateDocumentPayload{
					Name:        "Updated Document",
					Description: "Updated",
				},
				document.EntityName,
				healthyDocumentID,
				domain_helpers.TestActor(),
				nil,
			)),
			ExpectErr: "",
			ExpectEvent: commands.ToAny(commands.NewDomainEvent[document.UpdatedDocumentPayload, any](
				document.UpdatedDocument,
				document.UpdatedDocumentPayload{
					Name:        "Updated Document",
					Description: "Updated",
				},
				document.EntityName,
				healthyDocumentID,
				domain_helpers.TestActor(),
				nil,
			)),
		},
	}

	rh.RunRouterTests(t, setupCommands, tests, NewCommandRouter)
}

func TestCommandHandlers(t *testing.T) {
	type expectedMessage struct {
		Type          commands.MessageType
		Action        commands.Action
		AggregateType commands.AggregateType
	}

	tests := []struct {
		name     string
		input    *commands.AnyMessage
		expected []expectedMessage
	}{
		{
			name: "CreateProject creates project (saga disabled)",
			input: commands.ToAny(commands.NewCommand[project.CreateProjectPayload, any](
				project.CreateProject,
				project.CreateProjectPayload{Name: "New Project"},
				project.EntityName,
				"",
				domain_helpers.TestActor(),
				nil,
			)),
			expected: []expectedMessage{
				{commands.Command, project.CreateProject, project.EntityName},
				{commands.DomainEvent, project.CreatedProject, project.EntityName},
			},
		},
		{
			name: "CreateDocument does not trigger saga",
			input: commands.ToAny(commands.NewCommand[document.CreateDocumentPayload, any](
				document.CreateDocument,
				document.CreateDocumentPayload{ProjectID: healthyProjectID, Name: "Test"},
				document.EntityName,
				"",
				domain_helpers.TestActor(),
				nil,
			)),
			expected: []expectedMessage{
				{commands.Command, document.CreateDocument, document.EntityName},
				{commands.DomainEvent, document.CreatedDocument, document.EntityName},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := rh.NewTestRegistry(setupCommands)
			publisher := dispatch.NewInMemoryPublisher()

			var published []*commands.AnyMessage
			publisher.Subscribe(func(msg *commands.AnyMessage, _ dispatch.PublishFunc) (*commands.AnyMessage, error) {
				published = append(published, msg)
				if msg.Type == commands.DomainEvent {
					rh.ApplyTestEvent(reg, msg)
				}
				return nil, nil
			})

			for _, router := range CommandHandlers(reg) {
				publisher.Subscribe(router)
			}

			_, err := publisher.Publish(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(published) != len(tt.expected) {
				for i, msg := range published {
					t.Logf("published[%d]: Type=%s Action=%s AggregateType=%s", i, msg.Type, msg.Action, msg.AggregateType)
				}
				t.Fatalf("expected %d messages, got %d", len(tt.expected), len(published))
			}

			for i, exp := range tt.expected {
				got := published[i]
				if got.Type != exp.Type {
					t.Errorf("message[%d] Type: expected %s, got %s", i, exp.Type, got.Type)
				}
				if got.Action != exp.Action {
					t.Errorf("message[%d] Action: expected %s, got %s", i, exp.Action, got.Action)
				}
				if got.AggregateType != exp.AggregateType {
					t.Errorf("message[%d] AggregateType: expected %s, got %s", i, exp.AggregateType, got.AggregateType)
				}
			}
		})
	}
}
