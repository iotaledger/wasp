package chainutil

import (
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/viewcontext"
)

// CallView executes a view call on the latest block of the chain
func CallView(chainState state.State, ch chain.ChainCore, msg isc.Message) (isc.CallArguments, error) {
	vctx, err := viewcontext.New(ch, chainState, false)
	if err != nil {
		return nil, err
	}
	return vctx.CallViewExternal(msg)
}
