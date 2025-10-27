package main

import (
	"hermes-relay/internal/cqrs"
	"hermes-relay/internal/domain/entities/file"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

func PublishNewSourceFiles(publish cqrs.PublishFunc) error {
	homeDir, _ := os.UserHomeDir()
	files, err := filepath.Glob(filepath.Join(homeDir, "Documents/hermes-source-files/*.md"))
	if err != nil {
		return err
	}

	for _, path := range files {
		msg, err := NewCreatedFileAction(path)
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

func NewCreatedFileAction(filePath string) (*cqrs.AnyMessage, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	filename := filepath.Base(filePath)

	payload := file.CreateFilePayload{
		BaseFile: file.BaseFile{
			Name: filename,
			Attributes: file.Attributes{
				Title:   "",
				Summary: "",
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
	}

	return action, nil
}
