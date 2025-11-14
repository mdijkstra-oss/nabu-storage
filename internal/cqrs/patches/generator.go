package patches

import (
	"encoding/json"
	"fmt"

	jsonpatch "github.com/evanphx/json-patch/v5"
)

func GeneratePatch[T any](before, after T) ([]byte, error) {
	beforeJSON, err := json.Marshal(before)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal before state: %w", err)
	}

	afterJSON, err := json.Marshal(after)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal after state: %w", err)
	}

	patch, err := jsonpatch.CreateMergePatch(beforeJSON, afterJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to create patch: %w", err)
	}

	return patch, nil
}
