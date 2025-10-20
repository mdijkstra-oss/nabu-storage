package file

import (
	"hermes-relay/internal/cqrs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func CreateFileEventFromPath(filePath string) (*cqrs.Message, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	filename := filepath.Base(filePath)
	fileID := strings.TrimSuffix(filename, filepath.Ext(filename))

	event := &cqrs.Message{
		Action:        "CreatedFile",
		Type:          cqrs.DomainEvent,
		AggregateID:   fileID,
		AggregateType: "File",
		Payload: File{
			ID:      fileID,
			Content: string(content),
			Attributes: Attributes{
				Codes: make(map[string][]string),
			},
		},
		Timestamp: time.Now(),
	}

	return event, nil
}
