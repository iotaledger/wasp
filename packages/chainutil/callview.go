// Package chainutil provides utility functions for blockchain operations and interactions.
package chainutil

import (
	"github.com/iotaledger/hive.go/log"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/processors"
	"github.com/iotaledger/wasp/packages/vm/viewcontext"
)

// CallView executes a view call on the latest block of the chain
func CallView(
	chainID isc.ChainID,
	chainState state.State,
	processors *processors.Config,
	log log.Logger,
	msg isc.Message,
) (isc.CallArguments, error) {
	vctx, err := viewcontext.New(chainID, chainState, processors, log, false)
	if err != nil {
		return nil, err
	}
	return vctx.CallViewExternal(msg)
}
