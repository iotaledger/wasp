package statetest

import (
	"sync"

	"github.com/samber/lo"

	"github.com/iotaledger/wasp/v2/packages/kvstore"
	"github.com/iotaledger/wasp/v2/packages/state"
)

// NewStoreWithUniqueWriteMutex creates a store with a unique write mutex.
// Use only for testing -- writes will not be protected from parallel execution
func NewStoreWithUniqueWriteMutex(db kvstore.KVStore) state.Store {
	return lo.Must(state.NewStoreWithMetrics(db, true, new(sync.Mutex), nil))
}
