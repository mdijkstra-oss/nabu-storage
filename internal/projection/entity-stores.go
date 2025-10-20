package projection

import (
	"fmt"
	"hermes-relay/internal/cqrs"
	"hermes-relay/internal/domain/entities/code"
	"hermes-relay/internal/domain/entities/file"
)

var (
	// Type-specific stores
	FileStore *ProjectionStore[file.File]
	CodeStore *ProjectionStore[code.Code]

	// Type-erased registry for event application
	stores map[cqrs.AggregateType]EventApplier
)

func init() {
	// Initialize typed stores
	FileStore = NewStore(file.Reducer)
	CodeStore = NewStore(code.Reducer)

	// Register in type-erased map
	stores = map[cqrs.AggregateType]EventApplier{
		"File": FileStore,
		"Code": CodeStore,
	}
}

// StoreForNoun returns the type-erased store for event application
func StoreForNoun(noun cqrs.AggregateType) EventApplier {
	return stores[noun]
}

// Apply applies a domain event to the appropriate store
func Apply(message *cqrs.Message) error {
	if message.AggregateType == "" {
		return fmt.Errorf("event missing aggregate type %s", message.Action)
	}

	store := StoreForNoun(message.AggregateType)
	if store == nil {
		return fmt.Errorf("could not find store for noun %s", message.AggregateType)
	}

	err := store.ApplyEvent(message)

	// Todo: persist to disk? Separate func?
	return err
}
