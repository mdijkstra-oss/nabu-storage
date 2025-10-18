package events

import (
	"encoding/json"
	"fmt"
	jsonpatch "github.com/evanphx/json-patch"
	"hermes-relay/internal/commands"
)

const (
	Create string = "Created"
	Patch  string = "Patched"
	Delete string = "Deleted"
)

func ApplyEvent(state []byte, message commands.Message) ([]byte, error) {
	verb, _, err := commands.ParseAction(message.Action)
	if err != nil {
		return nil, err
	}

	switch verb {
	case Create:
		return json.Marshal(message.Payload)

	case Patch:
		patchBytes, err := json.Marshal(message.Payload)
		if err != nil {
			return nil, fmt.Errorf("marshal patch: %w", err)
		}

		patch, err := jsonpatch.DecodePatch(patchBytes)
		if err != nil {
			return nil, fmt.Errorf("decode patch: %w", err)
		}

		return patch.Apply(state)

	default:
		return nil, fmt.Errorf("unknown action: %s", message.Action)
	}
}

func ApplyEvents(state []byte, events []commands.Message) ([]byte, error) {
	var err error

	for _, event := range events {
		state, err = ApplyEvent(state, event)
		if err != nil {
			return nil, err
		}
	}

	return state, nil
}
