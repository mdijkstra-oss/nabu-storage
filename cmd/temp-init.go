package main

import (
	"context"
	"hermes-relay/internal/cqrs"
	"hermes-relay/internal/domain/entities/file"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func PublishNewSourceFiles(publish cqrs.PublishFunc, store interface {
	GetByID(string) (*file.File, error)
}) error {
	homeDir, _ := os.UserHomeDir()
	files, err := filepath.Glob(filepath.Join(homeDir, "Documents/hermes-source-files/*.md"))
	if err != nil {
		return err
	}

	for _, path := range files {
		fileID := strings.TrimSuffix(filepath.Base(path), ".md")

		if _, err := store.GetByID(fileID); err == nil {
			continue
		}

		msg, err := NewCreatedFileAction(path)
		if err != nil {
			slog.Warn("Create failed", "path", path, "error", err)
			continue
		}

		if _, err := publish(context.Background(), msg); err != nil {
			slog.Warn("Publish failed", "path", path, "error", err)
		}
	}

	return nil
}

func NewCreatedFileAction(filePath string) (*cqrs.AnyMessage, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	filename := filepath.Base(filePath)
	fileID := strings.TrimSuffix(filename, filepath.Ext(filename))

	event := &cqrs.AnyMessage{
		Action:        file.Create,
		Type:          cqrs.Command,
		AggregateID:   fileID,
		AggregateType: "File",
		Payload: file.CreatedFilePayload{
			ID:      fileID,
			Content: string(content),
		},
		Timestamp: time.Now(),
	}

	return event, nil
}
