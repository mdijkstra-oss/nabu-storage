package fileview

import "hermes-relay/internal/projection"

var Store = projection.NewStore(Reducer)
