package dispatch

import (
	"encoding/json"
	"fmt"

	jsonpatch "github.com/evanphx/json-patch"
)

const (
	Create string = "Created"
	Update string = "Updated"
)

func ApplyEvent(state []byte, event Message) ([]byte, error) {
	switch event.Action {
	case Create:
		return json.Marshal(event.Payload)

	case Update:
		patchBytes, err := json.Marshal(event.Payload)
		if err != nil {
			return nil, fmt.Errorf("marshal patch: %w", err)
		}

		patch, err := jsonpatch.DecodePatch(patchBytes)
		if err != nil {
			return nil, fmt.Errorf("decode patch: %w", err)
		}

		return patch.Apply(state)

	default:
		return nil, fmt.Errorf("unknown action: %s", event.Action)
	}
}

func ApplyEvents(state []byte, events []Message) ([]byte, error) {
	var err error

	for _, event := range events {
		state, err = ApplyEvent(state, event)
		if err != nil {
			return nil, err
		}
	}

	return state, nil
}
