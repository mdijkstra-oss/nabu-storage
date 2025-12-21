package documentview

import (
	"hermes-relay/internal/cqrs/projection"
	"hermes-relay/internal/domain/entities/document"
)

var Reducer = projection.WithTimestamp[document.Document](
	projection.WithVersionIncrement(
		projection.WithHealthCheck(
			projection.CombineReducers(
				DocumentReducer,
				BlockReducer,
				projection.DeletedProjectReducer[document.Document],
			),
		),
	),
)
