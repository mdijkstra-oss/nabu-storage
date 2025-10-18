package file

import (
	"hermes-relay/internal/commands"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func CreateFileEventFromPath(filePath string) (*commands.Message, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	filename := filepath.Base(filePath)
	fileID := strings.TrimSuffix(filename, filepath.Ext(filename))

	event := &commands.Message{
		Action:      "CreatedFile",
		AggregateID: fileID,
		Payload: File{
			ID:      fileID,
			Content: string(content),
			Attributes: Attributes{
				Codes: []Code{},
			},
		},
		Timestamp: time.Now(),
	}

	return event, nil
}
