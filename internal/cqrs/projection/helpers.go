package projection

import "hermes-relay/internal/lib/utils"

func EntityExists[T Entity](entities []T, id string) bool {
	return utils.Exists(entities, func(e T) bool {
		return e.GetID() == id
	})
}
