package events

import (
	"encoding/json"
	"hermes-relay/internal/commands"
	"os"
)

func LoadEvents(path string) ([]commands.Message, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var events []commands.Message
	if err := json.Unmarshal(data, &events); err != nil {
		return nil, err
	}

	return events, nil
}
