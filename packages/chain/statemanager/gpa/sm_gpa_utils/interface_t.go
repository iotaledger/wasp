package sm_gpa_utils

import (
	"github.com/iotaledger/wasp/packages/state"
)

// TestBlockWAL may be used only for tests; deleting in production should not be available.
type TestBlockWAL interface {
	BlockWAL
	Delete(state.BlockHash) bool
}
