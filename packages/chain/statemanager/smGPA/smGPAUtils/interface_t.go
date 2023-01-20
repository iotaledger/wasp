package smGPAUtils

import (
	"github.com/iotaledger/wasp/packages/state"
)

// May be used only for tests; deleting in production should not be available.
type TestBlockWAL interface {
	BlockWAL
	Delete(state.BlockHash) bool
}
