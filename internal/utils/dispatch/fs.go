package dispatch

import (
	"encoding/json"
	"os"
)

func LoadEvents(path string) ([]Message, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var events []Message
	if err := json.Unmarshal(data, &events); err != nil {
		return nil, err
	}

	return events, nil
}
