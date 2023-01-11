package chainutil

import (
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/core/errors"
)

func ResolveError(ch chain.ChainCore, e *isc.UnresolvedVMError) (*isc.VMError, error) {
	s, err := ch.LatestState(chain.LatestState)
	if err != nil {
		return nil, err
	}
	return errors.ResolveFromState(s, e)
}
