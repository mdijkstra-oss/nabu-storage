package patches

import (
	"encoding/json"
	"fmt"

	"github.com/wI2L/jsondiff"
)

func GeneratePatch[T any](before, after T) ([]byte, error) {
	patch, err := jsondiff.Compare(before, after)
	if err != nil {
		return nil, fmt.Errorf("failed to create patch: %w", err)
	}

	patchJSON, err := json.Marshal(patch)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal patch: %w", err)
	}

	return patchJSON, nil
}
