package main

import (
	"fmt"
	"github.com/google/uuid"
	"hermes-relay/internal/cqrs"
	"hermes-relay/internal/domain/entities/file"
	"hermes-relay/internal/domain/entities/project"
	projectview "hermes-relay/internal/domain/projections/project-entity"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

func PublishNewSourceFiles(publish cqrs.PublishFunc) error {
	// Ensure a project exists
	projectID, err := ensureDefaultProject(publish)
	if err != nil {
		return fmt.Errorf("failed to ensure default project: %w", err)
	}

	homeDir, _ := os.UserHomeDir()
	files, err := filepath.Glob(filepath.Join(homeDir, "Documents/hermes-source-files/*.md"))
	if err != nil {
		return err
	}

	for _, path := range files {
		msg, err := NewCreatedFileAction(path, projectID)
		if err != nil {
			slog.Warn("Create failed", "path", path, "error", err)
			continue
		}

		if _, err := publish(msg); err != nil {
			slog.Warn("Publish failed", "path", path, "error", err)
		}
	}

	return nil
}

func ensureDefaultProject(publish cqrs.PublishFunc) (string, error) {
	projects := projectview.Store.GetAll()

	// If a project already exists, return its ID
	if len(projects) > 0 {
		return projects[0].ID, nil
	}

	// Create a default project
	msg := cqrs.ToAny(cqrs.NewCommand[project.CreateProjectPayload, any](
		project.CreateProject,
		project.CreateProjectPayload{
			Name: "Default Project",
		},
		project.EntityName,
		uuid.New().String(), // todo: not do it this way ofc.
		nil,
	))

	result, err := publish(msg)
	if err != nil {
		return "", fmt.Errorf("failed to create default project: %w", err)
	}

	return result.AggregateID, nil
}

func NewCreatedFileAction(filePath string, projectID string) (*cqrs.AnyMessage, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	filename := filepath.Base(filePath)

	payload := file.CreateFilePayload{
		BaseFile: file.BaseFile{
			ProjectID: projectID,
			Name:      filename,
			Attributes: file.Attributes{
				Title:   filename,
				Summary: "TBD",
			},
		},
		Content: string(content),
	}

	action := &cqrs.AnyMessage{
		Action:        file.Create,
		Type:          cqrs.Command,
		AggregateType: "File",
		Payload:       payload,
		Timestamp:     time.Now(),
		AggregateID:   uuid.New().String(),
	}

	return action, nil
}
