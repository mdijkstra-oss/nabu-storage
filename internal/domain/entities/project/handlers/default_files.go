package handlers

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/dispatch"
	"hermes-relay/internal/domain/entities/file"
	"hermes-relay/internal/domain/entities/project"
	"hermes-relay/internal/domain/templates"
	"hermes-relay/internal/lib/utils"
	"time"
)

func DefaultFilesSaga() dispatch.CommandRouter {
	return dispatch.OnEvent(project.CreatedProject, createDefaultFiles)
}

func createDefaultFiles(message *commands.AnyMessage, _ project.CreatedProjectPayload) []*commands.AnyMessage {
	return utils.Map(templates.DefaultFiles(), func(df templates.DefaultFile) *commands.AnyMessage {
		return buildCreateFileCommand(message.AggregateID, df, message.Actor)
	})
}

func buildCreateFileCommand(projectID string, df templates.DefaultFile, actor commands.Actor) *commands.AnyMessage {
	return &commands.AnyMessage{
		AggregateType: file.EntityName,
		AggregateID:   utils.NewID(),
		Action:        file.CreateFile,
		Timestamp:     time.Now(),
		Actor:         actor,
		Payload: file.CreateFilePayload{
			ProjectID: projectID,
			Name:      df.Name,
			Content:   df.Content,
			Type:      df.Type,
		},
	}
}
