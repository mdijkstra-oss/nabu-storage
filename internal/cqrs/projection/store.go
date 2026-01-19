package projection

// Todo: I think this locks at whole system level per message
// Probably fine, but can probably quite easily do it on a per project level

import (
	"hermes-relay/internal/cqrs/commands"
	"sync"
)

var replayMode bool

func SetReplayMode(v bool) { replayMode = v }
func IsReplayMode() bool   { return replayMode }

type Store[S any] struct {
	mu      sync.RWMutex
	state   *S
	reducer Reducer[S]
}

func NewStore[S any](initial *S, reducer Reducer[S]) *Store[S] {
	return &Store[S]{state: initial, reducer: reducer}
}

func Apply[S any](store *Store[S], event *commands.AnyMessage) {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.state = store.reducer(store.state, event)
}

func Read[S, R any](store *Store[S], selector func(*S) R) R {
	store.mu.RLock()
	defer store.mu.RUnlock()
	return selector(store.state)
}
