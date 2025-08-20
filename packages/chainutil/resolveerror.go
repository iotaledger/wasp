package chainutil

import (
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/state"
	"github.com/iotaledger/wasp/v2/packages/vm/core/errors"
)

func ResolveError(chainState state.State, e *isc.UnresolvedVMError) (*isc.VMError, error) {
	return errors.NewStateReaderFromChainState(chainState).Resolve(e)
}
